package app

import (
	"encoding/xml"
	"fmt"
	stdlog "log"
	"math"
	"net/http"
	"sync"

	"github.com/go-martini/martini"
	"github.com/orofarne/hmetrics2"

	"program_version"
	"servicestatus"
)

type xmlstatusLogger struct {
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
		w.Header().Add("Content-type", "text/xml")
		w.Write([]byte(xml.Header))
		w.Write([]byte("<stat>\n"))
		for k, v := range data {
			w.Write([]byte(fmt.Sprintf("    <%s>%v</%s>\n", k, v, k)))
		}
		w.Write([]byte("</stat>\n"))
	})

	return m
}
