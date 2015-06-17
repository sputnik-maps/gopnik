package main

import (
	"fmt"
	"net"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

	"gopnik"
	"gopnikrpc"
	"gopnikrpcutils"
	"perflog"
)

type connection struct {
	addr             string
	transportFactory thrift.TTransportFactory
	protocolFactory  thrift.TProtocolFactory
	socket           *thrift.TSocket
	transport        thrift.TTransport
	renderClient     *gopnikrpc.RenderClient
}

func newConnection(addr string) *connection {
	return &connection{
		addr: addr,
	}
}

func (self *connection) Connect() error {
	log.Debug("Connecting to %v...", self.addr)

	// Creating connection
	var err error
	self.transportFactory = thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	self.protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	self.socket, err = thrift.NewTSocket(self.addr)
	if err != nil {
		return fmt.Errorf("socket error: %v", err.Error())
	}
	self.transport = self.transportFactory.GetTransport(self.socket)
	err = self.transport.Open()
	if err != nil {
		return fmt.Errorf("transport open error: %v", err.Error())
	}
	self.renderClient = gopnikrpc.NewRenderClientFactory(self.transport, self.protocolFactory)
	return nil
}

func (self *connection) Close() {
	self.transport.Close()
}

func (self *connection) Client() *gopnikrpc.RenderClient {
	return self.renderClient
}

func (self *connection) callRender(coord gopnik.TileCoord) (*perflog.PerfLogEntry, error) {
	resp, err := self.renderClient.Render(gopnikrpcutils.CoordToRPC(&coord), gopnikrpc.Priority_LOW, true)
	if err != nil {
		return nil, err
	}

	return &perflog.PerfLogEntry{
		Coord:      coord,
		Timestamp:  time.Now(),
		RenderTime: time.Duration(resp.RenderTime),
		SaverTime:  time.Duration(resp.SaveTime),
	}, nil
}

func (self *connection) ProcessTask(coord gopnik.TileCoord) (*perflog.PerfLogEntry, error) {
	for {
		log.Debug("Sending %v to %v...", coord, self.addr)
		res, err := self.callRender(coord)
		if err == nil {
			return res, nil
		}
		if e, ok := err.(*net.OpError); ok && e.Temporary() {
			log.Debug("Connection to %v temorary fail. Reconnecting...", self.addr)
			self.Close()
			err2 := self.Connect()
			if err2 != nil {
				return nil, err2
			}
		}
		return nil, err
	}
}
