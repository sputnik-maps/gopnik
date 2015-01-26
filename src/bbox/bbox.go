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

func minMax(a, b uint64) (min, max uint64) {
	if a > b {
		min = b
		max = a
	} else {
		min = a
		max = b
	}
	return
}

func contains(min, max, x uint64) bool {
	return x > min && x < max
}

func crosses(xMin, xMax, yMin, yMax uint64) bool {
	return contains(xMin, xMax, yMin) ||
		contains(xMin, xMax, yMax) ||
		contains(yMin, yMax, xMin) ||
		contains(yMin, yMax, xMax)
}

func (bb *BBox) Crosses(coord gopnik.TileCoord) bool {
	if coord.Zoom < bb.MinZoom || coord.Zoom > bb.MaxZoom {
		return false
	}

	bbCoord1 := gproj.FromLLToCoord(bb.MinLat, bb.MinLon, coord.Zoom)
	bbCoord2 := gproj.FromLLToCoord(bb.MaxLat, bb.MaxLon, coord.Zoom)

	bbMinX, bbMaxX := minMax(bbCoord1.X, bbCoord2.X)
	bbMinY, bbMaxY := minMax(bbCoord1.Y, bbCoord2.Y)

	if !crosses(bbMinX, bbMaxX, coord.X, coord.X+coord.Size) {
		return false
	}
	if !crosses(bbMinY, bbMaxY, coord.Y, coord.Y+coord.Size) {
		return false
	}
	return true
}
