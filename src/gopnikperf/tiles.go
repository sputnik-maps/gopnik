package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	"app"
	"gopnik"
	"gproj"
	"perflog"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/davecgh/go-spew/spew"
)

func getTileColor(perfData []perflog.PerfLogEntry, coord gopnik.TileCoord, maxTime float64) (col color.Color, renderTimeSeconds float64) {
	metaCoord := app.App.Metatiler().TileToMetatile(&coord)
	for _, entry := range perfData {
		if entry.Coord.Equals(&metaCoord) {
			col := color.RGBA{
				R: 255,
				G: 10,
				B: 10,
				A: uint8((entry.RenderTime.Seconds() / maxTime) * 200),
			}
			return col, entry.RenderTime.Seconds()
		}
	}
	return color.Transparent, -1.0
}

func convertCoordByZoom(coord gopnik.TileCoord, zoom uint64) gopnik.TileCoord {
	lat, lon, _ := gproj.FromCoordToLL(coord)
	return gproj.FromLLToCoord(lat, lon, zoom)
}

func getMaxTime(perfData []perflog.PerfLogEntry, zoom uint64) float64 {
	var maxTime float64 = 0.0
	for _, entry := range perfData {
		if entry.Coord.Zoom == zoom && entry.RenderTime.Seconds() > maxTime {
			maxTime = entry.RenderTime.Seconds()
		}
	}
	return maxTime
}

func convertTime(t float64) string {
	if t < 0. {
		return "--"
	}
	if t >= 86400. {
		return fmt.Sprintf("%.2fd", t/86400.)
	}
	if t >= 3600. {
		return fmt.Sprintf("%.2fh", t/3600.)
	}
	if t >= 60. {
		return fmt.Sprintf("%.2fm", t/60.)
	}
	return fmt.Sprintf("%.2fs", t)
}

func drawTime(img *image.RGBA, t float64) error {
	ctx := freetype.NewContext()
	fntData, err := Asset("public/fonts/UbuntuMono-R.ttf")
	if err != nil {
		return err
	}
	fnt, err := truetype.Parse(fntData)
	if err != nil {
		return err
	}
	ctx.SetFont(fnt)
	ctx.SetFontSize(22)
	ctx.SetSrc(&image.Uniform{color.Black})
	ctx.SetDst(img)
	ctx.SetClip(img.Bounds())
	pt := freetype.Pt(30, 30)
	tStr := convertTime(t)
	_, err = ctx.DrawString(tStr, pt)
	return err
}

func genPerfTile(perfData []perflog.PerfLogEntry, coord gopnik.TileCoord, zoom uint64) ([]byte, error) {
	// Get max time
	maxTime := getMaxTime(perfData, zoom)

	// Convert coordinates
	coordMin := convertCoordByZoom(coord, zoom)
	coord2 := gopnik.TileCoord{
		X:    coord.X + coord.Size,
		Y:    coord.Y + coord.Size,
		Zoom: coord.Zoom,
		Size: coord.Size,
	}
	coordMax := convertCoordByZoom(coord2, zoom)

	// Find all subtiles
	bounds := image.Rect(0, 0, 256, 256)
	img := image.NewRGBA(bounds)

	mult := 256 / math.Pow(2, float64(zoom-coord.Zoom))
	maxTimeS := float64(-1.)
	for x := coordMin.X; x < coordMax.X || (x == coordMin.X && x == coordMax.X); x += coord.Size {
		for y := coordMin.Y; y < coordMax.Y || (y == coordMin.Y && y == coordMax.Y); y += coord.Size {
			c := gopnik.TileCoord{
				X:    x,
				Y:    y,
				Zoom: zoom,
				Size: coord.Size,
			}
			col, renderTime := getTileColor(perfData, c, maxTime)
			if renderTime > maxTimeS {
				maxTimeS = renderTime
			}
			rect := image.Rect(
				int(math.Max(float64((c.X-coordMin.X)/c.Size)*mult, 0)),
				int(math.Max(float64((c.Y-coordMin.Y)/c.Size)*mult, 0)),
				int(math.Min(float64((c.X-coordMin.X)/c.Size+1)*mult, 256)),
				int(math.Min(float64((c.Y-coordMin.Y)/c.Size+1)*mult, 256)))
			draw.Draw(img, rect, &image.Uniform{col}, image.ZP, draw.Src)
		}
	}

	drawErr := drawTime(img, maxTimeS)
	if drawErr != nil {
		spew.Dump(drawErr)
		return nil, drawErr
	}

	outbuf := bytes.NewBuffer(nil)
	err := png.Encode(outbuf, img)
	if err != nil {
		return nil, err
	}

	return outbuf.Bytes(), nil
}
