package gopnik

import (
	"testing"

	"sampledata"

	"github.com/stretchr/testify/require"
)

func TestNewNormalTile(t *testing.T) {
	img, err := sampledata.Asset("1_0_0.png")
	require.Nil(t, err)

	tile, err := NewTile(img)
	require.Nil(t, err)

	for i := 0; i < len(img); i++ {
		if tile.Image[i] != img[i] {
			t.Error("tile.Image != img")
		}
	}

	require.Nil(t, tile.SingleColor)
}

func TestNewSingleColorTile(t *testing.T) {
	img, err := sampledata.Asset("s_color.png")
	require.Nil(t, err)

	tile, err := NewTile(img)
	require.Nil(t, err)

	for i := 0; i < len(img); i++ {
		if tile.Image[i] != img[i] {
			t.Error("tile.Image != img")
		}
	}
	if tile.SingleColor == nil {
		t.Error("SingleColor should not be nil")
	}
	r, g, b, a := tile.SingleColor.RGBA()
	if r != 181|(181<<8) {
		t.Error("Invalid SingleColor r: ", r)
	}
	if g != 208|(208<<8) {
		t.Error("Invalid SingleColor g: ", g)
	}
	if b != 208|(208<<8) {
		t.Error("Invalid SingleColor b: ", b)
	}
	if a != 255|(255<<8) {
		t.Error("Invalid SingleColor a: ", a)
	}
}
