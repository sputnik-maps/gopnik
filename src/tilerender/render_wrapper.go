package tilerender

import (
	"sync"
	"time"

	"gopnik"

	"github.com/orofarne/hmetrics2"
)

var hRenderTime = hmetrics2.MustRegisterPackageMetric("render_time", hmetrics2.NewHistogram()).(*hmetrics2.Histogram)
var hWaitTime = hmetrics2.MustRegisterPackageMetric("wait_time", hmetrics2.NewHistogram()).(*hmetrics2.Histogram)
var hQueueElems = hmetrics2.MustRegisterPackageMetric("queue_elems", hmetrics2.NewHistogram()).(*hmetrics2.Histogram)

type renderWrapper struct {
	render *TileRender
	tasks  *renderQueue
	cmd    []string
	ttl    uint
	ttlMu  sync.Mutex
	stop   chan int
}

func newRenderWrapper(tasks *renderQueue, cmd []string, ttl uint) (*renderWrapper, error) {
	self := new(renderWrapper)
	self.tasks = tasks
	self.cmd = cmd
	self.ttl = ttl
	self.stop = make(chan int)

	var err error
	self.render, err = NewTileRender(self.cmd)
	if err != nil {
		return nil, err
	}

	go self.worker()

	return self, nil
}

func (self *renderWrapper) Run() {
	go self.worker()
}

func (self *renderWrapper) Restart() {
	self.stop <- 2
}

func (self *renderWrapper) Stop() {
	self.stop <- 1
}

func (self *renderWrapper) SetTTL(ttl uint) {
	needRestart := false

	self.ttlMu.Lock()
	if self.ttl != ttl {
		self.ttl = ttl
		needRestart = true
	}
	self.ttlMu.Unlock()

	if needRestart {
		self.Restart()
	}
}

func (self *renderWrapper) renderOne(coord gopnik.TileCoord) *RenderPoolResponse {
	log.Debug("Rendering %v", coord)

	resp := new(RenderPoolResponse)
	resp.Coord = coord

	// Rendering
	timeBefore := time.Now()
	tiles, err := self.render.RenderTiles(coord)
	resp.RenderTime = time.Since(timeBefore)
	if err != nil {
		resp.Error = err
	} else {
		resp.Tiles = tiles
	}

	hRenderTime.AddPoint(resp.RenderTime.Seconds())
	log.Debug("Rendering %v done (%v seconds) %v",
		coord, resp.RenderTime.Seconds(), err)

	return resp
}

func (self *renderWrapper) killRender(restart bool) {
	if self.render != nil {
		self.render.Stop()
		self.render = nil

		if restart {
			self.startRender()
		}
	}
}

func (self *renderWrapper) startRender() {
	for self.render == nil {
		render, err := NewTileRender(self.cmd)
		if err != nil {
			log.Error("Failed to create render: %v")
			time.Sleep(5 * time.Second)
		} else {
			self.render = render
			go self.worker()
			return
		}
	}
}

func (self *renderWrapper) worker() {
	var stopFlag int
	tasks := self.tasks.TasksChan()

	self.ttlMu.Lock()
	ttl := self.ttl
	self.ttlMu.Unlock()

	waitTimeBefore := time.Now()

L:
	for {
		select {
		case stopFlag = <-self.stop:
			// stop signal recieved
			break L
		case task, ok := <-tasks:
			if !ok {
				// chan is closed
				break L
			}
			// Calculate statistics
			waitTime := time.Since(waitTimeBefore)
			hQueueElems.AddPoint(float64(len(tasks)))
			// Process task
			resp := self.renderOne(task)
			// Attach wait time to response
			resp.WaitTime = waitTime
			// Copy to global statistics
			hWaitTime.AddPoint(waitTime.Seconds())
			// Reset WaitTime timer
			waitTimeBefore = time.Now()
			// Cleanup wait queue
			if err := self.tasks.Done(task, resp); err != nil {
				log.Error("RenderPoolQueue: %v", err)
			}
			// Check TTL
			if ttl > 0 {
				ttl--
				if ttl == 0 {
					stopFlag = 2
					break L
				}
			}
		}
	}

	self.killRender(stopFlag == 2)
}
