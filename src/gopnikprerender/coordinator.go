package main

import (
	"fmt"
	"sync"
	"time"

	"gopnik"
	"gopnikrpc"
	"perflog"
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

func (self *coordinator) connF(addr string) error {
	conn := newConnection(addr)
	defer conn.Close()
	err := conn.Connect()
	if err != nil {
		return err
	}

	for {
		coord := self.tasks.GetTask()
		if coord == nil {
			return &stop{}
		}

		res, err := conn.ProcessTask(*coord)
		if err != nil || res == nil {
			self.tasks.FailTask(*coord)
			return fmt.Errorf("%v on %v", err, coord)
		}
		self.tasks.DoneTask(*coord)
		self.results <- *res
	}
	return nil
}

func (self *coordinator) monitorConnF(addr string, t time.Time) error {
	conn := newConnection(addr)
	err := conn.Connect()
	defer conn.Close()
	if err != nil {
		return err
	}

	client := conn.Client()
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

func (self *coordinator) connLoop(addr string) {
	// Send 'done' for _current_ goroutine
	defer self.connsWg.Done()

	// Process tasks
	for {
		err := self.connF(addr)
		if _, ok := err.(*stop); ok {
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
	if err := self.monitorConnF(addr, time.Now()); err != nil {
		log.Error("Connection [%v] error: %v", addr, err)
	}

	ticker := time.Tick(1 * time.Minute)
	for now := range ticker {
		if err := self.monitorConnF(addr, now); err != nil {
			log.Error("Connection [%v] error: %v", addr, err)
		}
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
