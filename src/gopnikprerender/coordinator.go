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
	"gopnikprerenderlib"
)

type coordinator struct {
	addrs   []string
	conns   map[string]*slaveConn
	connsMu sync.RWMutex
	connsWg sync.WaitGroup
	tasks   *plan
	results chan perflog.PerfLogEntry
}

func newCoordinator(addrs []string, bboxes []gopnik.TileCoord) *coordinator {
	p := new(coordinator)
	p.addrs = addrs
	p.results = make(chan perflog.PerfLogEntry)
	p.conns = make(map[string]*slaveConn)

	p.tasks = newPlan(bboxes)

	return p
}

func (p *coordinator) connSub(addr string) error {
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
		coord := p.tasks.GetTask()
		resp, err := renderClient.Render(gopnikrpcutils.CoordToRPC(coord), gopnikrpc.Priority_LOW, true)
		if err != nil {
			p.tasks.FailTask(*coord)
			return err
		}
		p.tasks.DoneTask(*coord)
		p.results <- perflog.PerfLogEntry{
			Coord:      *coord,
			Timestamp:  time.Now(),
			RenderTime: time.Duration(resp.RenderTime),
			SaverTime:  time.Duration(resp.SaveTime),
		}
	}
	return nil
}

func (p *coordinator) connLoop(addr string) {
	defer p.connsWg.Done()

	for {
		err := p.connSub(addr)
		if err == nil {
			return
		}
		log.Error("Slave connection error: %v", err)
		time.Sleep(10 * time.Second)

		// conn := newSlaveConn(p.tasks, p.resultsEx)
		// err := conn.Connect(addr)
		// if err != nil {
		// 	log.Error("Connect error: %v", err)
		// 	time.Sleep(10 * time.Second)
		// 	continue
		// }
		//
		// p.connsMu.Lock()
		// p.conns[addr] = conn
		// p.connsMu.Unlock()
		//
		// err = conn.Run()
		// if err != nil {
		// 	log.Error("Slave connection error: %v", err)
		// } else {
		// 	log.Debug("Nothing to do, stopping connection to %v", addr)
		// 	return
		// }
		//
		// p.connsMu.Lock()
		// delete(p.conns, addr)
		// p.connsMu.Unlock()
		//
		// time.Sleep(10 * time.Second)
	}
}

func (p *coordinator) Start() <-chan perflog.PerfLogEntry {
	p.connsWg.Add(len(p.addrs))
	for _, addr := range p.addrs {
		go p.connLoop(addr)
	}
	return p.results
}

func (p *coordinator) Wait() {
	p.connsWg.Wait()
}

func (p *coordinator) Nodes() []string {
	return p.addrs
}

func (p *coordinator) NodeConfig(addr string) (*gopnikprerenderlib.RConfig, error) {
	// p.connsMu.RLock()
	// defer p.connsMu.RUnlock()
	//
	// conn, found := p.conns[addr]
	// if !found {
	// 	return nil, fmt.Errorf("Connection to '%v' is unavailable", addr)
	// }
	// cfg := conn.Configuration()
	// return &cfg, nil
	return nil, fmt.Errorf("!!!!!")
}

func (p *coordinator) SetNodeConfig(addr string, cfg gopnikprerenderlib.RConfig) error {
	// p.connsMu.Lock()
	// defer p.connsMu.Unlock()
	//
	// conn, found := p.conns[addr]
	// if !found {
	// 	return fmt.Errorf("Connection to '%v' is unavailable", addr)
	// }
	// conn.Reconfigure(cfg)
	return nil
}

func (p *coordinator) DoneTasks() (done int, total int) {
	done = p.tasks.DoneTasks()
	total = p.tasks.TotalTasks()
	return
}

func (p *coordinator) NodeMonitor(addr string) (*monitor, error) {
	// p.connsMu.RLock()
	// defer p.connsMu.RUnlock()
	//
	// conn, found := p.conns[addr]
	// if !found {
	// 	return nil, fmt.Errorf("Connection to '%v' is unavailable", addr)
	// }
	// return conn.Monitor(), nil
	return nil, nil
}
