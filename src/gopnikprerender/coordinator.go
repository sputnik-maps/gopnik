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
	addrs   []string
	connsWg sync.WaitGroup
	tasks   *plan
	results chan perflog.PerfLogEntry
}

func newCoordinator(addrs []string, bboxes []gopnik.TileCoord) *coordinator {
	p := new(coordinator)
	p.addrs = addrs
	p.results = make(chan perflog.PerfLogEntry)

	p.tasks = newPlan(bboxes)

	return p
}

func (p *coordinator) connSub(addr string) error {
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
		coord := p.tasks.GetTask()
		if coord == nil {
			break
		}

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
	// Send 'done' for _current_ goroutine
	defer p.connsWg.Done()

	// Start another connection
	// TODO !!!!!
	// p.connsWg.Add(1)
	// go func() {
	// 	time.Sleep(100 * time.Millisecond)
	// 	p.connLoop(addr)
	// }()

	// Process queue
	for {
		err := p.connSub(addr)
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

func (p *coordinator) DoneTasks() (done int, total int) {
	done = p.tasks.DoneTasks()
	total = p.tasks.TotalTasks()
	return
}
