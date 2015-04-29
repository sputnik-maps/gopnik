package bbox

import (
	"testing"

	"gopnik"

	"github.com/stretchr/testify/require"
)

func TestBBoxInnerTileShouldCrossBBox(t *testing.T) {
	// Moscow
	bb := BBox{
		MinLat:  55.542618983877674,
		MaxLat:  55.952275476109435,
		MinLon:  37.31231689453125,
		MaxLon:  37.935791015625,
		MinZoom: 9,
		MaxZoom: 18,
	}

	coord := gopnik.TileCoord{
		Zoom: 16,
		X:    39614,
		Y:    20483,
		Size: 1,
	}

	require.True(t, bb.Crosses(coord))
}

func TestBBoxOuterTileShouldNotCrossBBox(t *testing.T) {
	// Moscow
	bb := BBox{
		MinLat:  54.190566,
		MaxLat:  56.972679,
		MinLon:  35.094452,
		MaxLon:  40.321198,
		MinZoom: 9,
		MaxZoom: 18,
	}

	coord2 := gopnik.TileCoord{
		Zoom: 16,
		X:    40212,
		Y:    20421,
		Size: 1,
	}

	require.False(t, bb.Crosses(coord2))
}
