package main

import (
	"net"
	"runtime"
	"time"

	"app"
	"gopnik"
	"gopnikprerenderlib"
	"perflog"
	"plugins"
	_ "plugins_enabled"
	"servicestatus"

	"github.com/op/go-logging"
	"github.com/orofarne/hmetrics2"
)

var log = logging.MustGetLogger("global")

// HMetrics
var Stats struct {
	RenderT     *hmetrics2.Histogram
	SaverT      *hmetrics2.Histogram
	RenderWaitT *hmetrics2.Histogram
	SaverWaitT  *hmetrics2.Histogram
	SaverQueue  *hmetrics2.Histogram
}

func init() {
	Stats.RenderT = hmetrics2.NewHistogram()
	hmetrics2.MustRegisterPackageMetric("render_time", Stats.RenderT)
	Stats.SaverT = hmetrics2.NewHistogram()
	hmetrics2.MustRegisterPackageMetric("saver_time", Stats.SaverT)
	Stats.RenderWaitT = hmetrics2.NewHistogram()
	hmetrics2.MustRegisterPackageMetric("render_wait_time", Stats.RenderWaitT)
	Stats.SaverWaitT = hmetrics2.NewHistogram()
	hmetrics2.MustRegisterPackageMetric("saver_wait_time", Stats.SaverWaitT)
	Stats.SaverQueue = hmetrics2.NewHistogram()
	hmetrics2.MustRegisterPackageMetric("saver_queue_elems", Stats.SaverQueue)
}

func main() {
	cfg := gopnikprerenderlib.PrerenderGlobalConfig{
		PrerenderSlave: gopnikprerenderlib.PrerenderSlaveConfig{
			SaverPoolSize: runtime.NumCPU(),
			RPCAddr:       ":8095",
			DebugAddr:     ":8096",
		},
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	app.App.Configure("PrerenderSlave", &cfg)

	log.Info("Preparing...")

	cpI, err := plugins.DefaultPluginStore.Create(cfg.CachePlugin.Plugin, cfg.CachePlugin.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	cp, ok := cpI.(gopnik.CachePluginInterface)
	if !ok {
		log.Fatal("Invalid cache plugin type")
	}

	// Setup perflog
	if cfg.PrerenderSlave.PerfLog != "" {
		perflog.SetupPerflog(cfg.PrerenderSlave.PerfLog)
	}

	servicestatus.SetOK() // Service is Ok if renders starts

	// Listen RPC addr
	log.Info("Starting on %v", cfg.PrerenderSlave.RPCAddr)
	ln, err := net.Listen("tcp", cfg.PrerenderSlave.RPCAddr)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Accept error: %v", err)
			continue
		}
		log.Info("New connection from %v", conn.RemoteAddr())

		log.Info("Preparing loop...")
		τ0 := time.Now()
		l, err := newLoop(cp, cfg.RenderPoolsConfig, cfg.PrerenderSlave.SaverPoolSize)
		if err != nil {
			log.Fatalf("Failed to create event loop: %v", err)
		}
		δ := time.Since(τ0)
		log.Info("Done in %v seconds", δ.Seconds())

		l.Run(conn)

		l.Kill()
		log.Info("Connection %v closed", conn.RemoteAddr())
	}
}
