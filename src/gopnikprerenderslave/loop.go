package main

import (
	"encoding/gob"
	"expvar"
	"fmt"
	"net"
	"strings"
	"time"

	"app"
	"gopnik"
	"gopnikprerenderlib"
	"tilerender"
)

type zoomRender struct {
	Render  *tilerender.RenderPool
	MinZoom uint
	MaxZoom uint
}

type loop struct {
	tiles      chan *tilerender.RenderPoolResponse // Render results
	saverQueue chan *tilerender.RenderPoolResponse // Render results
	done       chan tileReport                     // Render + Saver results
	reqTasks   chan int                            // Request N new tasks
	closeConn  chan int                            // Connection closed message
	savers     []*saver                            // Save pool
	cache      gopnik.CachePluginInterface         // Cache plugin
	renders    []zoomRender                        // Renders
	renderCfg  app.RenderPoolsConfig               // Renders config
}

func sArrEq(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func newLoop(cache gopnik.CachePluginInterface, renderCfg app.RenderPoolsConfig, saverPoolSize int) (*loop, error) {
	l := new(loop)

	qSize := uint(0)
	for _, rCfg := range renderCfg.RenderPools {
		qSize += rCfg.PoolSize
	}
	qSize *= 100
	for _, rCfg := range renderCfg.RenderPools {
		if rCfg.QueueSize < qSize {
			qSize = rCfg.QueueSize
		}
	}

	l.tiles = make(chan *tilerender.RenderPoolResponse)
	l.saverQueue = make(chan *tilerender.RenderPoolResponse, qSize)
	l.done = make(chan tileReport, qSize)
	l.reqTasks = make(chan int, 10)
	l.closeConn = make(chan int)
	l.savers = make([]*saver, saverPoolSize)
	l.cache = cache

	l.reqTasks <- int(qSize)

	// Starting render threads
	l.renderCfg = renderCfg
	err := l.recreateRenders()
	if err != nil {
		return nil, err
	}

	// Starting saver threads
	for i := 0; i < len(l.savers); i++ {
		l.savers[i] = newSaver(cache)
		go l.savers[i].Worker(l.saverQueue, l.done)
	}
	go l.tilesTrapper()

	return l, nil
}

func (l *loop) tilesTrapper() {
	for tile := range l.tiles {
		l.reqTasks <- 1
		l.saverQueue <- tile
	}
}

func (l *loop) updateConfig(config *gopnikprerenderlib.RConfig) error {
	// Savers
	if config.SaverThreads < len(l.savers) {
		for i, s := range l.savers[config.SaverThreads:] {
			// Kill saver thread
			s.Stop()
			l.savers[i] = nil
		}
		// Remove dead savers
		l.savers = l.savers[:config.SaverThreads]
	}

	if config.SaverThreads > len(l.savers) {
		for len(l.savers) < config.SaverThreads {
			// Create new saver and start it
			s := newSaver(l.cache)
			l.savers = append(l.savers, s)
			go s.Worker(l.saverQueue, l.done)
		}
	}

	// Renders
	l.renderCfg = config.RenderPoolsConfig

	if len(l.renders) != len(config.RenderPools) {
		return l.recreateRenders()
	}
	for i := 0; i < len(l.renders); i++ {
		if sArrEq(l.renders[i].Render.Cmd(), config.RenderPools[i].Cmd) ||
			uint(l.renders[i].Render.QueueSize()) != config.RenderPools[i].QueueSize {
			return l.recreateRenders()
		}
	}

	for i := 0; i < len(l.renders); i++ {
		l.renders[i].MinZoom = config.RenderPools[i].MinZoom
		l.renders[i].MaxZoom = config.RenderPools[i].MaxZoom
		l.renders[i].Render.Resize(int(config.RenderPools[i].PoolSize))
		l.renders[i].Render.SetTTL(config.RenderPools[i].RenderTTL)
	}

	return nil
}

func (l *loop) recreateRenders() error {
	// Destroy
	for _, rend := range l.renders {
		rend.Render.Stop()
	}

	// Create
	l.renders = make([]zoomRender, len(l.renderCfg.RenderPools))
	for i := 0; i < len(l.renderCfg.RenderPools); i++ {
		var err error
		l.renders[i].MinZoom = l.renderCfg.RenderPools[i].MinZoom
		l.renders[i].MaxZoom = l.renderCfg.RenderPools[i].MaxZoom
		l.renders[i].Render, err = tilerender.NewRenderPool(
			l.renderCfg.RenderPools[i].Cmd,
			l.renderCfg.RenderPools[i].PoolSize,
			l.renderCfg.RenderPools[i].QueueSize,
			l.renderCfg.RenderPools[i].RenderTTL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *loop) writer(conn net.Conn) {
	enc := gob.NewEncoder(conn)

	// Send Hello message
	err := enc.Encode(&gopnikprerenderlib.RResponse{
		Type: gopnikprerenderlib.Hello,
		Hello: &gopnikprerenderlib.RHello{
			SaverPool:         len(l.savers),
			RenderPoolsConfig: l.renderCfg,
		},
	})
	if err != nil {
		log.Error("RPC Encode error: %v", err)
		log.Debug("Waiting for connection termination...")
		<-l.closeConn
		return
	}

	monitoringTicker := time.Tick(1 * time.Minute)

L:
	for {
		select {
		case reqNTasks := <-l.reqTasks:
			log.Debug("Request %v new tasks", reqNTasks)

			for i := 0; i < reqNTasks; i++ {
				var resp gopnikprerenderlib.RResponse
				resp.Type = gopnikprerenderlib.GetTask
				err = enc.Encode(&resp)
				if err != nil {
					log.Error("RPC Encode error: %v", err)
					conn.Close()
					break L
				}
			}
		case tr := <-l.done:
			log.Debug("Done: %v", tr)

			var resp gopnikprerenderlib.RResponse
			resp.Coord = &tr.Coord
			if tr.Error != nil {
				log.Error("Tile error: %v", tr.Error) // log err
				resp.Type = gopnikprerenderlib.Error
				resp.Error = tr.Error.Error()
			} else {
				resp.Type = gopnikprerenderlib.Stat
				resp.Stat = &gopnikprerenderlib.RStat{
					RenderTime: tr.RenderTime,
					SaveTime:   tr.SaveTime,
				}
			}
			err = enc.Encode(&resp)
			if err != nil {
				log.Error("RPC Encode error: %v", err)
				conn.Close()
				break L
			}
		case <-monitoringTicker:
			monData := expvar.Get("metrics").String()
			log.Debug("Monitoring: %v", monData)

			var resp gopnikprerenderlib.RResponse
			resp.Type = gopnikprerenderlib.Monitoring
			resp.Mon = monData
			err = enc.Encode(&resp)
			if err != nil {
				log.Error("RPC Encode error: %v", err)
				conn.Close()
				break L
			}
		case <-l.closeConn:
			log.Debug("Stop writer")
			return
		}
	}
	log.Debug("Waiting for connection termination...")
	<-l.closeConn
}

func (l *loop) Run(conn net.Conn) {
	dec := gob.NewDecoder(conn)

	go l.writer(conn)

	defer func() {
		l.closeConn <- 1
	}()

	// Waiting for commands
	for {
		var task gopnikprerenderlib.RTask
		err := dec.Decode(&task)
		if err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				log.Error("RPC Decode error: %v", err)
			}
			return
		}
		log.Debug("New task: %v", task)

		switch task.Type {
		case gopnikprerenderlib.RenderTask:
			if task.Coord != nil {
				err = l.enqueueRequest(*task.Coord, l.tiles)
				if err != nil {
					log.Error("EnqueueRequest error: %v", err)
				}
			} else {
				log.Error("Empty task!")
			}
		case gopnikprerenderlib.Config:
			if task.Config != nil {
				err := l.updateConfig(task.Config)
				if err != nil {
					log.Error("updateConfig error: %v", err)
					return
				}
			} else {
				log.Error("Empty configuration!")
			}
		default:
			log.Error("Invalid type %v", task.Type)
			return
		}
	}
}

func (l *loop) enqueueRequest(coord gopnik.TileCoord, resCh chan<- *tilerender.RenderPoolResponse) error {
	for _, rCfg := range l.renders {
		if coord.Zoom < uint64(rCfg.MinZoom) || coord.Zoom > uint64(rCfg.MaxZoom) {
			continue
		}
		return rCfg.Render.EnqueueRequest(coord, resCh)
	}
	return fmt.Errorf("Invalid zoom %v", coord.Zoom)
}

func (l *loop) Kill() {
	for _, render := range l.renders {
		render.Render.Stop()
	}
}
