package gproj

import (
	"math"
)

// This has been reimplemented based on OpenStreetMap generate_tiles.py
func minmax(a, b, c float64) float64 {
	a = math.Max(a, b)
	a = math.Min(a, c)
	return a
}

var gp struct {
	Bc []float64
	Cc []float64
	zc [][2]float64
	Ac []float64
}

func init() {
	c := 256.0
	for d := 0; d < 30; d++ {
		e := c / 2
		gp.Bc = append(gp.Bc, c/360.0)
		gp.Cc = append(gp.Cc, c/(2*math.Pi))
		gp.zc = append(gp.zc, [2]float64{e, e})
		gp.Ac = append(gp.Ac, c)
		c *= 2
	}
}

func fromLLtoPixel(ll [2]float64, zoom uint64) [2]float64 {
	d := gp.zc[zoom]
	e := math.Trunc((d[0] + ll[0]*gp.Bc[zoom]) + 0.5)
	f := minmax(math.Sin(ll[1]*math.Pi/180.0), -0.9999, 0.9999)
	g := math.Trunc((d[1] + 0.5*math.Log((1+f)/(1-f))*-gp.Cc[zoom]) + 0.5)
	return [2]float64{e, g}
}

func fromPixelToLL(px [2]float64, zoom uint64) [2]float64 {
	e := gp.zc[zoom]
	f := (px[0] - e[0]) / gp.Bc[zoom]
	g := (px[1] - e[1]) / -gp.Cc[zoom]
	h := 180.0 / math.Pi * (2*math.Atan(math.Exp(g)) - 0.5*math.Pi)
	return [2]float64{f, h}
}
