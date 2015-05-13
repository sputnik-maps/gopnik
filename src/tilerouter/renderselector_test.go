package tilerouter

import (
	"gopnikrpc"
	"gopnikrpc/types"
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/require"

	"gopnik"
)

type fakeRender struct {
}

func (self *fakeRender) Status() (r bool, err error)    { return true, nil }
func (self *fakeRender) Version() (r string, err error) { return "?", nil }
func (self *fakeRender) Config() (r string, err error)  { return "?", nil }
func (self *fakeRender) Stat() (r map[string]float64, err error) {
	return map[string]float64{}, nil
}
func (self *fakeRender) Render(coord *types.Coord, prio gopnikrpc.Priority, wait_storage bool) (r *gopnikrpc.RenderResponse, err error) {
	return nil, nil
}

func runFakeRender(addr string) {
	var tTS fakeRender
	processor := gopnikrpc.NewRenderProcessor(&tTS)

	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		panic(err)
	}
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	if err = server.Serve(); err != nil {
		panic(err)
	}
}

func TestRoute(t *testing.T) {
	renders := []string{"localhost:9001", "localhost:9002", "localhost:9003"}

	for _, r := range renders {
		go runFakeRender(r)
	}
	time.Sleep(100 * time.Millisecond)

	rs, err := NewRenderSelector(renders, time.Second, 30*time.Second)
	require.Nil(t, err)
	defer rs.Stop()

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	back1, err1 := rs.SelectRender(coord)
	require.Nil(t, err1)
	coord.X = 3
	back2, err2 := rs.SelectRender(coord)
	require.Nil(t, err2)
	coord.Y = 4
	back3, err3 := rs.SelectRender(coord)
	require.Nil(t, err3)
	coord.Zoom = 5
	back4, err4 := rs.SelectRender(coord)
	require.Nil(t, err4)

	require.True(t,
		back1.Addr != back2.Addr || back1.Addr != back3.Addr || back1.Addr != back4.Addr,
	)
}
