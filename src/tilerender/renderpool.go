package tilerender

import (
	"fmt"
	"time"

	"github.com/op/go-logging"

	"gopnik"
)

var log = logging.MustGetLogger("global")

type RenderPoolResponse struct {
	Coord      gopnik.TileCoord
	Error      error
	Tiles      []gopnik.Tile
	RenderTime time.Duration
	WaitTime   time.Duration
}

type RenderPool struct {
	tasks   *renderQueue
	cmd     []string
	ttl     uint
	renders []*renderWrapper
}

func NewRenderPool(cmd []string, poolSize, queueSize, ttl uint) (*RenderPool, error) {
	self := &RenderPool{}
	self.cmd = cmd
	self.tasks = newRenderQueue(queueSize)
	self.renders = make([]*renderWrapper, poolSize)
	self.ttl = ttl

	errCh := make(chan error, poolSize)
	for i, _ := range self.renders {
		go func(k int) {
			var err error
			self.renders[k], err = newRenderWrapper(self.tasks, self.cmd, self.ttl)
			errCh <- err
		}(i)
	}

	var errs []error
	for i := uint(0); i < poolSize; i++ {
		err := <-errCh
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		for _, render := range self.renders {
			if render != nil {
				render.Stop()
			}
		}
		return nil, fmt.Errorf("Failed to create some renders: %v", errs)
	}

	return self, nil
}

func (self *RenderPool) EnqueueRequest(coord gopnik.TileCoord, resCh chan<- *RenderPoolResponse) error {
	return self.tasks.Push(coord, resCh)
}

func (self *RenderPool) Reload() {
	for _, render := range self.renders {
		render.Restart()
	}
}
func (self *RenderPool) Stop() {
	for _, render := range self.renders {
		render.Stop()
	}
}
func (self *RenderPool) Size() int {
	return len(self.renders)
}

func (self *RenderPool) QueueSize() int {
	return self.tasks.Size()
}

func (self *RenderPool) Cmd() []string {
	return self.cmd
}

func (self *RenderPool) Resize(newPoolSize int) {
	if newPoolSize < len(self.renders) {
		for i := newPoolSize; i < len(self.renders); i++ {
			self.renders[i].Stop()
			self.renders[i] = nil
		}
		self.renders = self.renders[:newPoolSize]
	}
	for newPoolSize > len(self.renders) {
		for {
			render, err := newRenderWrapper(self.tasks, self.cmd, self.ttl)
			if err != nil {
				log.Error("Failed to create render: %v", err)
			} else {
				self.renders = append(self.renders, render)

				break
			}
		}
	}
}

func (self *RenderPool) SetTTL(ttl uint) {
	for _, render := range self.renders {
		render.SetTTL(ttl)
	}
}
