package tilerender

import (
	"fmt"
	"time"

	"github.com/op/go-logging"

	"gopnik"
	"gopnikrpc"
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
	hpTasks *renderQueue
	lpTasks *renderQueue
	cmd     []string
	ttl     uint
	renders []*renderWrapper
}

func NewRenderPool(cmd []string, poolSize, hpQueueSize, lpQueueSize, ttl uint) (*RenderPool, error) {
	self := &RenderPool{}
	self.cmd = cmd
	self.hpTasks = newRenderQueue(hpQueueSize)
	self.lpTasks = newRenderQueue(lpQueueSize)
	self.renders = make([]*renderWrapper, poolSize)
	self.ttl = ttl

	errCh := make(chan error, poolSize)
	for i, _ := range self.renders {
		go func(k int) {
			var err error
			self.renders[k], err = newRenderWrapper(self.hpTasks, self.lpTasks, self.cmd, self.ttl)
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

func (self *RenderPool) EnqueueRequest(coord gopnik.TileCoord, resCh chan<- *RenderPoolResponse, prio gopnikrpc.Priority) error {
	switch prio {
	case gopnikrpc.Priority_HIGH:
		return self.hpTasks.Push(coord, resCh)
	case gopnikrpc.Priority_LOW:
		return self.lpTasks.Push(coord, resCh)
	default:
		return fmt.Errorf("unknown priority: %v", prio)
	}
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
	return self.hpTasks.Size() + self.lpTasks.Size()
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
			render, err := newRenderWrapper(self.hpTasks, self.lpTasks, self.cmd, self.ttl)
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
