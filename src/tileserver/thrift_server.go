package tileserver

import (
	"git.apache.org/thrift.git/lib/go/thrift"

	"gopnikrpc"
	"gopnikrpc/types"
)

type thriftTileServer struct {
	tileServer *TileServer
}

func (self *thriftTileServer) Render(coord *types.Coord, prio gopnikrpc.Priority, wait_storage bool) (r *types.Tile, err error) {
	return nil, nil
}

func (self *thriftTileServer) Status() (r bool, err error)    { return true, nil }
func (self *thriftTileServer) Version() (r string, err error) { return "?", nil }
func (self *thriftTileServer) Config() (r string, err error)  { return "?", nil }
func (self *thriftTileServer) Stat() (r map[string]float64, err error) {
	return map[string]float64{}, nil
}

func RunServer(addr string, tileServer *TileServer) error {
	tTS := &thriftTileServer{
		tileServer: tileServer,
	}
	processor := gopnikrpc.NewRenderProcessor(tTS)

	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		return err
	}
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	return server.Serve()
}
