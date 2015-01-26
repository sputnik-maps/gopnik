package gopnik

import (
	"testing"
)

func TestMetatilerCreate(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	if metatiler.TileSize() != 256 {
		t.Errorf("Invalid TileSize: %v != 256", metatiler.TileSize())
	}
}

func TestMetatilerNormalZoom(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	if metatiler.MetaSize(uint64(10)) != 8 {
		t.Errorf("Invalid MetaSize(10): %v != 8", metatiler.MetaSize(uint64(10)))
	}

	if metatiler.NTiles(uint64(10)) != 8*8 {
		t.Errorf("Invalid NTiles(10): %v != %v", metatiler.NTiles(uint64(10)), 8*8)
	}
}

func TestMetatilerCoordNormalZoom(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	coord := TileCoord{
		X:    133,
		Y:    246,
		Zoom: 11,
		Size: 1,
	}

	metaCoordMaster := TileCoord{
		X:    128,
		Y:    240,
		Zoom: 11,
		Size: 8,
	}

	metaCoord := metatiler.TileToMetatile(&coord)

	if !metaCoordMaster.Equals(&metaCoord) {
		t.Errorf("Coords differs: %v != %v", metaCoord, metaCoordMaster)
	}
}

func TestMetatiler1Zoom(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	if metatiler.MetaSize(uint64(1)) != 2 {
		t.Errorf("Invalid MetaSize(1): %v != 2", metatiler.MetaSize(uint64(1)))
	}

	if metatiler.NTiles(uint64(1)) != 2*2 {
		t.Errorf("Invalid NTiles(1): %v != %v", metatiler.NTiles(uint64(1)), 2*2)
	}
}

func TestMetatilerCoord1Zoom(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	coord := TileCoord{
		X:    1,
		Y:    1,
		Zoom: 1,
		Size: 1,
	}

	metaCoordMaster := TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 2,
	}

	metaCoord := metatiler.TileToMetatile(&coord)

	if !metaCoordMaster.Equals(&metaCoord) {
		t.Errorf("Coords differs: %v != %v", metaCoord, metaCoordMaster)
	}
}
