package tilerouter

import (
	"fmt"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

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

func (self *TileRouter) getTile(addr string, coord gopnik.TileCoord) (img []byte, err error) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	socket, err := thrift.NewTSocketTimeout(addr, self.timeout)
	if err != nil {
		return nil, fmt.Errorf("socket error: %v", err.Error())
	}
	transport := transportFactory.GetTransport(socket)
	defer transport.Close()
	err = transport.Open()
	if err != nil {
		return nil, fmt.Errorf("transport open error: %v", err.Error())
	}

	renderClient := gopnikrpc.NewRenderClientFactory(transport, protocolFactory)
	tiles, err := renderClient.Render(gopnikrpcutils.CoordToRPC(&coord), gopnikrpc.Priority_HIGH, false)
	if err != nil {
		return nil, err
	}

	if len(tiles) != 1 {
		return nil, fmt.Errorf("Invalid render response size %v", len(tiles))
	}

	return tiles[0].Image, err
}

func (self *TileRouter) Tile(coord gopnik.TileCoord) (img []byte, err error) {
	for i := 0; i < ATTEMPTS; i++ {
		addr := self.selector.SelectRender(coord)
		if addr == "" {
			img, err = nil, fmt.Errorf("No available renders")
			time.Sleep(10 * time.Second)
			continue
		}
		img, err = self.getTile(addr, coord)
		if err == nil {
			return
		}
	}

	return
}
