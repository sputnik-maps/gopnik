package sampledata

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	_ "image/png"
	"math"
)

type TesterLight interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

func DecodeImg(sample string) []byte {
	buf, err := base64.StdEncoding.DecodeString(sample)
	if err != nil {
		panic(err)
	}
	return buf
}

func CmpColors(c1, c2 color.Color, delta int32) bool {
	delta |= delta << 8
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	res := math.Abs(float64(r1)-float64(r2)) < float64(delta) &&
		math.Abs(float64(g1)-float64(g2)) < float64(delta) &&
		math.Abs(float64(b1)-float64(b2)) < float64(delta) &&
		math.Abs(float64(a1)-float64(a2)) < float64(delta)

	return res
}

func CheckImage(t TesterLight, img []byte, sample []byte) {
	i1, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		t.Fatal(err)
	}

	i2, _, err := image.Decode(bytes.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}

	if !i1.Bounds().Eq(i2.Bounds()) {
		t.Fatalf("Different image bounds: %v and %v", i1.Bounds(), i2.Bounds())
	}

	bounds := i1.Bounds()
	for x := bounds.Min.X; x < bounds.Max.Y; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c1 := i1.At(x, y)
			c2 := i2.At(x, y)
			if !CmpColors(c1, c2, 30) {
				t.Fatalf("Different colors at (%v, %v): %v vs %v", x, y, c1, c2)
			}
		}
	}
}

func CheckTile(t TesterLight, img []byte, sampleName string) {
	sampleImg, err := Asset(sampleName)
	if err != nil {
		t.Fatalf("Invalid sample: %v", err)
	}
	CheckImage(t, img, sampleImg)
}
