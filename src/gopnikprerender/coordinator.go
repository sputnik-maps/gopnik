package main

import (
	"fmt"
	"sync"
	"time"

	"gopnik"
	"gopnikprerenderlib"
)

type coordinator struct {
	addrs     []string
	conns     map[string]*slaveConn
	connsMu   sync.RWMutex
	connsWg   sync.WaitGroup
	tasks     *plan
	resultsEx chan SlaveResponse
}

func newCoordinator(addrs []string, bboxes []gopnik.TileCoord) *coordinator {
	p := new(coordinator)
	p.addrs = addrs
	p.resultsEx = make(chan SlaveResponse)
	p.conns = make(map[string]*slaveConn)

	p.tasks = newPlan(bboxes)

	return p
}

func (p *coordinator) connSub(addr string) {
	defer p.connsWg.Done()

	for {
		conn := newSlaveConn(p.tasks, p.resultsEx)
		err := conn.Connect(addr)
		if err != nil {
			log.Error("Connect error: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		p.connsMu.Lock()
		p.conns[addr] = conn
		p.connsMu.Unlock()

		err = conn.Run()
		if err != nil {
			log.Error("Slave connection error: %v", err)
		} else {
			log.Debug("Nothing to do, stopping connection to %v", addr)
			return
		}

		p.connsMu.Lock()
		delete(p.conns, addr)
		p.connsMu.Unlock()

		time.Sleep(10 * time.Second)
	}
}

func (p *coordinator) Start() <-chan SlaveResponse {
	p.connsWg.Add(len(p.addrs))
	for _, addr := range p.addrs {
		go p.connSub(addr)
	}
	return p.resultsEx
}

func (p *coordinator) Wait() {
	p.connsWg.Wait()
}

func (p *coordinator) Nodes() []string {
	return p.addrs
}

func (p *coordinator) NodeConfig(addr string) (*gopnikprerenderlib.RConfig, error) {
	p.connsMu.RLock()
	defer p.connsMu.RUnlock()

	conn, found := p.conns[addr]
	if !found {
		return nil, fmt.Errorf("Connection to '%v' is unavailable", addr)
	}
	cfg := conn.Configuration()
	return &cfg, nil
}

func (p *coordinator) SetNodeConfig(addr string, cfg gopnikprerenderlib.RConfig) error {
	p.connsMu.Lock()
	defer p.connsMu.Unlock()

	conn, found := p.conns[addr]
	if !found {
		return fmt.Errorf("Connection to '%v' is unavailable", addr)
	}
	conn.Reconfigure(cfg)
	return nil
}

func (p *coordinator) DoneTasks() (done int, total int) {
	done = p.tasks.DoneTasks()
	total = p.tasks.TotalTasks()
	return
}

func (p *coordinator) NodeMonitor(addr string) (*monitor, error) {
	p.connsMu.RLock()
	defer p.connsMu.RUnlock()

	conn, found := p.conns[addr]
	if !found {
		return nil, fmt.Errorf("Connection to '%v' is unavailable", addr)
	}
	return conn.Monitor(), nil
}
