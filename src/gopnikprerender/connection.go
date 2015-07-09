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
	timeout          time.Duration
	transportFactory thrift.TTransportFactory
	protocolFactory  thrift.TProtocolFactory
	socket           *thrift.TSocket
	transport        thrift.TTransport
	renderClient     *gopnikrpc.RenderClient
}

func newConnection(addr string, timeout time.Duration) *connection {
	return &connection{
		addr:    addr,
		timeout: timeout,
	}
}

func (self *connection) Connect() error {
	log.Debug("Connecting to %v...", self.addr)

	// Creating connection
	var err error
	self.transportFactory = thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	self.protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	self.socket, err = thrift.NewTSocketTimeout(self.addr, self.timeout)
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
		var tmpErr error
		if e, ok := err.(*net.OpError); ok && e.Temporary() {
			tmpErr = e
		}
		if ex, ok := err.(thrift.TTransportException); ok && ex.TypeId() == thrift.TIMED_OUT {
			tmpErr = ex
		}
		if tmpErr != nil {
			log.Info("Connection to %v temorary fail (%v). Reconnecting...", self.addr, tmpErr)
			self.Close()
			err2 := self.Connect()
			if err2 != nil {
				return nil, err2
			}
		} else {
			return nil, err
		}
	}
}
