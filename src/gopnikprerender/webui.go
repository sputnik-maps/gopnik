package main

import (
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"strconv"

	"gopnik"
	"gopnikwebstatic"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/staticbin"
	"github.com/yosssi/rendergold"
)

type webUILogger struct {
}

func (self *webUILogger) Write(p []byte) (n int, err error) {
	log.Debug("%s", p)
	return len(p), nil
}

func runWebUI(addr string, p *coordinator, cache gopnik.CachePluginInterface) {
	m := staticbin.Classic(gopnikwebstatic.Asset)

	var logger webUILogger
	m.Map(stdlog.New(&logger, "[martini] ", stdlog.LstdFlags))

	m.Use(rendergold.Renderer(rendergold.Options{Asset: Asset}))

	m.Get("/", func(r rendergold.Render) {
		done, total := p.DoneTasks()
		progress := float64(done) / float64(total) * 100.0

		r.HTML(
			http.StatusOK,
			"status",
			map[string]interface{}{
				"Page":     "Status",
				"Total":    total,
				"Done":     done,
				"Progress": fmt.Sprintf("%.02f", progress),
			},
		)
	})

	m.Get("/nodes", func(r rendergold.Render) {
		type node struct {
			Addr   string
			Status bool
		}
		addrs := p.Nodes()
		nodes := make([]node, len(addrs))
		for i, addr := range addrs {
			nodes[i].Addr = addr
			nodes[i].Status = p.NodeStatus(addr)
		}

		r.HTML(
			http.StatusOK,
			"nodes",
			map[string]interface{}{
				"Page":  "Nodes",
				"Nodes": nodes,
			},
		)
	})

	m.Get("/charts/:addr", func(params martini.Params, r rendergold.Render) {
		addr := params["addr"]
		mon := p.NodeMonitor(addr)
		if mon == nil {
			r.Redirect("/nodes")
		}
		r.HTML(
			http.StatusOK,
			"charts",
			map[string]interface{}{
				"Page":    "Charts",
				"Addr":    addr,
				"Metrics": mon.Metrics(),
			},
		)
	})

	m.Get("/chart/:addr", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		addr := params["addr"]
		mon := p.NodeMonitor(addr)
		if mon == nil {
			res.WriteHeader(400)
			return
		}
		metricArgs := req.URL.Query()["metric"]
		if len(metricArgs) != 1 {
			res.WriteHeader(400)
			return
		}
		metric := metricArgs[0]
		points := mon.Points(metric)
		enc := json.NewEncoder(res)
		err := enc.Encode(points)
		if err != nil {
			log.Error("%v", err)
		}
	})

	m.Get("/map", func(r rendergold.Render) {
		r.HTML(
			http.StatusOK,
			"map",
			map[string]interface{}{
				"Page": "Map",
			},
		)
	})

	m.Get("/tiles/:z/:x/:y.png", func(params martini.Params) []byte {
		z, _ := strconv.ParseUint(params["z"], 10, 64)
		x, _ := strconv.ParseUint(params["x"], 10, 64)
		y, _ := strconv.ParseUint(params["y"], 10, 64)
		coord := gopnik.TileCoord{x, y, z, 1, nil}

		res, err := cache.Get(coord)
		if err != nil {
			log.Error("%v", err)
			return nil
		}
		if res == nil {
			res, err = gopnikwebstatic.Asset("img/notile.png")
			if err != nil {
				log.Error("%v", err)
				return nil
			}
		}
		return res
	})

	m.Get("/progressTasksCoord", func(r rendergold.Render) {
		progressTasksCoord := p.ProgressTasksCoord()

		r.HTML(
			http.StatusOK,
			"coordinates",
			map[string]interface{}{
				"Page": "In Progress",
				"ProgressTasksCoord":     progressTasksCoord,
			},
		)
	})

	log.Info("Starting WebUI on %v", addr)
	log.Fatal(http.ListenAndServe(addr, m))
}
