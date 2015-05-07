package tileserver

import (
	"testing"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	json "github.com/orofarne/strict-json"
	"github.com/stretchr/testify/require"

	"app"
	"gopnik"
	"gopnikrpc"
	"gopnikrpc/types"
	"plugins"
	_ "plugins_enabled"
	"sampledata"
)

type Config struct {
	CachePlugin string                     //
	Plugins     map[string]json.RawMessage //
}

func TestSimple(t *testing.T) {
	addr := "127.0.0.1:5342"

	cfg := []byte(`{
		"UseMultilevel": true,
		"Backend": {
			"Plugin":       "MemoryKV",
			"PluginConfig": {}
		}
	}`)
	renderPoolsConfig := app.RenderPoolsConfig{
		[]app.RenderPoolConfig{
			app.RenderPoolConfig{
				Cmd:         sampledata.SlaveCmd, // Render slave binary
				MinZoom:     0,
				MaxZoom:     19,
				PoolSize:    1,
				HPQueueSize: 10,
				LPQueueSize: 10,
				RenderTTL:   0,
			},
		},
	}

	cpI, err := plugins.DefaultPluginStore.Create("KVStorePlugin", json.RawMessage(cfg))
	if err != nil {
		t.Fatal(err)
	}
	cp, ok := cpI.(gopnik.CachePluginInterface)
	if !ok {
		t.Fatal("Invalid cache plugin type")
	}

	ts, err := NewTileServer(renderPoolsConfig, cp, time.Duration(0))
	if err != nil {
		t.Fatalf("Failed to create tile server: %v", err)
	}
	go func() {
		err := RunServer(addr, ts)
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Millisecond)

	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	socket, err := thrift.NewTSocket(addr)
	require.Nil(t, err)
	transport := transportFactory.GetTransport(socket)
	defer transport.Close()
	err = transport.Open()
	if err != nil {
		t.Errorf("transport open: %v", err.Error())
	}

	renderClient := gopnikrpc.NewRenderClientFactory(transport, protocolFactory)
	tile, err := renderClient.Render(&types.Coord{
		Zoom: 1,
		X:    0,
		Y:    0,
		Size: 1,
	},
		gopnikrpc.Priority_HIGH, false)

	require.Nil(t, err)
	require.NotNil(t, tile)
	require.NotNil(t, tile.Image)

	sampledata.CheckTile(t, tile.Image, "1_0_0.png")
}

func TestErrorRaising(t *testing.T) {
	addr := "127.0.0.1:5347"

	cfg := []byte(`{
		"SetError": "MyTestError"
	}`)
	renderPoolsConfig := app.RenderPoolsConfig{
		[]app.RenderPoolConfig{
			app.RenderPoolConfig{
				Cmd:         sampledata.SlaveCmd, // Render slave binary
				MinZoom:     0,
				MaxZoom:     19,
				PoolSize:    1,
				HPQueueSize: 10,
				LPQueueSize: 10,
				RenderTTL:   0,
			},
		},
	}

	cpI, err := plugins.DefaultPluginStore.Create("FakeCachePlugin", json.RawMessage(cfg))
	if err != nil {
		t.Fatal(err)
	}
	cp, ok := cpI.(gopnik.CachePluginInterface)
	if !ok {
		t.Fatal("Invalid cache plugin type")
	}

	ts, err := NewTileServer(renderPoolsConfig, cp, time.Duration(0))
	if err != nil {
		t.Fatalf("Failed to create tile server: %v", err)
	}
	go func() {
		err := RunServer(addr, ts)
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Millisecond)

	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	socket, err := thrift.NewTSocket(addr)
	require.Nil(t, err)
	transport := transportFactory.GetTransport(socket)
	defer transport.Close()
	err = transport.Open()
	if err != nil {
		t.Errorf("transport open: %v", err.Error())
	}

	renderClient := gopnikrpc.NewRenderClientFactory(transport, protocolFactory)
	_, err = renderClient.Render(&types.Coord{
		Zoom: 1,
		X:    0,
		Y:    0,
		Size: 1,
	},
		gopnikrpc.Priority_HIGH, true)

	require.NotNil(t, err)
	err2, ok := err.(*gopnikrpc.RenderError)
	require.True(t, ok)
	require.Equal(t, "RenderError({Message:MyTestError})", err2.Error())
}
