package main

import (
	"net/http"
	"os"
	"os/signal"
	"plugins"
	_ "plugins_enabled"
	"servicestatus"
	"syscall"
	"tileserver"
	"time"

	"app"
	"gopnik"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

type RenderConfig struct {
	Threads          int             // Use >= n threads
	Addr             string          // Bind addr
	DebugAddr        string          // Address for statistics
	HTTPReadTimeout  string          //
	HTTPWriteTimeout string          //
	HotCacheDelay    string          // Time period after cache set is done and before drop hot cache
	Logging          json.RawMessage // see loghelper.go
}

type Config struct {
	Render                RenderConfig     // Render config
	CachePlugin           app.PluginConfig //
	app.CommonConfig                       //
	app.RenderPoolsConfig                  //
	json.OtherKeys                         //
}

func sigHandler(sig os.Signal, actionFunc func() error, errMsg string) {
	ch := make(chan os.Signal)
	signal.Notify(ch, sig)

	for s := range ch {
		if s == sig {
			err := actionFunc()
			if err != nil {
				log.Error("%s: %v", errMsg, err)
			}
		}
	}
}

func main() {
	cfg := Config{
		Render: RenderConfig{
			Threads:          1,
			Addr:             ":8090",
			DebugAddr:        ":9090",
			HTTPReadTimeout:  "60s",
			HTTPWriteTimeout: "60s",
			HotCacheDelay:    "0s",
		},
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	app.App.Configure("Render", &cfg)

	τ0 := time.Now()

	cpI, err := plugins.DefaultPluginStore.Create(cfg.CachePlugin.Plugin, cfg.CachePlugin.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	cp, ok := cpI.(gopnik.CachePluginInterface)
	if !ok {
		log.Fatal("Invalid cache plugin type")
	}

	removeDelay, err := time.ParseDuration(cfg.Render.HotCacheDelay)
	if err != nil {
		log.Fatalf("Invalid HotCacheDelay: %v", cfg.Render.HotCacheDelay)
	}
	ts, err := tileserver.NewTileServer(cfg.RenderPoolsConfig, cp, removeDelay)
	if err != nil {
		log.Fatalf("Failed to create tile server: %v", err)
	}

	readTimeout, err := time.ParseDuration(cfg.Render.HTTPReadTimeout)
	if err != nil {
		log.Fatalf("Invalid read timeout: %v", err)
	}
	writeTimeout, err := time.ParseDuration(cfg.Render.HTTPWriteTimeout)
	if err != nil {
		log.Fatalf("Invalid write timeout: %v", err)
	}
	δ := time.Since(τ0)
	log.Info("Done in %v seconds", δ.Seconds())

	// USR1, SIGHUP for style reload
	go sigHandler(syscall.SIGUSR1, ts.ReloadStyle, "ReloadStyle error")
	go sigHandler(syscall.SIGHUP, ts.ReloadStyle, "ReloadStyle error")

	servicestatus.SetOK() // Service is Ok if renders starts

	s := &http.Server{
		Addr:           cfg.Render.Addr,
		Handler:        ts,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Info("Starting on %s...", cfg.Render.Addr)
	log.Fatal(s.ListenAndServe())
}
