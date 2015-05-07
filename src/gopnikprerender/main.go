package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"os"

	"app"
	"gopnik"
	"gopnikprerenderlib"
	"perflog"
	"plugins"
	_ "plugins_enabled"

	"github.com/cheggaaa/pb"
	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

var planFile = flag.String("plan", "", "JSON plan file (see gopnikprerenderimport)")

func loadPlanFile() (coords []gopnik.TileCoord, err error) {
	var fin *os.File
	fin, err = os.Open(*planFile)
	if err != nil {
		err = fmt.Errorf("Failed to load plan file: %v", err)
		return
	}
	dec := json.NewDecoder(fin)
	err = dec.Decode(&coords)
	if err != nil {
		err = fmt.Errorf("Failed to decode plan file: %v", err)
		return
	}
	// Ensure metatile coord
	for i, oldCoord := range coords {
		coords[i] = app.App.Metatiler().TileToMetatile(&oldCoord)
	}
	return
}

func main() {
	cfg := gopnikprerenderlib.PrerenderGlobalConfig{
		Prerender: gopnikprerenderlib.PrerenderConfig{
			UIAddr:        ":8088",
			DebugAddr:     ":8097",
			NodeQueueSize: 100,
		},
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	app.App.Configure("Prerender", &cfg)

	clI, err := plugins.DefaultPluginStore.Create(cfg.Prerender.Slaves.Plugin, cfg.Prerender.Slaves.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	cl, ok := clI.(gopnik.ClusterPluginInterface)
	if !ok {
		log.Fatal("Invalid cache plugin type")
	}

	coords, err := loadPlanFile()
	if err != nil {
		log.Fatal(err)
	}
	coordsLen := len(coords)
	slavesAddrs, err := cl.GetRenders()
	if err != nil {
		log.Fatalf("GetRenders error: %v", err)
	}

	// Setup perflog
	if cfg.Prerender.PerfLog != "" {
		perflog.SetupPerflog(cfg.Prerender.PerfLog)
	}

	// Plan
	coordinator := newCoordinator(slavesAddrs, cfg.Prerender.NodeQueueSize, coords)
	coords = nil
	resChan := coordinator.Start()

	// WebUI
	/*
		if cfg.Prerender.UIAddr != "" {
			cpI, err := plugins.DefaultPluginStore.Create(cfg.CachePlugin.Plugin, cfg.CachePlugin.PluginConfig)
			if err != nil {
				log.Fatal(err)
			}
			cp, ok := cpI.(gopnik.CachePluginInterface)
			if !ok {
				log.Fatal("Invalid cache plugin type")
			}

			go runWebUI(cfg.Prerender.UIAddr, coordinator, cp)
		}
	*/

	// Cli and log

	bar := pb.StartNew(coordsLen)
	for res := range resChan {
		perflog.SavePerf(res)
		bar.Increment()
	}

	bar.FinishPrint("Done")

	// Waiting...
	coordinator.Wait()
}
