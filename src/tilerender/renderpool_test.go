package tilerender

import (
	"fmt"
	"testing"

	"gopnik"
	"sampledata"

	. "gopkg.in/check.v1"
)

type RenderPoolSuite struct{}

var _ = Suite(&RenderPoolSuite{})

func (s *RenderPoolSuite) TestOneRender(c *C) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 1, 1, 0)
	c.Assert(err, IsNil)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	ansCh := make(chan *RenderPoolResponse)
	err = rpool.EnqueueRequest(coord, ansCh)
	c.Assert(err, IsNil)
	ans := <-ansCh
	c.Assert(ans.Error, IsNil)
	c.Assert(len(ans.Tiles), Equals, 1)
	sampledata.CheckTile(c, ans.Tiles[0].Image, "1_0_0.png")
}

func (s *RenderPoolSuite) Test5Renders(c *C) {
	const nTiles = 15

	rpool, err := NewRenderPool(sampledata.SlaveCmd, 5, nTiles, 0)
	c.Assert(err, IsNil)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	ansCh := make(chan *RenderPoolResponse)
	for i := 0; i < nTiles; i++ {
		err = rpool.EnqueueRequest(coord, ansCh)
		c.Assert(err, IsNil)
	}
	for i := 0; i < nTiles; i++ {
		ans := <-ansCh
		c.Assert(ans.Error, IsNil)
		c.Assert(len(ans.Tiles), Equals, 1)
		sampledata.CheckTile(c, ans.Tiles[0].Image, "1_0_0.png")
	}
}

func (s *RenderPoolSuite) TestTTL(c *C) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 1, 4, 2)
	c.Assert(err, IsNil)

	ansCh := make(chan *RenderPoolResponse)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			coord := gopnik.TileCoord{
				X:    uint64(i),
				Y:    uint64(j),
				Zoom: 1,
				Size: 1,
			}
			err = rpool.EnqueueRequest(coord, ansCh)
			c.Assert(err, IsNil)
		}
	}
	for i := 0; i < 4; i++ {
		ans := <-ansCh
		c.Assert(ans.Error, IsNil)
		c.Assert(len(ans.Tiles), Equals, 1)
	}
}

func (s *RenderPoolSuite) TestOneRender4Tiles(c *C) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 1, 1, 0)
	c.Assert(err, IsNil)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 2,
	}
	ansCh := make(chan *RenderPoolResponse)
	err = rpool.EnqueueRequest(coord, ansCh)
	c.Assert(err, IsNil)
	ans := <-ansCh
	c.Assert(ans.Error, IsNil)
	c.Assert(len(ans.Tiles), Equals, 4)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			sampledata.CheckTile(c, ans.Tiles[i*2+j].Image,
				fmt.Sprintf("1_%d_%d.png", j, i))
		}
	}
}

func Benchmark5Renders(b *testing.B) {
	rpool, err := NewRenderPool(sampledata.SlaveCmd, 5, uint(b.N), 0)
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
		err = rpool.EnqueueRequest(coord, ansCh)
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
