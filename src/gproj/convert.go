package gproj

import "gopnik"

func FromLLToCoord(lat, lon float64, zoom uint64) gopnik.TileCoord {
	ll := [2]float64{lon, lat}

	var pixC [2]float64 = fromLLtoPixel(ll, zoom)

	return gopnik.TileCoord{
		X:    uint64(pixC[0] / 256.0),
		Y:    uint64(pixC[1] / 256.0),
		Zoom: zoom,
		Size: 1,
	}
}

func FromCoordToLL(coord gopnik.TileCoord) (lat, lon float64, zoom uint64) {
	zoom = coord.Zoom

	pixC := [2]float64{float64(coord.X) * 256.0, float64(coord.Y) * 256.0}
	var ll [2]float64 = fromPixelToLL(pixC, zoom)
	lon = ll[0]
	lat = ll[1]
	return
}
