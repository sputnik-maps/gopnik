package tileserver

import (
	"container/list"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"app"
	"gopnik"
	"tilerender"

	"github.com/op/go-logging"
	"github.com/orofarne/hmetrics2"
)

var log = logging.MustGetLogger("global")

var hReqT = hmetrics2.MustRegisterPackageMetric("request_time", hmetrics2.NewHistogram()).(*hmetrics2.Histogram)
var hReq200 = hmetrics2.MustRegisterPackageMetric("code_200", hmetrics2.NewCounter()).(*hmetrics2.Counter)
var hReq400 = hmetrics2.MustRegisterPackageMetric("code_400", hmetrics2.NewCounter()).(*hmetrics2.Counter)
var hReq500 = hmetrics2.MustRegisterPackageMetric("code_500", hmetrics2.NewCounter()).(*hmetrics2.Counter)

type TileServer struct {
	renders     []zoomRender
	cache       gopnik.CachePluginInterface
	saveList    *list.List
	saveListMu  sync.RWMutex
	removeDelay time.Duration
}

type zoomRender struct {
	Render  *tilerender.RenderPool
	MinZoom uint
	MaxZoom uint
	Tags    []string
}

type saveQueueElem struct {
	gopnik.TileCoord
	Data []gopnik.Tile
}

var pathRegex = regexp.MustCompile(`/([0-9]+)/([0-9]+)/([0-9]+)\.png`)

func NewTileServer(poolsCfg app.RenderPoolsConfig, cp gopnik.CachePluginInterface, removeDelay time.Duration) (*TileServer, error) {
	srv := &TileServer{
		renders:     make([]zoomRender, len(poolsCfg.RenderPools)),
		cache:       cp,
		saveList:    list.New(),
		removeDelay: removeDelay,
	}

	for i := 0; i < len(srv.renders); i++ {
		var err error
		srv.renders[i].MinZoom = poolsCfg.RenderPools[i].MinZoom
		srv.renders[i].MaxZoom = poolsCfg.RenderPools[i].MaxZoom
		srv.renders[i].Tags = poolsCfg.RenderPools[i].Tags
		srv.renders[i].Render, err = tilerender.NewRenderPool(
			poolsCfg.RenderPools[i].Cmd, poolsCfg.RenderPools[i].PoolSize,
			poolsCfg.RenderPools[i].QueueSize, poolsCfg.RenderPools[i].RenderTTL)
		if err != nil {
			return nil, err
		}
	}

	return srv, nil
}

func (srv *TileServer) cacheMetatile(tc gopnik.TileCoord, tiles []gopnik.Tile) {
	listElem := srv.saveQueuePut(tc, tiles)
	defer func() {
		time.Sleep(srv.removeDelay)
		srv.saveQueueRemove(listElem)
	}()

	err := srv.cache.Set(tc, tiles)
	if err != nil {
		log.Error("Cache write error: %v", err)
	}
}

func (srv *TileServer) saveQueuePut(coord gopnik.TileCoord, tiles []gopnik.Tile) *list.Element {
	srv.saveListMu.Lock()
	defer srv.saveListMu.Unlock()

	elem := saveQueueElem{
		TileCoord: coord,
		Data:      tiles,
	}
	return srv.saveList.PushFront(&elem)
}

func (srv *TileServer) saveQueueRemove(elem *list.Element) {
	srv.saveListMu.Lock()
	defer srv.saveListMu.Unlock()

	srv.saveList.Remove(elem)
}

func (srv *TileServer) saveQueueGet(coord gopnik.TileCoord) []gopnik.Tile {
	srv.saveListMu.RLock()
	defer srv.saveListMu.RUnlock()

	for e := srv.saveList.Front(); e != nil; e = e.Next() {
		elem := e.Value.(*saveQueueElem)
		if elem.Equals(&coord) {
			return elem.Data
		}
	}
	return nil
}

func (srv *TileServer) checkSaveQueue(coord gopnik.TileCoord) []byte {
	metacoord := app.App.Metatiler().TileToMetatile(&coord)

	data := srv.saveQueueGet(metacoord)
	if data == nil {
		return nil
	}

	index := (coord.Y-metacoord.Y)*metacoord.Size + (coord.X - metacoord.X)
	return data[index].Image
}

func (srv *TileServer) serveTileRequest(render *tilerender.RenderPool, w http.ResponseWriter, r *http.Request, tc gopnik.TileCoord) {
	if data := srv.checkSaveQueue(tc); data != nil {
		w.Header().Set("Content-Type", "image/png")
		_, err := w.Write(data)
		if err != nil {
			log.Warning("HTTP Write error: %v", err)
		}
		return
	}

	metacoord := app.App.Metatiler().TileToMetatile(&tc)

	ansCh := make(chan *tilerender.RenderPoolResponse)

	err := render.EnqueueRequest(metacoord, ansCh)
	if err != nil {
		log.Error("EnqueueRequest error: %v", err)
		http.Error(w, err.Error(), 500)
		hReq500.Inc()
		return
	}

	ans := <-ansCh
	if ans.Error != nil {
		log.Error("Render error: %v", ans.Error)
		http.Error(w, ans.Error.Error(), 500)
		hReq500.Inc()
		return
	}

	go srv.cacheMetatile(metacoord, ans.Tiles)
	index := (tc.Y-metacoord.Y)*metacoord.Size + (tc.X - metacoord.X)

	w.Header().Set("Content-Type", "image/png")
	_, err = w.Write(ans.Tiles[index].Image)
	if err != nil {
		log.Warning("HTTP Write error: %v", err)
	}
}

func (srv *TileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	τ0 := time.Now()
	defer hReqT.AddPoint(time.Since(τ0).Seconds())

	var err error
	log.Info("New request: %v", r.URL.String())

	if strings.HasSuffix(r.URL.String(), "/status") {
		w.Write([]byte{'O', 'k'})
		return
	}

	path := pathRegex.FindStringSubmatch(r.URL.Path)
	tags := r.URL.Query()["tag"]

	if path == nil {
		log.Debug("nil path: %v", r.Header)
		http.Error(w, "nil path", 400)
		hReq400.Inc()
		return
	}

	z, _ := strconv.ParseUint(path[1], 10, 64)
	x, _ := strconv.ParseUint(path[2], 10, 64)
	y, _ := strconv.ParseUint(path[3], 10, 64)

	size := uint64(1)
	if sizeQuery, found := r.URL.Query()["size"]; found && len(sizeQuery) > 0 {
		size, err = strconv.ParseUint(sizeQuery[0], 10, 0)
		if err != nil {
			log.Debug("Invalid size: %v", err)
			http.Error(w, err.Error(), 400)
			hReq400.Inc()
			return
		}
	}

	coord := gopnik.TileCoord{
		X:    x,
		Y:    y,
		Zoom: z,
		Size: size,
		Tags: tags,
	}

RL:
	for _, renderCfg := range srv.renders {
		if renderCfg.Tags != nil {
		TL:
			for _, cfgT := range renderCfg.Tags {
				for _, inT := range tags {
					if inT == cfgT {
						continue TL
					}
				}
				continue RL
			}
		}
		if z < uint64(renderCfg.MinZoom) || z > uint64(renderCfg.MaxZoom) {
			continue
		}
		srv.serveTileRequest(renderCfg.Render, w, r, coord)

		hReq200.Inc()
		return
	}

	log.Debug("Invalid zoom %v", z)
	http.Error(w, "Invalid zoom", 400)
	hReq400.Inc()
}

func (srv *TileServer) ReloadStyle() error {
	for _, r := range srv.renders {
		r.Render.Reload()
	}
	return nil
}
