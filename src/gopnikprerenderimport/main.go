package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"app"
	"gopnik"
	"gopnikprerenderlib"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

var planFile = flag.String("plan", "", "JSON plan file (see gopnikprerenderimport)")
var zoom = flag.String("zoom", "", "Zoom to render. Format: 1-9,14,16")

var csvFile = flag.String("csv", "", "CSV file with bboxes")

var latMin = flag.Float64("latMin", 0.0, "min latitude")
var lonMin = flag.Float64("lonMin", 0.0, "min longitude")
var latMax = flag.Float64("latMax", 0.0, "max latitude")
var lonMax = flag.Float64("lonMax", 0.0, "max longitude")

var allWorld = flag.Bool("world", false, "render world for current zoom")

var doAppend = flag.Bool("append", false, "Append data to plan file")

var tagsF = flag.String("tags", "", "comma separetad list of tags")

func parseZoom(zoomString string) (res []uint64, err error) {
	elems := strings.Split(zoomString, ",")
	for _, elem := range elems {
		if strings.Contains(elem, "-") {
			var z1, z2 uint64
			_, err = fmt.Sscanf(elem, "%v-%v", &z1, &z2)
			if err != nil {
				return
			}
			if z1 > z2 {
				err = fmt.Errorf("Invalid zoom: '%v'", elem)
				return
			}
			for z := z1; z <= z2; z++ {
				res = append(res, z)
			}
		} else {
			var z uint64
			_, err = fmt.Sscan(elem, &z)
			if err != nil {
				return
			}
			res = append(res, z)
		}
	}
	return
}

func checkLat(lat float64) {
	if lat < -85.0 || lat > 85.0 {
		log.Fatalf("Invalid lat: %v", lat)
	}
}

func checkLon(lon float64) {
	if lon < -180.0 || lon > 180.0 {
		log.Fatalf("Invalid lon: %v", lon)
	}
}

func latLon(zoom uint64) (coords []gopnik.TileCoord, err error) {
	if *latMin != 0.0 && *lonMin != 0.0 && *latMax != 0.0 && *lonMax != 0.0 {
		checkLat(*latMin)
		checkLat(*latMax)
		checkLon(*lonMin)
		checkLon(*lonMax)
		bbox := [4]float64{*latMin, *latMax, *lonMin, *lonMax}
		coords, err = genCoords(bbox, zoom)
	}
	return
}

func main() {
	cfg := gopnikprerenderlib.PrerenderGlobalConfig{
		Prerender: gopnikprerenderlib.PrerenderConfig{
			UIAddr:    ":8088",
			DebugAddr: ":8098",
		},
		CommonConfig: app.CommonConfig{
			MetaSize: 8,
			TileSize: 256,
		},
	}

	app.App.Configure("Prerender", &cfg)

	zooms, err := parseZoom(*zoom)
	if err != nil {
		log.Fatalf("Invalid zoom: %v", err)
	}

	var tags []string
	if *tagsF != "" {
		tags = strings.Split(*tagsF, ",")
	}

	var coords []gopnik.TileCoord

	if *doAppend {
		f, err := os.Open(*planFile)
		if err != nil {
			if !strings.Contains(err.Error(), "no such file or directory") {
				log.Fatalf("Failed to open plan file: %v", err)
			}
		} else {
			dec := json.NewDecoder(f)
			err = dec.Decode(&coords)
			if err != nil {
				log.Fatalf("Failed to parse plan file: %v", err)
			}
		}
	}

	if *allWorld {
		for _, zoom := range zooms {
			bbox := [4]float64{-85 /*latMin*/, 85 /*latMax*/, -180 /*lonMin*/, 180 /*lonMax*/}
			zoomCoords, err := genCoords(bbox, zoom)
			if err != nil {
				log.Fatal(err)
			}
			// append tags
			for i := 0; i < len(zoomCoords); i++ {
				zoomCoords[i].Tags = tags
			}
			coords = append(coords, zoomCoords...)
		}
	} else {
		if *csvFile != "" {
			csvCoords, err := readCSVFile(*csvFile, zooms)
			if err != nil {
				log.Fatal(err)
			}
			// append tags
			for i := 0; i < len(csvCoords); i++ {
				csvCoords[i].Tags = tags
			}
			coords = append(coords, csvCoords...)
		}

		for _, zoom := range zooms {
			flagCoords, err := latLon(zoom)
			if err != nil {
				log.Fatal(err)
			}
			// append tags
			for i := 0; i < len(flagCoords); i++ {
				flagCoords[i].Tags = tags
			}
			coords = append(coords, flagCoords...)
		}
	}

	fout, err := os.OpenFile(*planFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open plan file: %v", err)
	}
	enc := json.NewEncoder(fout)
	err = enc.Encode(coords)
	if err != nil {
		log.Fatalf("Failed to encode plan file: %v", err)
	}
}
