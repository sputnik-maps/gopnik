package gopnik

import (
	"bytes"
	"image"
	"image/color"
	_ "image/png"
)

type Tile struct {
	Image       []byte
	SingleColor color.Color
}

func NewTile(rawImage []byte) (*Tile, error) {
	res := &Tile{
		Image: rawImage,
	}

	img, _, err := image.Decode(bytes.NewReader(rawImage))
	if err != nil {
		return nil, err
	}

	cmp := func(a, b color.Color) bool {
		ar, ag, ab, aa := a.RGBA()
		br, bg, bb, ba := b.RGBA()
		return ar == br && ag == bg && ab == bb && aa == ba
	}

	bounds := img.Bounds()
	res.SingleColor = img.At(bounds.Min.X, bounds.Min.Y)
L:
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if !cmp(img.At(x, y), res.SingleColor) {
				res.SingleColor = nil
				break L
			}
		}
	}

	return res, nil
}
