package tilerender

import (
	"gopnik"
	"sampledata"

	. "gopkg.in/check.v1"
	"time"
)

type TileRenderSuite struct{}

var _ = Suite(&TileRenderSuite{})

func (s *TileRenderSuite) TestTileRender(c *C) {
	render, err := NewTileRender(sampledata.SlaveCmd, time.Duration(5*time.Second))

	c.Assert(err, IsNil)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	tiles, err := render.RenderTiles(coord)

	c.Assert(err, IsNil)
	c.Assert(len(tiles), Equals, 1)

	sampledata.CheckTile(c, tiles[0].Image, "1_0_0.png")
}

func (s *TileRenderSuite) TestTimeoutTileRender(c *C) {
	render, err := NewTileRender(sampledata.SleepSlaveCmd, time.Duration(100*time.Millisecond))

	c.Assert(err, IsNil)

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	t0 := time.Now()
	_, err = render.RenderTiles(coord)

	c.Assert(err, Not(IsNil))
	c.Assert(time.Since(t0).Seconds() < 1.0, Equals, true)
}