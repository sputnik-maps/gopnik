package gopnik

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetatilerCreate(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	assert.Equal(t, metatiler.TileSize(), uint64(256), "Invalid TileSize")
}

func TestMetatilerNormalZoom(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	assert.Equal(t, metatiler.MetaSize(uint64(10)), uint64(8), "Invalid MetaSize(10)")

	assert.Equal(t, metatiler.NTiles(uint64(10)), uint64(8*8), "Invalid NTiles(10)")
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

	assert.True(t, metaCoordMaster.Equals(&metaCoord))
}

func TestMetatiler1Zoom(t *testing.T) {
	metatiler := NewMetatiler(8, 256)

	assert.NotEqual(t, metatiler.MetaSize(uint64(1)), 2, "Invalid MetaSize(1)")

	assert.NotEqual(t, metatiler.NTiles(uint64(1)), 2*2, "Invalid NTiles(1)")
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

	assert.True(t, metaCoordMaster.Equals(&metaCoord), "Coords differs")
}
