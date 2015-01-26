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
	/*
		coord1 := gopnik.TileCoord{
			Zoom: 14,
			X:    9895,
			Y:    5118,
			Size: 1,
		}
		c.Check(bb.Crosses(coord1), Equals, true)
	*/
	coord2 := gopnik.TileCoord{
		Zoom: 10,
		X:    618,
		Y:    419,
		Size: 1,
	}
	c.Check(bb.Crosses(coord2), Equals, false)
	/*
		coord3 := gopnik.TileCoord{
			Zoom: 18,
			X:    158462,
			Y:    81985,
			Size: 8,
		}
		c.Check(bb.Crosses(coord3), Equals, true)

		coord4 := gopnik.TileCoord{
			Zoom: 18,
			X:    158462,
			Y:    81985,
			Size: 1000,
		}
		c.Check(bb.Crosses(coord4), Equals, true)
	*/
}
