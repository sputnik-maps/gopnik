package main

import (
	"app"
	"gopnik"
	"gproj"
)

func minMaxUint64(x, y uint64) (uint64, uint64) {
	if x > y {
		return y, x
	}
	return x, y
}

func genCoords(bbox [4]float64, zoom uint64) (res []gopnik.TileCoord, err error) {
	minC := gproj.FromLLToCoord(bbox[0], bbox[2], zoom)
	maxC := gproj.FromLLToCoord(bbox[1], bbox[3], zoom)
	metaSize := app.App.Metatiler().MetaSize(zoom)
	mask := metaSize - 1
	minX, maxX := minMaxUint64(minC.X & ^mask, maxC.X & ^mask)
	minY, maxY := minMaxUint64(minC.Y & ^mask, maxC.Y & ^mask)
	for x := minX; x <= maxX; x += metaSize {
		for y := minY; y <= maxY; y += metaSize {
			res = append(res, gopnik.TileCoord{
				X:    x,
				Y:    y,
				Zoom: zoom,
				Size: metaSize,
			})
		}
	}
	return
}
