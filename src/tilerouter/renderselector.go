package tilerouter

import (
	"fmt"
	"gopnikrpc"
	"hash/adler32"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

	"gopnik"
	"servicestatus"
)

const (
	Offline = iota
	Online
)

type renderPoint struct {
	Addr   string
	Status int
}

type RenderSelector struct {
	renders []renderPoint
	timeout time.Duration
	closed  chan struct{}
}

func NewRenderSelector(renders []string, pingPeriod time.Duration, timeout time.Duration) (*RenderSelector, error) {
	rs := new(RenderSelector)
	rs.renders = make([]renderPoint, len(renders))
	for i, addr := range renders {
		rs.renders[i].Addr = addr
		rs.renders[i].Status = Offline
	}
	rs.timeout = timeout
	rs.pingAll()
	rs.updateServiceStatus()
	rs.closed = make(chan struct{})
	go func() {
		period := pingPeriod
		for {
			select {
			case <-rs.closed:
				return
			case t1 := <-time.After(period):
				rs.pingAll()
				rs.updateServiceStatus()

				Δt := time.Since(t1)
				if Δt >= pingPeriod {
					period = 0
				} else {
					period = pingPeriod - Δt
				}
			}
		}
	}()
	return rs, nil
}

func (rs *RenderSelector) hash(str string) uint32 {
	return adler32.Checksum([]byte(str))
}

func (rs *RenderSelector) statusToString(status int) string {
	switch status {
	case Offline:
		return "Offline"
	case Online:
		return "Online"
	default:
		return "<unknown>"
	}
	panic("?!")
}

func (rs *RenderSelector) pingAll() {
	var wg sync.WaitGroup
	for i := 0; i < len(rs.renders); i++ {
		wg.Add(1)
		go func(i int) {
			oldStatus := rs.renders[i].Status
			rs.renders[i].Status = rs.ping(i)

			log.Debug("'%v' is %v", rs.renders[i].Addr, rs.statusToString(rs.renders[i].Status))
			if rs.renders[i].Status != oldStatus {
				log.Info("New status for '%v': %v", rs.renders[i].Addr, rs.statusToString(rs.renders[i].Status))
			}

			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (rs *RenderSelector) updateServiceStatus() {
	for _, render := range rs.renders {
		if render.Status == Online {
			servicestatus.SetOK()
			return
		}
	}
	servicestatus.SetFAIL()
}

func (self *RenderSelector) ping(i int) int {
	addr := self.renders[i].Addr

	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	socket, err := thrift.NewTSocketTimeout(addr, self.timeout)
	if err != nil {
		return Offline
	}
	transport := transportFactory.GetTransport(socket)
	defer transport.Close()
	err = transport.Open()
	if err != nil {
		return Offline
	}

	renderClient := gopnikrpc.NewRenderClientFactory(transport, protocolFactory)
	status, err := renderClient.Status()

	if err != nil || !status {
		return Offline
	}

	return Online
}

func (rs *RenderSelector) SetStatus(addr string, status int) {
	for i := 0; i < len(rs.renders); i++ {
		if rs.renders[i].Addr == addr {
			rs.renders[i].Status = status
			log.Info("New status for '%v': %v", addr, rs.statusToString(status))
			return
		}
	}
}

func (rs *RenderSelector) aliveRenders() (aRenders []int) {
	for i := 0; i < len(rs.renders); i++ {
		if rs.renders[i].Status == Online {
			aRenders = append(aRenders, i)
		}
	}
	return
}

func (rs *RenderSelector) SelectRender(coord gopnik.TileCoord) string {
	aRenders := rs.aliveRenders()
	if len(aRenders) == 0 {
		return ""
	}
	coordHash := rs.hash(fmt.Sprintf("%v/%v/%v", coord.Zoom, coord.X, coord.Y))
	renderId := aRenders[int(coordHash)%len(aRenders)]
	return rs.renders[renderId].Addr
}

func (rs *RenderSelector) Stop() {
	close(rs.closed)
}
