package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"

	"app"
	"gopnik"
	"perflog"
	"plugins"
	_ "plugins_enabled"
	"servicestatus"
	"tileserver"
)

var log = logging.MustGetLogger("global")

type RenderConfig struct {
	Threads       int             // Use >= n threads
	Addr          string          // Bind addr
	DebugAddr     string          // Address for statistics
	HotCacheDelay string          // Time period after cache set is done and before drop hot cache
	PerfLog       string          //
	Logging       json.RawMessage // see loghelper.go
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
			Threads:       1,
			Addr:          ":8090",
			DebugAddr:     ":9090",
			HotCacheDelay: "0s",
		},
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	if cfg.Render.PerfLog != "" {
		perflog.SetupPerflog(cfg.Render.PerfLog)
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

	δ := time.Since(τ0)
	log.Info("Done in %v seconds", δ.Seconds())

	// USR1, SIGHUP for style reload
	go sigHandler(syscall.SIGUSR1, ts.ReloadStyle, "ReloadStyle error")
	go sigHandler(syscall.SIGHUP, ts.ReloadStyle, "ReloadStyle error")

	servicestatus.SetOK() // Service is Ok if renders starts

	log.Info("Starting on %s...", cfg.Render.Addr)
	log.Fatal(tileserver.RunServer(cfg.Render.Addr, ts))
}
