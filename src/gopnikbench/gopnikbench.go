package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"app"
	"gopnik"
	"gproj"
	"tilerender"
)

var xMin = flag.Uint64("xMin", 0, "min x coord")
var yMin = flag.Uint64("yMin", 0, "min y coord")
var xMax = flag.Uint64("xMax", 0, "max x coord")
var yMax = flag.Uint64("yMax", 0, "max y coord")
var latMin = flag.Float64("latMin", 0.0, "min latitude")
var lonMin = flag.Float64("lonMin", 0.0, "min longitude")
var latMax = flag.Float64("latMax", 0.0, "max latitude")
var lonMax = flag.Float64("lonMax", 0.0, "max longitude")
var zoom = flag.Uint64("zoom", 0, "zoom")
var size = flag.Uint64("size", 8, "size")
var n = flag.Int("n", 1, "Iterations")
var showProgress = flag.Bool("progress", false, "Show progress")

func latLon() (minC, maxC *gopnik.TileCoord) {
	if *latMin != 0.0 && *lonMin != 0.0 && *latMax != 0.0 && *lonMax != 0.0 {
		c1 := gproj.FromLLToCoord(*latMin, *lonMin, *zoom)
		c2 := gproj.FromLLToCoord(*latMax, *lonMax, *zoom)
		maxC = &c1
		minC = &c2
		if maxC.X < minC.X {
			maxC.X, minC.X = minC.X, maxC.X
		}
		if maxC.Y < minC.Y {
			maxC.Y, minC.Y = minC.Y, maxC.Y
		}
	}
	return
}

func getCoords() (minC, maxC *gopnik.TileCoord) {
	minC, maxC = latLon()
	if minC == nil {
		minC = &gopnik.TileCoord{
			X:    *xMin,
			Y:    *yMin,
			Zoom: *zoom,
			Size: *size,
		}
	}
	if maxC == nil {
		maxC = &gopnik.TileCoord{
			X:    *xMax,
			Y:    *yMax,
			Zoom: *zoom,
			Size: *size,
		}
	}
	return
}

func getMetaCoords() (metaCoords []gopnik.TileCoord) {
	minC, maxC := getCoords()
	metaMinC := app.App.Metatiler().TileToMetatile(minC)
	metaMaxC := app.App.Metatiler().TileToMetatile(maxC)

	for x := metaMinC.X; x <= metaMaxC.X; x += metaMinC.Size {
		for y := metaMinC.Y; y <= metaMaxC.Y; y += metaMinC.Size {
			metaCoords = append(metaCoords, gopnik.TileCoord{
				X:    x,
				Y:    y,
				Zoom: metaMinC.Zoom,
				Size: metaMinC.Size,
			})
		}
	}

	return
}

type Config struct {
	app.CommonConfig
	app.RenderPoolsConfig
}

func main() {
	cfg := Config{
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	app.App.Configure("", &cfg)

	var slaveCmd []string
	for _, rCfg := range cfg.RenderPools {
		if *zoom < uint64(rCfg.MinZoom) || *zoom > uint64(rCfg.MaxZoom) {
			continue
		}
		slaveCmd = rCfg.Cmd
	}
	if slaveCmd == nil {
		log.Fatalf("Render configuration for zoom %v not found", *zoom)
	}

	metaCoords := getMetaCoords()

	render, err := tilerender.NewTileRender(slaveCmd)
	if err != nil {
		log.Fatalf("NewTileRender error: %v", err)
	}

	var totalTime time.Duration
	for i := 0; i < *n; i++ {
		before := time.Now()
		for _, mCoord := range metaCoords {
			_, err := render.RenderTiles(mCoord)
			if err != nil {
				log.Fatalf("RenderTile error: %v", err)
			}
		}
		totalTime += time.Since(before)
		if *showProgress {
			log.Printf("  -- %v%% done", 100*i / *n)
		}
	}
	totalTime = time.Duration(int64(totalTime) / int64(*n))
	fmt.Printf("%v metatiles, %v iterations. Avg time per bbox: %vs\n", len(metaCoords), *n, totalTime.Seconds())
}
