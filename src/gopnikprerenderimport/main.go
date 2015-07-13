package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"app"
	"gopnik"
	"gopnikprerenderlib"
	"perflog"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

var planFile = flag.String("plan", "", "JSON plan file (see gopnikprerenderimport)")
var zoom = flag.String("zoom", "", "Zoom to render. Format: 1-9,14,16")

var csvFile = flag.String("csv", "", "CSV file with bboxes")
var urlsFile = flag.String("urls", "", "file with urls")

var latMin = flag.Float64("latMin", 0.0, "min latitude")
var lonMin = flag.Float64("lonMin", 0.0, "min longitude")
var latMax = flag.Float64("latMax", 0.0, "max latitude")
var lonMax = flag.Float64("lonMax", 0.0, "max longitude")

var allWorld = flag.Bool("world", false, "render world for current zoom")

var doAppend = flag.Bool("append", false, "Append data to plan file")

var tagsF = flag.String("tags", "", "comma separetad list of tags")

var subtractPerflog = flag.String("subtractPerflog", "", "subtract perflog items")
var subtractPerflogSince = flag.String("subtractPerflogSince", "", "Use values since time (RFC3339 or @UnixTimestamp)")
var subtractPerflogUntil = flag.String("subtractPerflogUntil", "", "Use values until time (RFC3339 or @UnixTimestamp)")

var shuffle = flag.Bool("shuffle", false, "Shuffle plan")

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

func parseTime(str string) (time.Time, error) {
	if str == "" {
		return time.Time{}, nil
	}

	if str[0] == '@' {
		var unixts int64
		_, err := fmt.Sscanf(str, "@%v", &unixts)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(unixts, 0), nil
	} else {
		return time.Parse(time.RFC3339, str)
	}
}

func subtractPerflogItems(coords []gopnik.TileCoord, perflogData []perflog.PerfLogEntry) []gopnik.TileCoord {
	for i := 0; i < len(coords); i++ {
		for _, logLine := range perflogData {
			if coords[i].Equals(&logLine.Coord) {
				coords = coords[:i+copy(coords[i:], coords[i+1:])]
				i--
				break
			}
		}
	}
	return coords
}

func shuffleCoords(coords []gopnik.TileCoord) {
	for i := range coords {
		j := rand.Intn(i + 1)
		coords[i], coords[j] = coords[j], coords[i]
	}
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

	if *urlsFile != "" {
		urlsCoords, err := readUrlsFile(*urlsFile)
		if err != nil {
			log.Fatal(err)
		}
		coords = append(coords, urlsCoords...)
	} else {
		zooms, err := parseZoom(*zoom)
		if err != nil {
			log.Fatalf("Invalid zoom: %v", err)
		}

		var tags []string
		if *tagsF != "" {
			tags = strings.Split(*tagsF, ",")
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
	}

	if *subtractPerflog != "" {
		sinceT, err := parseTime(*subtractPerflogSince)
		if err != nil {
			log.Fatalf("Invalid since param: %v", err)
		}
		untilT, err := parseTime(*subtractPerflogUntil)
		if err != nil {
			log.Fatalf("Invalid unitl param: %v", err)
		}

		perfData, err := perflog.LoadPerf(*subtractPerflog, sinceT, untilT)
		if err != nil {
			log.Fatalf("Failed to load performance log: %v", err)
		}

		coords = subtractPerflogItems(coords, perfData)
	}

	if *shuffle {
		rand.Seed(time.Now().UnixNano())
		shuffleCoords(coords)
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
