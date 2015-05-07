package tileserver

import (
	"git.apache.org/thrift.git/lib/go/thrift"

	"gopnikrpc"
	"gopnikrpc/types"
	"gopnikrpcutils"
	"rpcbaseservice"
)

type thriftTileServer struct {
	*rpcbaseservice.Service
	tileServer *TileServer
}

func (self *thriftTileServer) Render(coord *types.Coord, prio gopnikrpc.Priority, wait_storage bool) (r *types.Tile, err error) {
	tc := gopnikrpcutils.CoordFromRPC(coord)

	tile, err := self.tileServer.ServeTileRequest(tc, prio, wait_storage)
	if err != nil {
		return nil, &gopnikrpc.RenderError{Message: err.Error()}
	}

	return gopnikrpcutils.TileToRPC(tile), nil
}

func RunServer(addr string, tileServer *TileServer) error {
	tTS := &thriftTileServer{
		tileServer: tileServer,
		Service:    rpcbaseservice.NewService(),
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
