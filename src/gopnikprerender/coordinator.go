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
	addrs          []string
	connsWg        sync.WaitGroup
	tasks          *plan
	results        chan perflog.PerfLogEntry
	nodeQueueSize  int
	connStatuses   map[string]bool
	connStatusesMu sync.Mutex
	monitors       map[string]*monitor
}

type stop struct{}

func (self *stop) Error() string { return "Stop" }

func newCoordinator(addrs []string, nodeQueueSize int, bboxes []gopnik.TileCoord) *coordinator {
	self := new(coordinator)
	self.addrs = addrs
	self.results = make(chan perflog.PerfLogEntry)
	self.nodeQueueSize = nodeQueueSize

	self.tasks = newPlan(bboxes)

	// Monitoring
	self.connStatuses = make(map[string]bool)
	self.monitors = make(map[string]*monitor)
	for _, addr := range self.addrs {
		self.monitors[addr] = newMonitor()
	}
	for _, addr := range self.addrs {
		go self.monitorLoop(addr)
	}

	return self
}

func (self *coordinator) connSub(addr string, f func(addr string, client *gopnikrpc.RenderClient) error) error {
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
		err := f(addr, renderClient)
		if err != nil {
			if _, ok := err.(*stop); ok {
				break
			} else {
				return err
			}
		}
	}
	return nil
}

func (self *coordinator) taskConnF(addr string, client *gopnikrpc.RenderClient) error {
	coord := self.tasks.GetTask()
	if coord == nil {
		return &stop{}
	}

	resp, err := client.Render(gopnikrpcutils.CoordToRPC(coord), gopnikrpc.Priority_LOW, true)
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
	return nil
}

func (self *coordinator) monitorConnF(addr string, client *gopnikrpc.RenderClient) error {
	worker := func(t time.Time) error {
		data, err := client.Stat()
		if err != nil {
			self.setNodeStatus(addr, false)
			return err
		}
		self.setNodeStatus(addr, true)

		mon := self.NodeMonitor(addr)
		for k, v := range data {
			mon.AddPoint(k, t, v)
		}
		return nil
	}

	if err := worker(time.Now()); err != nil {
		return err
	}

	ticker := time.Tick(1 * time.Minute)
	for now := range ticker {
		if err := worker(now); err != nil {
			return err
		}
	}
	return nil
}

func (self *coordinator) connLoop(addr string) {
	// Send 'done' for _current_ goroutine
	defer self.connsWg.Done()

	// Process queue
	for {
		err := self.connSub(addr, self.taskConnF)
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

func (self *coordinator) monitorLoop(addr string) {
	for {
		err := self.connSub(addr, self.monitorConnF)
		if err == nil {
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

func (self *coordinator) setNodeStatus(node string, status bool) {
	self.connStatusesMu.Lock()
	self.connStatuses[node] = status
	self.connStatusesMu.Unlock()
}

func (self *coordinator) Nodes() []string {
	return self.addrs
}

func (self *coordinator) NodeStatus(node string) bool {
	self.connStatusesMu.Lock()
	defer self.connStatusesMu.Unlock()
	return self.connStatuses[node]
}

func (self *coordinator) NodeMonitor(node string) *monitor {
	return self.monitors[node]
}

func (self *coordinator) DoneTasks() (done int, total int) {
	done = self.tasks.DoneTasks()
	total = self.tasks.TotalTasks()
	return
}
