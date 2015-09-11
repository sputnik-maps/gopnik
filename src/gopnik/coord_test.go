package gopnik

import (
	"testing"
)

func TestCoordEqualss(t *testing.T) {
	coord := TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
		Tags: []string{"Tag1", "Tag2", "Tag3"},
	}

	coord2 := &TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
		Tags: []string{"Tag1", "Tag2", "Tag3"},
	}

	if !coord.Equals(coord2) {
		t.Error("Not equals, should be equal")
	}
}

func TestCoordNotEqualss(t *testing.T) {
	coord := TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
	}

	coord2 := &TileCoord{
		X:    10,
		Y:    7,
		Zoom: 12,
	}

	if coord.Equals(coord2) {
		t.Error("(coord2) Equalss, should not be equal")
	}

	coord3 := &TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
		Size: 4,
	}

	if coord.Equals(coord3) {
		t.Error("(coord3) Equalss, should not be equal")
	}
}
