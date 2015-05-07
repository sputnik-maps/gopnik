package tileserver

import (
	"gopnik"
	"time"

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

func (self *thriftTileServer) Render(coord *types.Coord, prio gopnikrpc.Priority, wait_storage bool) (r *gopnikrpc.RenderResponse, err error) {
	tc := gopnikrpcutils.CoordFromRPC(coord)
	r = gopnikrpc.NewRenderResponse()

	var tiles []gopnik.Tile
	var renderTime, saveTime time.Duration
	tiles, renderTime, saveTime, err = self.tileServer.ServeTileRequest(tc, prio, wait_storage)
	if err != nil {
		return
	}

	// Times
	r.RenderTime = renderTime.Nanoseconds()
	r.SaveTime = saveTime.Nanoseconds()

	// Tiles
	for i := 0; i < len(tiles); i++ {
		r.Tiles = append(r.Tiles, gopnikrpcutils.TileToRPC(&tiles[i]))
	}

	return
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
