package main

import (
	"net/http"
	"runtime"
	"time"

	"app"
	"gopnik"
	"plugins"
	_ "plugins_enabled"
	"tilerouter"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

type DispatcherConfig struct {
	Addr             string           // Bind addr
	DebugAddr        string           // Address for statistics
	HTTPReadTimeout  string           //
	HTTPWriteTimeout string           //
	RequestTimeout   string           // Timeout for request to gopnik
	PingPeriod       string           // Ping gopniks every PingPeriod
	Threads          int              // GOMAXPROCS
	Logging          json.RawMessage  // see loghelper.go
	ClusterPlugin    app.PluginConfig // dynamic rendering cluster
	FilterPlugin     app.PluginConfig // coordinate filter
}

type Config struct {
	Dispatcher       DispatcherConfig //
	CachePlugin      app.PluginConfig //
	app.CommonConfig                  //
	json.OtherKeys                    //
}

func main() {
	cfg := Config{
		Dispatcher: DispatcherConfig{
			Addr:             ":8080",
			DebugAddr:        ":9080",
			HTTPReadTimeout:  "60s",
			HTTPWriteTimeout: "60s",
			RequestTimeout:   "600s",
			PingPeriod:       "30s",
			Threads:          runtime.NumCPU(),
			FilterPlugin: app.PluginConfig{
				Plugin: "EmptyFilterPlugin",
			},
		},
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	app.App.Configure("Dispatcher", &cfg)

	clI, err := plugins.DefaultPluginStore.Create(cfg.Dispatcher.ClusterPlugin.Plugin, cfg.Dispatcher.ClusterPlugin.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	cl, ok := clI.(gopnik.ClusterPluginInterface)
	if !ok {
		log.Fatal("Invalid cache plugin type")
	}

	cpI, err := plugins.DefaultPluginStore.Create(cfg.CachePlugin.Plugin, cfg.CachePlugin.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	cp, ok := cpI.(gopnik.CachePluginInterface)
	if !ok {
		log.Fatal("Invalid cache plugin type")
	}

	filterI, err := plugins.DefaultPluginStore.Create(cfg.Dispatcher.FilterPlugin.Plugin, cfg.Dispatcher.FilterPlugin.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	filter, ok := filterI.(gopnik.FilterPluginInterface)
	if !ok {
		log.Fatal("Invalid filter plugin type")
	}

	readTimeout, err := time.ParseDuration(cfg.Dispatcher.HTTPReadTimeout)
	if err != nil {
		log.Fatalf("Invalid read timeout: %v", err)
	}
	writeTimeout, err := time.ParseDuration(cfg.Dispatcher.HTTPWriteTimeout)
	if err != nil {
		log.Fatalf("Invalid write timeout: %v", err)
	}
	requestTimeout, err := time.ParseDuration(cfg.Dispatcher.RequestTimeout)
	if err != nil {
		log.Fatalf("Invalid request timeout: %v", err)
	}
	pingPeriod, err := time.ParseDuration(cfg.Dispatcher.PingPeriod)
	if err != nil {
		log.Fatalf("Invalid ping period: %v", err)
	}

	rs, err := tilerouter.NewRouterServer(cl, cp, filter, requestTimeout, pingPeriod)
	if err != nil {
		log.Fatalf("Failed to start router: %v", err)
	}

	s := &http.Server{
		Addr:           cfg.Dispatcher.Addr,
		Handler:        rs,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Info("Starting on %s...", cfg.Dispatcher.Addr)
	log.Fatal(s.ListenAndServe())
}
