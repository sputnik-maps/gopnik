package app

import (
	"encoding/xml"
	"fmt"
	stdlog "log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-martini/martini"
	"github.com/orofarne/hmetrics2"

	"program_version"
	"servicestatus"
)

var startTime time.Time

type xmlstatusLogger struct {
}

func init() {
	startTime = time.Now()
}

func (self *xmlstatusLogger) Write(p []byte) (n int, err error) {
	log.Debug("%s", p)
	return len(p), nil
}

func CreateXMLStatusHandler() http.Handler {
	m := martini.Classic()

	var mu sync.Mutex
	var data = make(map[string]float64)
	hmetrics2.AddHook(func(newData map[string]float64) {
		mu.Lock()
		defer mu.Unlock()
		data = make(map[string]float64)
		for k, v := range newData {
			if !math.IsNaN(v) && !math.IsInf(v, 0) {
				data[k] = v
			}
		}
	})

	var logger xmlstatusLogger
	m.Map(stdlog.New(&logger, "[martini] ", stdlog.LstdFlags))

	m.Get("/mon", func() string {
		return servicestatus.GetString()
	})

	m.Get("/version", func() string {
		return program_version.GetVersion()
	})

	m.Get("/config", func(w http.ResponseWriter) {
		w.Header().Add("Content-type", "application/json")
		w.Write([]byte(App.Config()))
	})

	m.Get("/stat", func(w http.ResponseWriter) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Add("Content-type", "application/xml")
		w.Write([]byte(xml.Header))
		w.Write([]byte(fmt.Sprintf("<stat:document xmlns:stat=\"http://xml.sputnik.ru/stat\" name=\"%v\" version=\"%v\">\n", os.Args[0], program_version.GetVersion())))
		w.Write([]byte(fmt.Sprintf("  <stat>\n    <start_time>%v</start_time>\n  </stat>\n", startTime)))
		w.Write([]byte("  <user>\n"))
		for k, v := range data {
			w.Write([]byte(fmt.Sprintf("    <%s>%v</%s>\n", k, v, k)))
		}
		w.Write([]byte("  </user>\n"))
		w.Write([]byte("</stat:document>\n"))
	})

	return m
}
