package main

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"

	"app"
	"gopnik"
	"gopnikprerenderlib"

	"github.com/davecgh/go-spew/spew"
	json "github.com/orofarne/strict-json"
)

type SlaveResponse struct {
	Addr string
	gopnikprerenderlib.RResponse
}

type slaveConn struct {
	app.RenderPoolsConfig
	conn         net.Conn
	addr         string
	saversPool   int
	dec          *gob.Decoder
	enc          *gob.Encoder
	writeQ       chan interface{}
	taskReqQ     chan int
	tasksGlobal  *plan
	inProgress   list.List
	inProgressMu sync.Mutex
	results      chan<- SlaveResponse
	stopFlag     bool
	stopFlagMu   sync.Mutex
	cfgLock      sync.RWMutex
	mon          *monitor
}

func newSlaveConn(tasks *plan, results chan<- SlaveResponse) *slaveConn {
	self := new(slaveConn)
	self.tasksGlobal = tasks
	self.results = results
	self.mon = newMonitor()
	return self
}

func (self *slaveConn) Addr() string {
	return self.addr
}

func (self *slaveConn) Connect(addr string) error {
	self.addr = addr
	log.Debug("Connecting to %v...", self.addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	self.conn = conn

	self.dec = gob.NewDecoder(self.conn)
	self.enc = gob.NewEncoder(self.conn)

	var resp gopnikprerenderlib.RResponse
	err = self.dec.Decode(&resp)
	if err != nil {
		return err
	}

	if resp.Type != gopnikprerenderlib.Hello {
		self.conn.Close()
		return fmt.Errorf("Invalid 'Hello' message (type = %v)", resp.Type)
	}
	if resp.Hello == nil {
		self.conn.Close()
		return fmt.Errorf("Empty 'Hello' message")
	}

	log.Debug("Hello from %v: %v", self.addr, spew.Sdump(resp.Hello))

	self.cfgLock.Lock()
	self.saversPool = resp.Hello.SaverPool
	self.RenderPools = resp.Hello.RenderPools
	// Detect slave queue size
	qSize := uint(0)
	for _, rCfg := range self.RenderPools {
		qSize += rCfg.PoolSize
	}
	qSize *= 100
	for _, rCfg := range self.RenderPools {
		if rCfg.QueueSize < qSize {
			qSize = rCfg.QueueSize
		}
	}
	qSize += 2
	self.cfgLock.Unlock()

	self.writeQ = make(chan interface{}, qSize)
	self.taskReqQ = make(chan int, qSize)

	return nil
}

// Return all undone tasks to queue
func (self *slaveConn) fail() {
	self.inProgressMu.Lock()
	defer self.inProgressMu.Unlock()

	log.Debug("Connection to %v: returning %v tasks to queue...", self.Addr(), self.inProgress.Len())

	for e := self.inProgress.Front(); e != nil; e = e.Next() {
		self.tasksGlobal.FailTask(*e.Value.(*gopnik.TileCoord))
	}
}

func (self *slaveConn) beginTask(coord *gopnik.TileCoord) {
	self.inProgressMu.Lock()
	self.inProgress.PushBack(coord)
	self.inProgressMu.Unlock()
}

func (self *slaveConn) endTask(coord *gopnik.TileCoord, done bool) {
	var err error
	if done {
		log.Debug("Task %v done", coord)
		err = self.tasksGlobal.DoneTask(*coord)
	} else {
		log.Debug("Task %v fail", coord)
		err = self.tasksGlobal.FailTask(*coord)
	}
	if err != nil {
		log.Error("Plan error: %v", err)
	}

	self.inProgressMu.Lock()
	defer self.inProgressMu.Unlock()

	// Remove complete task from list
	for e := self.inProgress.Front(); e != nil; e = e.Next() {
		if coord.Equals(e.Value.(*gopnik.TileCoord)) {
			self.inProgress.Remove(e)
			return
		}
	}
}

func (self *slaveConn) writer() {
	defer self.fail()
	for msg := range self.writeQ {
		err := self.enc.Encode(msg)
		if err != nil {
			log.Error("Encode error [%v]: %v", self.addr, err)
			return
		}
	}
}

// Writes tasks and commands to slave connection
func (self *slaveConn) feeder() {
L:
	for rN := range self.taskReqQ {
		for i := 0; i < rN; i++ {
			task := self.tasksGlobal.GetTask()
			if task != nil {
				self.stopFlagMu.Lock()
				sf := self.stopFlag
				self.stopFlagMu.Unlock()
				if sf {
					self.tasksGlobal.FailTask(*task)
					return
				}

				self.beginTask(task)
				self.writeQ <- &gopnikprerenderlib.RTask{
					Type:  gopnikprerenderlib.RenderTask,
					Coord: task,
				}
			} else {
				break L
			}
		}
	}

	self.stopFlagMu.Lock()
	self.stopFlag = true
	self.stopFlagMu.Unlock()
}

// Parse and save metrics from slave
func (self *slaveConn) procesMonitoring(data string) error {
	var decodedData map[string]float64

	err := json.Unmarshal([]byte(data), &decodedData)
	if err != nil {
		return err
	}
	now := time.Now()
	for k, v := range decodedData {
		self.mon.AddPoint(k, now, v)
	}
	return nil
}

// Run connection event loop
func (self *slaveConn) Run() error {
	go self.feeder()
	go self.writer()

	// Run read loop
	return self.readLoop()
}

// Read loop
func (self *slaveConn) readLoop() error {
	defer func() {
		self.stopFlagMu.Lock()
		self.stopFlag = true
		self.stopFlagMu.Unlock()

		close(self.taskReqQ)
		close(self.writeQ)

		log.Debug("Connection to %v stopped", self.addr)
	}()

	for {
		self.stopFlagMu.Lock()
		sf := self.stopFlag
		self.stopFlagMu.Unlock()
		if sf {
			return nil
		}

		// Parse response
		var resp SlaveResponse
		err := self.dec.Decode(&resp.RResponse)
		if err != nil {
			return err
		}
		log.Debug("Response from %v: %v %v", self.addr, resp, resp.Stat)
		if resp.Addr == "" {
			resp.Addr = self.Addr()
		}

		switch resp.Type {
		case gopnikprerenderlib.Stat:
			// Remove task from list
			self.endTask(resp.Coord, true)

			// Push response
			self.results <- resp
		case gopnikprerenderlib.Error:
			// Remove task from list
			if resp.Coord != nil {
				self.endTask(resp.Coord, false)
			}

			// Push response
			self.results <- resp
		case gopnikprerenderlib.GetTask:
			// Slave requests new tasks
			self.taskReqQ <- 1

		case gopnikprerenderlib.Monitoring:
			err = self.procesMonitoring(resp.Mon)
			if err != nil {
				log.Error("RPC: failed to decode monitoring message: %v", err)
			}
		default:
			log.Error("Unknown RPC message type: %v", resp.Type)
		}
	}
	panic("?!")
}

func (self *slaveConn) Reconfigure(cfg gopnikprerenderlib.RConfig) {
	self.stopFlagMu.Lock()
	sf := self.stopFlag
	self.stopFlagMu.Unlock()
	if sf {
		return
	}

	self.cfgLock.Lock()
	defer self.cfgLock.Unlock()

	self.saversPool = cfg.SaverThreads
	self.RenderPools = cfg.RenderPools

	self.writeQ <- gopnikprerenderlib.RTask{
		Type:   gopnikprerenderlib.Config,
		Config: &cfg,
	}
}

func (self *slaveConn) Configuration() (cfg gopnikprerenderlib.RConfig) {
	self.cfgLock.RLock()
	defer self.cfgLock.RUnlock()

	cfg.SaverThreads = self.saversPool
	cfg.RenderPools = self.RenderPools
	return
}

func (self *slaveConn) Monitor() (mon *monitor) {
	return self.mon
}
