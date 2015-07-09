package tilerender

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gopnik"
	"gopnikrpc"
	"sampledata"
)

var executionTimeout time.Duration =  60 * time.Second

func TestOneRender(t *testing.T) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 1, 1, 1, 0, executionTimeout)
	require.Nil(t, err)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	ansCh := make(chan *RenderPoolResponse)
	err = rpool.EnqueueRequest(coord, ansCh, gopnikrpc.Priority_HIGH)
	require.Nil(t, err)
	ans := <-ansCh
	require.Nil(t, ans.Error)
	require.Equal(t, len(ans.Tiles), 1)
	sampledata.CheckTile(t, ans.Tiles[0].Image, "1_0_0.png")
}

func Test5RendersHP(t *testing.T) {
	const nTiles = 15

	rpool, err := NewRenderPool(sampledata.SlaveCmd, 5, nTiles, 0, 0, executionTimeout)
	require.Nil(t, err)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	ansCh := make(chan *RenderPoolResponse)
	for i := 0; i < nTiles; i++ {
		err = rpool.EnqueueRequest(coord, ansCh, gopnikrpc.Priority_HIGH)
		require.Nil(t, err)
	}
	for i := 0; i < nTiles; i++ {
		ans := <-ansCh
		require.Nil(t, ans.Error)
		require.Equal(t, len(ans.Tiles), 1)
		sampledata.CheckTile(t, ans.Tiles[0].Image, "1_0_0.png")
	}
}

func Test5RendersLP(t *testing.T) {
	const nTiles = 15

	rpool, err := NewRenderPool(sampledata.SlaveCmd, 5, 0, nTiles, 0, executionTimeout)
	require.Nil(t, err)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	ansCh := make(chan *RenderPoolResponse)
	for i := 0; i < nTiles; i++ {
		err = rpool.EnqueueRequest(coord, ansCh, gopnikrpc.Priority_LOW)
		require.Nil(t, err)
	}
	for i := 0; i < nTiles; i++ {
		ans := <-ansCh
		require.Nil(t, ans.Error)
		require.Equal(t, len(ans.Tiles), 1)
		sampledata.CheckTile(t, ans.Tiles[0].Image, "1_0_0.png")
	}
}

func TestTTL(t *testing.T) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 1, 4, 0, 2, executionTimeout)
	require.Nil(t, err)

	ansCh := make(chan *RenderPoolResponse)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			coord := gopnik.TileCoord{
				X:    uint64(i),
				Y:    uint64(j),
				Zoom: 1,
				Size: 1,
			}
			err = rpool.EnqueueRequest(coord, ansCh, gopnikrpc.Priority_HIGH)
			require.Nil(t, err)
		}
	}
	for i := 0; i < 4; i++ {
		ans := <-ansCh
		require.Nil(t, ans.Error)
		require.Equal(t, len(ans.Tiles), 1)
	}
}

func TestOneRender4Tiles(t *testing.T) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 1, 1, 0, 0, executionTimeout)
	require.Nil(t, err)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 2,
	}
	ansCh := make(chan *RenderPoolResponse)
	err = rpool.EnqueueRequest(coord, ansCh, gopnikrpc.Priority_HIGH)
	require.Nil(t, err)
	ans := <-ansCh
	require.Nil(t, ans.Error)
	require.Equal(t, len(ans.Tiles), 4)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			sampledata.CheckTile(t, ans.Tiles[i*2+j].Image,
				fmt.Sprintf("1_%d_%d.png", j, i))
		}
	}
}

func Benchmark5Renders(b *testing.B) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 5, uint(b.N), 0, 0, executionTimeout)
	if err != nil {
		b.Errorf("NewRenderPool error: %v", err)
	}
	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	ansCh := make(chan *RenderPoolResponse)
	for i := 0; i < b.N; i++ {
		err = rpool.EnqueueRequest(coord, ansCh, gopnikrpc.Priority_HIGH)
		if err != nil {
			b.Errorf("EnqueueRequest error: %v", err)
		}
	}
	for i := 0; i < b.N; i++ {
		ans := <-ansCh
		if ans.Error != nil {
			b.Errorf("Got error: %v", ans.Error)
		}
		if len(ans.Tiles) != 1 {
			b.Errorf("Tiles len = %v", len(ans.Tiles))
		}
	}
}
