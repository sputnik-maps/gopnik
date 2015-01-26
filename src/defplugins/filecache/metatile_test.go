package filecache

import (
	"bytes"

	"gopnik"

	. "gopkg.in/check.v1"
)

type MetatileSuite struct{}

var _ = Suite(&MetatileSuite{})

func (s *MetatileSuite) TestHeaderWriter(c *C) {
	ml := &metaLayout{
		Magic: []byte("META"),
		Count: 2,
		X:     0,
		Y:     0,
		Z:     1,
		Index: []metaEntry{
			metaEntry{
				Offset: 0,
				Size:   10,
			},
		},
	}
	buf := new(bytes.Buffer)
	err := encodeHeader(buf, ml)
	c.Assert(err, IsNil)
	c.Check(buf.Len(), Equals, 28)
}

/* FIXME
func (s *MetatileSuite) TestMetatileEncoder(c *C) {
	img := sampledata.DecodeImg(sampledata.MetaTileImage)
	imgBuf := bytes.NewReader(img)
	imgImg, _, err := image.Decode(imgBuf)
	c.Assert(err, IsNil)
	tiles, err := SplitToTiles(imgImg)
	c.Assert(err, IsNil)
	var buf bytes.Buffer
	coord := tilerender.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 8,
	}
	err = EncodeMetatile(&buf, coord, tiles)
	c.Assert(err, IsNil)
}

func (s *MetatileSuite) TestMetatileDecoder(c *C) {
	img := sampledata.DecodeImg(sampledata.MetaTileImage)
	imgBuf := bytes.NewReader(img)
	imgImg, _, err := image.Decode(imgBuf)
	c.Assert(err, IsNil)
	tiles, err := SplitToTiles(imgImg)
	c.Assert(err, IsNil)
	var buf bytes.Buffer
	coord := tilerender.TileCoord{
		X:    1,
		Y:    0,
		Zoom: 1,
		Size: 8,
	}
	err = EncodeMetatile(&buf, coord, tiles)
	c.Assert(err, IsNil)

	rsBuf := bytes.NewReader(buf.Bytes())
	resTile, err := GetRawTileFromMetatile(rsBuf, coord)
	c.Assert(err, IsNil)
	tiles2BinBuf := bytes.NewBuffer(nil)
	err = png.Encode(tiles2BinBuf, tiles[2])
	c.Assert(err, IsNil)
	tiles2Bin := tiles2BinBuf.Bytes()
	c.Assert(len(resTile), Equals, len(tiles2Bin))
	for i, b := range tiles2Bin {
		c.Check(resTile[i], Equals, b)
	}
}
*/

func (s *MetatileSuite) TestMetatileHashPath(c *C) {
	coord := gopnik.TileCoord{
		X:    4946,
		Y:    2559,
		Zoom: 13,
		Size: 8,
	}
	dirName, fName := MetatileHashPath("/tmp/default", coord)
	c.Check(dirName, Equals, "/tmp/default/13/0/16/57/95")
	c.Check(fName, Equals, "8.meta")
}
