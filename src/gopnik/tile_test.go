package gopnik

import (
	"testing"

	"sampledata"
)

func TestNewNormalTile(t *testing.T) {
	img, err := sampledata.Asset("1_0_0.png")
	if err != nil {
		t.Error(err)
	}
	tile, err := NewTile(img)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < len(img); i++ {
		if tile.Image[i] != img[i] {
			t.Error("tile.Image != img")
		}
	}
	if tile.SingleColor != nil {
		t.Error("SingleColor should be nil")
	}
}

func TestNewSingleColorTile(t *testing.T) {
	img, err := sampledata.Asset("s_color.png")
	if err != nil {
		t.Error(err)
	}
	tile, err := NewTile(img)
	if err != nil {
		t.Error(err)
	}
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
		t.Error("Invalid SingleColor: ", r)
	}
	if g != 208|(208<<8) {
		t.Error("Invalid SingleColor: ", g)
	}
	if b != 208|(208<<8) {
		t.Error("Invalid SingleColor: ", b)
	}
	if a != 255|(255<<8) {
		t.Error("Invalid SingleColor: ", a)
	}
}
