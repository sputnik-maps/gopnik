package tilerouter

import (
	"container/list"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopnik"

	"github.com/op/go-logging"
	"github.com/orofarne/hmetrics2"
)

var log = logging.MustGetLogger("global")

var hReqT = hmetrics2.MustRegisterPackageMetric("request_time", hmetrics2.NewHistogram()).(*hmetrics2.Histogram)
var hCacheT = hmetrics2.MustRegisterPackageMetric("cache_time", hmetrics2.NewHistogram()).(*hmetrics2.Histogram)
var hRenderT = hmetrics2.MustRegisterPackageMetric("render_time", hmetrics2.NewHistogram()).(*hmetrics2.Histogram)
var hReq200 = hmetrics2.MustRegisterPackageMetric("code_200", hmetrics2.NewCounter()).(*hmetrics2.Counter)
var hReq400 = hmetrics2.MustRegisterPackageMetric("code_400", hmetrics2.NewCounter()).(*hmetrics2.Counter)
var hReq500 = hmetrics2.MustRegisterPackageMetric("code_500", hmetrics2.NewCounter()).(*hmetrics2.Counter)

type renderResponse struct {
	Error error
	Tile  []byte
}

type renderTask struct {
	gopnik.TileCoord
	ResultCh []chan renderResponse
}

type RouterServer struct {
	router  *TileRouter
	cluster gopnik.ClusterPluginInterface
	cache   gopnik.CachePluginInterface
	filter  gopnik.FilterPluginInterface
	tasks   *list.List
	tasksMu sync.Locker
}

func NewRouterServer(cl gopnik.ClusterPluginInterface,
	cp gopnik.CachePluginInterface, filter gopnik.FilterPluginInterface,
	renderTimeout time.Duration, pingPeriod time.Duration) (*RouterServer, error) {
	var err error

	srv := new(RouterServer)
	srv.cluster = cl
	srv.cache = cp
	srv.filter = filter
	srv.tasks = list.New()
	srv.tasksMu = new(sync.Mutex)

	srv.router, err = NewTileRouter(nil, renderTimeout, pingPeriod)
	if err != nil {
		return nil, err
	}

	err = srv.updateConfig()
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (srv *RouterServer) updateConfig() error {
	renders, err := srv.cluster.GetRenders()
	if err != nil {
		return err
	}
	srv.router.UpdateRenders(renders)
	return nil
}

func (srv *RouterServer) tryRenderTile(coord gopnik.TileCoord) ([]byte, error) {
	var resp renderResponse

	log.Info("Request to render tile %v", coord)

	resp.Tile, resp.Error = srv.router.Tile(coord)

	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Tile, nil
}

func (srv *RouterServer) tryCache(coord gopnik.TileCoord) ([]byte, error) {
	τ0 := time.Now()
	defer hCacheT.AddPoint(time.Since(τ0).Seconds())

	return srv.cache.Get(coord)
}

func (srv *RouterServer) serveTileRequest(w http.ResponseWriter, r *http.Request, coord gopnik.TileCoord) bool {
	// Apply filter
	coord, err := srv.filter.Filter(coord)
	if err != nil {
		log.Debug("Filter error: %v", err, r.URL.Path)
		http.Error(w, fmt.Sprintf("Filter error: %v", err), 400)
		hReq400.Inc()
		return false
	}
	log.Debug("Filtred coord: %v", coord)

	// Try cache
	tile, err := srv.tryCache(coord)
	if err != nil {
		log.Error("Cache read error: %v", err)
	}

	if err != nil || tile == nil || len(tile) == 0 {
		// Try render
		tile, err = srv.tryRenderTile(coord)
		if err != nil {
			log.Error("Metatile render error: %v", err)
			http.Error(w, err.Error(), 500)
			hReq500.Inc()
			return false
		}
	}

	w.Header().Set("Content-Type", "image/png")
	_, err = w.Write(tile)
	if err != nil {
		log.Warning("HTTP Write error: %v", err)
	}

	return true
}

func (srv *RouterServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	τ0 := time.Now()
	defer hReqT.AddPoint(time.Since(τ0).Seconds())

	log.Info("New request: %v", r.URL.String())

	// URL format: blah-blah-blah/$z/$x/$y.png

	if !strings.HasSuffix(r.URL.Path, ".png") {
		log.Debug("Invalid tile extension", r.URL.Path)
		http.Error(w, "Invalid tile extension", 400)
		hReq400.Inc()
		return
	}

	pathParts := strings.Split(r.URL.Path[0:len(r.URL.Path)-4], "/")

	if len(pathParts) < 3 {
		log.Debug("invalid path: %v", r.Header)
		http.Error(w, "invalid path", 400)
		hReq400.Inc()
		return
	}

	z, _ := strconv.ParseUint(pathParts[len(pathParts)-3], 10, 64)
	x, _ := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 64)
	y, _ := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 64)
	tags := r.URL.Query()["tag"]

	if srv.serveTileRequest(w, r, gopnik.TileCoord{
		X:    x,
		Y:    y,
		Zoom: z,
		Size: 1,
		Tags: tags,
	}) {
		hReq200.Inc()
	}
}
