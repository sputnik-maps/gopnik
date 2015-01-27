package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"os"
	"sync"

	"app"
	"gopnik"
	"plugins"
	_ "plugins_enabled"

	"github.com/cheggaaa/pb"
	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

var planFile = flag.String("plan", "", "JSON plan file (see gopnikprerenderimport)")

type CopyConfig struct {
	From    app.PluginConfig
	To      app.PluginConfig
	Threads int
	Retries int
	Logging json.RawMessage
}

type Config struct {
	Copy CopyConfig
	app.CommonConfig
	json.OtherKeys
}

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

func copyMetaTile(metaCoord gopnik.TileCoord, cfg *Config, from, to gopnik.CachePluginInterface) {
	var metaTile []gopnik.Tile

	for y := uint(0); y < cfg.MetaSize; y++ {
		for x := uint(0); x < cfg.MetaSize; x++ {
			coord := gopnik.TileCoord{
				X:    metaCoord.X + uint64(x),
				Y:    metaCoord.Y + uint64(y),
				Zoom: metaCoord.Zoom,
				Size: 1,
				Tags: metaCoord.Tags,
			}

			attempt := 0
		TRYLOOP:
			for {
				if cfg.Copy.Retries > 0 {
					attempt++
				}

				var err error
				var rawTile []byte

				rawTile, err = from.Get(coord)
				if err != nil {
					if attempt <= cfg.Copy.Retries {
						log.Error("Get error: %v", err)
						continue
					} else {
						log.Fatalf("Get error: %v", err)
					}
				}

				tile, err := gopnik.NewTile(rawTile)
				if err != nil {
					if attempt <= cfg.Copy.Retries {
						log.Error("NewTile error: %v", err)
						continue
					} else {
						log.Fatalf("NewTile error: %v", err)
					}
				}

				metaTile = append(metaTile, *tile)
				break TRYLOOP
			}
		}
	}

	attempt := 0
TRYLOOP2:
	for {
		if cfg.Copy.Retries > 0 {
			attempt++
		}

		err := to.Set(metaCoord, metaTile)
		if err != nil {
			if attempt <= cfg.Copy.Retries {
				log.Error("Set error: %v", err)
				continue
			} else {
				log.Fatalf("Set error: %v", err)
			}
		}
		break TRYLOOP2
	}
}

func main() {
	// Init...
	cfg := Config{
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	app.App.Configure("Copy", &cfg)

	if cfg.Copy.Threads < 1 {
		cfg.Copy.Threads = 1
	}

	fromI, err := plugins.DefaultPluginStore.Create(cfg.Copy.From.Plugin, cfg.Copy.From.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	from, ok := fromI.(gopnik.CachePluginInterface)
	if !ok {
		log.Fatal("Invalid cache plugin type")
	}

	toI, err := plugins.DefaultPluginStore.Create(cfg.Copy.To.Plugin, cfg.Copy.To.PluginConfig)
	if err != nil {
		log.Fatal(err)
	}
	to, ok := toI.(gopnik.CachePluginInterface)
	if !ok {
		log.Fatal("Invalid cache plugin type")
	}

	// Load plan...
	coords, err := loadPlanFile()
	if err != nil {
		log.Fatal(err)
	}

	// Process...
	bar := pb.StartNew(len(coords))
	var barMu sync.Mutex
	var wg sync.WaitGroup

	for k := 0; k < cfg.Copy.Threads; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for i := k; i < len(coords); i += cfg.Copy.Threads {
				copyMetaTile(coords[i], &cfg, from, to)

				barMu.Lock()
				bar.Increment()
				barMu.Unlock()
			}
		}(k)
	}
	wg.Wait()
	bar.FinishPrint("Done")
}
