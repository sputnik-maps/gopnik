package main

import (
	"fmt"
	"gopnikrpc"
	"gopnikrpcutils"
	"perflog"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

	"gopnik"
)

type coordinator struct {
	addrs         []string
	connsWg       sync.WaitGroup
	tasks         *plan
	results       chan perflog.PerfLogEntry
	nodeQueueSize int
}

func newCoordinator(addrs []string, nodeQueueSize int, bboxes []gopnik.TileCoord) *coordinator {
	self := new(coordinator)
	self.addrs = addrs
	self.results = make(chan perflog.PerfLogEntry)
	self.nodeQueueSize = nodeQueueSize

	self.tasks = newPlan(bboxes)

	return self
}

func (self *coordinator) connSub(addr string) error {
	// Creating connection
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	socket, err := thrift.NewTSocket(addr)
	if err != nil {
		return fmt.Errorf("socket error: %v", err.Error())
	}
	transport := transportFactory.GetTransport(socket)
	defer transport.Close()
	err = transport.Open()
	if err != nil {
		return fmt.Errorf("transport open error: %v", err.Error())
	}
	renderClient := gopnikrpc.NewRenderClientFactory(transport, protocolFactory)

	for {
		coord := self.tasks.GetTask()
		if coord == nil {
			break
		}

		resp, err := renderClient.Render(gopnikrpcutils.CoordToRPC(coord), gopnikrpc.Priority_LOW, true)
		if err != nil {
			self.tasks.FailTask(*coord)
			return err
		}
		self.tasks.DoneTask(*coord)
		self.results <- perflog.PerfLogEntry{
			Coord:      *coord,
			Timestamp:  time.Now(),
			RenderTime: time.Duration(resp.RenderTime),
			SaverTime:  time.Duration(resp.SaveTime),
		}
	}
	return nil
}

func (self *coordinator) connLoop(addr string) {
	// Send 'done' for _current_ goroutine
	defer self.connsWg.Done()

	// Process queue
	for {
		err := self.connSub(addr)
		if err == nil {
			return
		}
		if _, ok := err.(*gopnikrpc.QueueLimitExceeded); ok {
			return
		}
		log.Error("Slave connection error: %v", err)
		time.Sleep(10 * time.Second)
	}
}

func (self *coordinator) Start() <-chan perflog.PerfLogEntry {
	for _, addr := range self.addrs {
		for i := 0; i < self.nodeQueueSize; i++ {
			self.connsWg.Add(1)
			go self.connLoop(addr)
		}
	}
	go self.wait()
	return self.results
}

func (self *coordinator) wait() {
	self.connsWg.Wait()
	close(self.results)
}

func (self *coordinator) Nodes() []string {
	return self.addrs
}

func (self *coordinator) DoneTasks() (done int, total int) {
	done = self.tasks.DoneTasks()
	total = self.tasks.TotalTasks()
	return
}
