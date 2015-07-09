package tilerouter

import (
	"fmt"
	"sync"
	"time"

	"gopnik"
	"gopnikrpc"
	"gopnikrpcutils"
)

var ATTEMPTS = 2

type TileRouter struct {
	timeout     time.Duration
	pingPeriod  time.Duration
	rendersLock sync.RWMutex
	selector    *RenderSelector
}

func NewTileRouter(renders []string, timeout time.Duration, pingPeriod time.Duration) (*TileRouter, error) {
	self := &TileRouter{
		timeout:     timeout,
		pingPeriod:  pingPeriod,
		rendersLock: sync.RWMutex{},
	}

	self.UpdateRenders(renders)

	return self, nil
}

func (self *TileRouter) UpdateRenders(renders []string) {
	selector, err := NewRenderSelector(renders, self.pingPeriod, self.timeout)
	if err != nil {
		log.Error("Failed to recreate RenderSelector: %v", err)
	}

	self.rendersLock.Lock()
	defer self.rendersLock.Unlock()
	if self.selector != nil {
		self.selector.Stop()
	}
	self.selector = selector
}

func (self *TileRouter) getTile(conn *gopnikrpc.RenderClient, coord gopnik.TileCoord) (img []byte, err error) {
	resp, err := conn.Render(gopnikrpcutils.CoordToRPC(&coord), gopnikrpc.Priority_HIGH, false)
	if err != nil {
		return nil, err
	}

	if len(resp.Tiles) != 1 {
		return nil, fmt.Errorf("Invalid render response size %v", len(resp.Tiles))
	}

	return resp.Tiles[0].Image, err
}

func (self *TileRouter) Tile(coord gopnik.TileCoord) (img []byte, err error) {
	for i := 0; i < ATTEMPTS; i++ {
		var conn *thriftConn
		conn, err = self.selector.SelectRender(coord)
		if conn == nil {
			img, err = nil, fmt.Errorf("No available renders")
			time.Sleep(10 * time.Second)
			continue
		}
		img, err = self.getTile(conn.Client, coord)
		if err == nil {
			self.selector.FreeConnection(conn)
			return
		}
	}

	return
}
