package bbox

import (
	"gopnik"
	"gproj"
)

type BBox struct {
	MinLat, MaxLat   float64
	MinLon, MaxLon   float64
	MinZoom, MaxZoom uint64
}

func minMax(a, b float64) (min, max float64) {
	if a > b {
		min = b
		max = a
	} else {
		min = a
		max = b
	}
	return
}

func contains(min, max, x float64) bool {
	return x >= min && x <= max
}

func crosses(xMin, xMax, yMin, yMax float64) bool {
	return contains(xMin, xMax, yMin) ||
		contains(xMin, xMax, yMax) ||
		contains(yMin, yMax, xMin) ||
		contains(yMin, yMax, xMax)
}

func (bb *BBox) Crosses(coord gopnik.TileCoord) bool {
	if coord.Zoom < bb.MinZoom || coord.Zoom > bb.MaxZoom {
		return false
	}

	lat1, lon1, _ := gproj.FromCoordToLL(coord)
	lat2, lon2, _ := gproj.FromCoordToLL(gopnik.TileCoord{
		X:    coord.X + coord.Size,
		Y:    coord.Y + coord.Size,
		Zoom: coord.Zoom,
	})

	latMin, latMax := minMax(lat1, lat2)
	lonMin, lonMax := minMax(lon1, lon2)

	if !crosses(bb.MinLat, bb.MaxLat, latMin, latMax) {
		return false
	}
	if !crosses(bb.MinLon, bb.MaxLon, lonMin, lonMax) {
		return false
	}
	return true
}
