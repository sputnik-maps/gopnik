package tilerouter

import (
	"container/list"
	"fmt"
	"gopnikrpc"
	"hash/adler32"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

	"app"
	"gopnik"
	"servicestatus"
)

const (
	Offline = iota
	Online
)

type thriftConn struct {
	Addr      string
	Socket    *thrift.TSocket
	Transport thrift.TTransport
	Client    *gopnikrpc.RenderClient
}

func (self *thriftConn) Close() {
	self.Transport.Close()
}

type renderPoint struct {
	Addr         string
	Status       int
	Connections  *list.List
	MinFreeConns int
	Mu           sync.Mutex
}

type RenderSelector struct {
	renders          []renderPoint
	timeout          time.Duration
	closed           chan struct{}
	transportFactory thrift.TTransportFactory
	protocolFactory  thrift.TProtocolFactory
}

func NewRenderSelector(renders []string, pingPeriod time.Duration, timeout time.Duration) (*RenderSelector, error) {
	self := new(RenderSelector)
	self.renders = make([]renderPoint, len(renders))
	for i, addr := range renders {
		self.renders[i].Addr = addr
		self.renders[i].Status = Offline
		self.renders[i].Connections = list.New()
	}
	self.timeout = timeout
	self.transportFactory = thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	self.protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()

	self.pingAll()
	self.updateServiceStatus()
	self.closed = make(chan struct{})
	go func() {
		period := pingPeriod
		for {
			select {
			case <-self.closed:
				self.closed <- struct{}{}
				return
			case t1 := <-time.After(period):
				self.pingAll()
				self.updateServiceStatus()
				self.cleanConnections()

				Δt := time.Since(t1)
				if Δt >= pingPeriod {
					period = 0
				} else {
					period = pingPeriod - Δt
				}
			}
		}
	}()
	return self, nil
}

func (self *RenderSelector) hash(str string) uint32 {
	return adler32.Checksum([]byte(str))
}

func (self *RenderSelector) statusToString(status int) string {
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

func (self *RenderSelector) pingAll() {
	var wg sync.WaitGroup
	for i := 0; i < len(self.renders); i++ {
		wg.Add(1)
		go func(i int) {
			oldStatus := self.renders[i].Status
			self.renders[i].Status = self.ping(i)

			log.Debug("'%v' is %v", self.renders[i].Addr, self.statusToString(self.renders[i].Status))
			if self.renders[i].Status != oldStatus {
				log.Info("New status for '%v': %v", self.renders[i].Addr, self.statusToString(self.renders[i].Status))
			}

			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (self *RenderSelector) cleanConnections() {
	for i := 0; i < len(self.renders); i++ {
		self.cleanConnection(&self.renders[i])
	}
}

func (self *RenderSelector) updateServiceStatus() {
	for _, render := range self.renders {
		if render.Status == Online {
			servicestatus.SetOK()
			return
		}
	}
	servicestatus.SetFAIL()
}

func (self *RenderSelector) connect(addr string) (*thriftConn, error) {
	var err error
	conn := &thriftConn{
		Addr: addr,
	}

	conn.Socket, err = thrift.NewTSocketTimeout(addr, self.timeout)
	if err != nil {
		return nil, err
	}

	conn.Transport = self.transportFactory.GetTransport(conn.Socket)
	err = conn.Transport.Open()
	if err != nil {
		return nil, err
	}

	conn.Client = gopnikrpc.NewRenderClientFactory(conn.Transport, self.protocolFactory)

	return conn, nil
}

func (self *RenderSelector) checkConnectionCache(rp *renderPoint) (conn *thriftConn) {
	rp.Mu.Lock()
	defer rp.Mu.Unlock()

	if rp.Connections.Len() > 0 {
		elem := rp.Connections.Front()
		conn = elem.Value.(*thriftConn)
		rp.Connections.Remove(elem)
		connLen := rp.Connections.Len()
		if connLen < rp.MinFreeConns {
			rp.MinFreeConns = connLen
		}
		return conn
	}

	return nil
}

func (self *RenderSelector) getConnection(rp *renderPoint) (conn *thriftConn, err error) {
	conn = self.checkConnectionCache(rp)
	if conn != nil {
		return
	}

	return self.connect(rp.Addr)
}

func (self *RenderSelector) cleanConnection(rp *renderPoint) {
	rp.Mu.Lock()
	defer rp.Mu.Unlock()

	connLen := rp.Connections.Len()
	if connLen > rp.MinFreeConns {
		N := connLen - rp.MinFreeConns
		if N > connLen/2 {
			N = connLen / 2
		}
		for i := 0; i < N; i++ {
			rp.Connections.Remove(rp.Connections.Front())
		}
	}
}

func (self *RenderSelector) ping(i int) int {
	conn, err := self.getConnection(&self.renders[i])
	if err != nil {
		return Offline
	}
	defer self.FreeConnection(conn)

	status, err := conn.Client.Status()

	if err != nil || !status {
		return Offline
	}

	return Online
}

func (self *RenderSelector) SetStatus(addr string, status int) {
	for i := 0; i < len(self.renders); i++ {
		if self.renders[i].Addr == addr {
			self.renders[i].Status = status
			log.Info("New status for '%v': %v", addr, self.statusToString(status))
			return
		}
	}
}

func (self *RenderSelector) aliveRenders() (aRenders []int) {
	for i := 0; i < len(self.renders); i++ {
		if self.renders[i].Status == Online {
			aRenders = append(aRenders, i)
		}
	}
	return
}

func (self *RenderSelector) SelectRender(coord gopnik.TileCoord) (*thriftConn, error) {
	aRenders := self.aliveRenders()
	if len(aRenders) == 0 {
		return nil, nil
	}
	metacoord := app.App.Metatiler().TileToMetatile(&coord)
	coordHash := self.hash(fmt.Sprintf("%v/%v/%v", metacoord.Zoom, metacoord.X, metacoord.Y))
	renderId := aRenders[int(coordHash)%len(aRenders)]
	return self.getConnection(&self.renders[renderId])
}

func (self *RenderSelector) FreeConnection(conn *thriftConn) {
	for _, rp := range self.renders {
		if rp.Addr == conn.Addr {
			rp.Mu.Lock()
			rp.Connections.PushBack(conn)
			rp.Mu.Unlock()
			return
		}
	}
}

func (self *RenderSelector) Stop() {
	self.closed <- struct{}{}
	<-self.closed
	for _, r := range self.renders {
		for e := r.Connections.Front(); e != nil; e = e.Next() {
			e.Value.(*thriftConn).Close()
		}
	}
}
