package tilerender

import (
	"gopnik"
	"sampledata"

	. "gopkg.in/check.v1"
)

type TileRenderSuite struct{}

var _ = Suite(&TileRenderSuite{})

func (s *TileRenderSuite) TestTileRender(c *C) {
	render, err := NewTileRender(sampledata.SlaveCmd)

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
