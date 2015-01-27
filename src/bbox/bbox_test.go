package bbox

import (
	"gopnik"

	. "gopkg.in/check.v1"
)

type BBoxSuite struct{}

var _ = Suite(&BBoxSuite{})

func (s *BBoxSuite) Test1(c *C) {
	// Moscow
	bb := BBox{
		MinLat:  55.542618983877674,
		MaxLat:  55.952275476109435,
		MinLon:  37.31231689453125,
		MaxLon:  37.935791015625,
		MinZoom: 9,
		MaxZoom: 18,
	}

	coord2 := gopnik.TileCoord{
		Zoom: 10,
		X:    618,
		Y:    419,
		Size: 1,
	}
	c.Check(bb.Crosses(coord2), Equals, false)
}
