package tilerouter

/* FIXME

import (
	"defplugins/testcache"
	json "github.com/orofarne/strict-json"
	"imgsplitter"
	"interfaces"
	"net/http"
	"plugin"
	"runtime"
	"sampledata"
	"testing"
	"tilerender"
	"tileserver"
	"time"
)

type Config struct {
	Stylesheet       string                     // Mapnik stylesheet
	PoolSize         int                        // Render pool size
	QueueSize        int                        // Render queue size
	Threads          int                        // Use >= n threads
	FontPath         string                     // Mapnik font path
	Addr             string                     // Bind addr
	DebugAddr        string                     // Address for statistics
	HTTPReadTimeout  string                     //
	HTTPWriteTimeout string                     //
	CachePlugin      string                     //
	MetaSize         uint64                     //
	Plugins          map[string]json.RawMessage //
	Logging          json.RawMessage            // see loghelper.go
}

func TestOneRequest(t *testing.T) {
	addr := "127.0.0.1:5342"

	cfg := Config{
		Stylesheet:       "sampledata/stylesheet.xml",
		PoolSize:         runtime.NumCPU(),
		QueueSize:        100,
		Threads:          1,
		FontPath:         "",
		Addr:             ":8090",
		DebugAddr:        ":9090",
		HTTPReadTimeout:  "60s",
		HTTPWriteTimeout: "60s",
		CachePlugin:      "TestCachePlugin",
		Plugins:          map[string]json.RawMessage{"TestCachePlugin": {}},
		MetaSize:         8,
	}

	var cp interfaces.CachePluginInterface
	var ok bool
	imgsplitter.SetMetaSize(2)
	plugin.DefaultPluginStore.AddPlugin(new(testcache.TestCachePluginFactory))

	if plug, err := plugin.DefaultPluginStore.New(cfg.CachePlugin, cfg.Plugins[cfg.CachePlugin]); err == nil {
		if cp, ok = plug.(interfaces.CachePluginInterface); cp == nil || !ok {
			log.Fatalf("Invalid plugin '%s', '%v', '%v'", cfg.CachePlugin, cp, ok)
		}
	} else {
		log.Fatalf("Failed to configure plugin '%s': %v", cfg.CachePlugin, err)
	}

	ts, err := tileserver.NewTileServer("gopnikslave", sampledata.Stylesheet, 1, 1, cp, 0)
	if err != nil {
		t.Errorf("NewTilesServer error: %v", err)
	}
	s := &http.Server{
		Addr:           addr,
		Handler:        ts,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go s.ListenAndServe()

	router, err := NewTileRouter([]string{addr}, 1*time.Second, 30*time.Second)
	if err != nil {
		t.Errorf("tileserver.NewTileRouter error: %v", err)
	}
	coord := tilerender.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	img, err := router.Tile(coord)
	if err != nil {
		t.Errorf("router.Tile error: %v", err)
	}

	sampledata.CheckTile(t, img, sampledata.TileImage)
}
*/
