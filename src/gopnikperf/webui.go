package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gopnik"
	"perflog"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/staticbin"
	json "github.com/orofarne/strict-json"
	"github.com/yosssi/rendergold"
)

type kv struct {
	K string
	V string
}

type zoomStats struct {
	Zoom  uint64
	Stats []kv
}

type zoomRange struct {
	MinZoom uint64
	MaxZoom uint64
}

func getStats(perfData []perflog.PerfLogEntry, zoom *uint64) (stats []kv) {
	totalRenderTimeS := float64(0)
	totalSaveTimeS := float64(0)
	count := uint64(0)

	for _, entry := range perfData {
		if zoom != nil && entry.Coord.Zoom != *zoom {
			continue
		}
		count++
		totalRenderTimeS += entry.RenderTime.Seconds()
		totalSaveTimeS += entry.SaverTime.Seconds()
	}

	stats = append(stats, kv{
		K: "Metatiles",
		V: fmt.Sprintf("%v", count),
	})
	stats = append(stats, kv{
		K: "Total render time",
		V: fmt.Sprintf("%0.2fs", totalRenderTimeS),
	})
	stats = append(stats, kv{
		K: "Avg render time",
		V: fmt.Sprintf("%0.2fs", totalRenderTimeS/float64(count)),
	})
	stats = append(stats, kv{
		K: "Total save time",
		V: fmt.Sprintf("%0.2fs", totalSaveTimeS),
	})
	stats = append(stats, kv{
		K: "Avg save time",
		V: fmt.Sprintf("%0.2fs", totalSaveTimeS/float64(count)),
	})

	return
}

func getZooms(perfData []perflog.PerfLogEntry) (zrng zoomRange) {
	if len(perfData) < 1 {
		return
	}

	zrng.MaxZoom = perfData[0].Coord.Zoom
	zrng.MinZoom = perfData[0].Coord.Zoom

	for i := 1; i < len(perfData); i++ {
		zoom := perfData[i].Coord.Zoom
		if zoom < zrng.MinZoom {
			zrng.MinZoom = zoom
		}
		if zoom > zrng.MaxZoom {
			zrng.MaxZoom = zoom
		}
	}
	return
}

func runWebUI(addr string, perfData []perflog.PerfLogEntry) {
	m := staticbin.Classic(Asset)

	m.Map(log.New(os.Stderr, "", log.LstdFlags))

	m.Use(rendergold.Renderer(rendergold.Options{Asset: Asset}))

	m.Get("/", func(r rendergold.Render) {

		totalStats := getStats(perfData, nil)
		var zoomsStats []zoomStats
		zrng := getZooms(perfData)
		for z := zrng.MinZoom; z <= zrng.MaxZoom; z++ {
			zoomsStats = append(zoomsStats, zoomStats{
				Zoom:  z,
				Stats: getStats(perfData, &z),
			})
		}

		r.HTML(
			http.StatusOK,
			"index",
			map[string]interface{}{
				"Page":       "Results",
				"TotalStats": totalStats,
				"ZoomsStats": zoomsStats,
			},
		)
	})

	m.Get("/heatmap", func(r rendergold.Render) {
		r.HTML(
			http.StatusOK,
			"heatmap",
			map[string]interface{}{
				"Page": "Heat",
			},
		)
	})

	m.Get("/heattiles_zooms", func(res http.ResponseWriter) {
		var zooms struct {
			Min uint64
			Max uint64
		}

		if len(perfData) > 0 {
			zooms.Min = perfData[0].Coord.Zoom
			zooms.Max = perfData[0].Coord.Zoom

			for i := 1; i < len(perfData); i++ {
				zoom := perfData[i].Coord.Zoom
				if zoom < zooms.Min {
					zooms.Min = zoom
				}
				if zoom > zooms.Max {
					zooms.Max = zoom
				}
			}
		}

		enc := json.NewEncoder(res)
		if err := enc.Encode(zooms); err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
	})

	m.Get("/heattiles/:zoom_orig/:zoom/:x/:y.png", func(params martini.Params, res http.ResponseWriter) {
		var coord gopnik.TileCoord

		coord.Size = 1
		_, err := fmt.Sscan(params["zoom"], &coord.Zoom)
		if err != nil {
			http.Error(res, err.Error(), 400)
			return
		}
		_, err = fmt.Sscan(params["x"], &coord.X)
		if err != nil {
			http.Error(res, err.Error(), 400)
			return
		}
		_, err = fmt.Sscan(params["y"], &coord.Y)
		if err != nil {
			http.Error(res, err.Error(), 400)
			return
		}
		var zoomOrig uint64
		_, err = fmt.Sscan(params["zoom_orig"], &zoomOrig)
		if err != nil {
			http.Error(res, err.Error(), 400)
			return
		}

		tile, err := genPerfTile(perfData, coord, zoomOrig)
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}

		res.Header().Set("Content-Type", "image/png")
		res.Write(tile)
	})

	log.Printf("Starting WebUI on %v", addr)
	log.Fatal(http.ListenAndServe(addr, m))
}
