package couchcache

/* FIXME

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"imgsplitter"
	"sampledata"
	"testing"
	"tilerender"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CouchcacheSuite struct {
	CouchbaseCachePlugin
}

var _ = Suite(&CouchcacheSuite{})

func (s *CouchcacheSuite) SetUpSuite(c *C) {
	err := s.init()
	c.Assert(err, IsNil)
	s.bucket = newMockCouchbase()
}

func (s *CouchcacheSuite) TearDownTest(c *C) {
	s.bucket.(*mockCouchbase).Cache = make(map[string][]byte)
}

func (s *CouchcacheSuite) TestSimpleSetGet(c *C) {
	imgsplitter.SetMetaSize(2)
	s.config.UseMultilevel = true

	imgBin := sampledata.DecodeImg(sampledata.MetaTileImage)
	imgBuf := bytes.NewReader(imgBin)
	img, _, err := image.Decode(imgBuf)
	c.Assert(err, IsNil)
	tiles, err := imgsplitter.SplitToTiles(img)
	if err != nil {
		c.Fatalf("SplitToTiles error: %v", err)
	}

	tilesBin := make([][]byte, len(tiles))
	for i, tile := range tiles {
		buf := bytes.NewBuffer(nil)
		err = png.Encode(buf, tile)
		tilesBin[i] = buf.Bytes()
	}

	coord := tilerender.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 14,
		Size: 2,
	}
	metacoord := imgsplitter.TileToMetatile(coord)

	c.Assert(len(tiles), Equals, 4)
	err = s.Set(metacoord, tiles, tilesBin)
	c.Assert(err, IsNil)

	coord.X += 1 // 3th tile
	coord.Size = 1
	buf, err := s.Get(coord)
	c.Assert(err, IsNil)
	buf2 := bytes.NewBuffer(nil)
	err = png.Encode(buf2, tiles[2])
	c.Assert(err, IsNil)
	c.Assert(buf2.Bytes(), Not(IsNil))
	for i, b := range buf2.Bytes() {
		c.Check(buf[i], Equals, b)
	}
}

func (s *CouchcacheSuite) TestEmptyTile(c *C) {
	imgsplitter.SetMetaSize(2)
	s.config.UseMultilevel = true

	bounds := image.Rect(0, 0, 256, 256)
	wimg := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			wimg.Set(x, y, color.White)
		}
	}
	outbuf := bytes.NewBuffer(nil)
	err := png.Encode(outbuf, wimg)
	c.Assert(err, IsNil)
	wimgBuf := outbuf.Bytes()

	imgBin := sampledata.DecodeImg(sampledata.MetaTileImage)
	imgBuf := bytes.NewReader(imgBin)
	img, _, err := image.Decode(imgBuf)
	c.Assert(err, IsNil)
	tiles, err := imgsplitter.SplitToTiles(img)
	if err != nil {
		c.Fatalf("SplitToTiles error: %v", err)
	}
	tiles[1] = wimg

	tilesBin := make([][]byte, len(tiles))
	for i, tile := range tiles {
		buf := bytes.NewBuffer(nil)
		err = png.Encode(buf, tile)
		tilesBin[i] = buf.Bytes()
	}

	coord := tilerender.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 14,
		Size: 2,
	}
	metacoord := imgsplitter.TileToMetatile(coord)

	c.Assert(len(tiles), Equals, 4)
	err = s.Set(metacoord, tiles, tilesBin)
	c.Assert(err, IsNil)

	coord.X = 1 // 3th tile
	coord.Size = 1
	buf, err := s.Get(coord)
	c.Assert(err, IsNil)
	buf2 := bytes.NewBuffer(nil)
	err = png.Encode(buf2, tiles[2])
	c.Assert(err, IsNil)
	c.Assert(buf2.Bytes(), Not(IsNil))
	for i, b := range buf2.Bytes() {
		c.Check(buf[i], Equals, b)
	}

	coord.X = 0
	coord.Y = 1
	coord.Size = 1
	buf, err = s.Get(coord)
	c.Assert(err, IsNil)
	c.Assert(len(buf), Not(Equals), 0)
	for i, b := range wimgBuf {
		c.Check(buf[i], Equals, b)
	}
}
*/
