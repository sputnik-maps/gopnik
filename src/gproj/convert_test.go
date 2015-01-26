package gproj

import (
	"math"
	"testing"
)

func TestInternalConvertors(t *testing.T) {
	lat := 55.89067
	lon := 37.59057
	zoom := uint64(16)
	ll0 := [2]float64{lat, lon}

	pixC := fromLLtoPixel(ll0, zoom)
	ll1 := fromPixelToLL(pixC, zoom)

	ε := float64(0.0001)
	if math.Abs(ll1[0]-ll0[0]) > ε || math.Abs(ll1[1]-ll0[1]) > ε {
		t.Errorf("%v != %v", ll1, ll0)
	}
}

func TestFromLLToCoord(t *testing.T) {
	lat := 55.89069
	lon := 37.58974
	zoom := uint64(16)

	coord := FromLLToCoord(lat, lon, zoom)

	if coord.Size != 1 {
		t.Errorf("Invalid size: %v != %v", coord.Size, 1)
	}
	if coord.Zoom != zoom {
		t.Errorf("Invalid zoom: %v != %v", coord.Zoom, zoom)
	}
	if coord.X != 39611 {
		t.Errorf("Invalid x: %v != %v", coord.X, 39611)
	}
	if coord.Y != 20443 {
		t.Errorf("Invalid y: %v != %v", coord.Y, 20443)
	}
}

func TestFromCoordToLL(t *testing.T) {
	lat := 55.89069
	lon := 37.58974
	zoom := uint64(16)

	coord := FromLLToCoord(lat, lon, zoom)

	lat2, lon2, zoom2 := FromCoordToLL(coord)

	ε := float64(0.0001)
	if math.Abs(lat2-lat) > ε {
		t.Errorf("Invalid lat: %v != %v", lat2, lat)
	}
	if math.Abs(lon2-lon) > ε {
		t.Errorf("Invalid lon: %v != %v", lon2, lon)
	}
	if zoom2 != zoom {
		t.Errorf("Invalid zoom: %v != %v", zoom2, zoom)
	}

}
