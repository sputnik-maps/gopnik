package filecache

/* FIXME
import (
	"bytes"
	"os"
	"testing"

	"gopnik"
	"sampledata"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilecacheSuite struct{}

var _ = Suite(&FilecacheSuite{})

func (s *FilecacheSuite) TestSimpleSetGet(c *C) {
	root := "/tmp/FileCachePlugin/Test"

	os.RemoveAll(root)

	cache := FileCachePlugin{
		cfg: fileCacheConfig{
			Root:    root,
			UseHash: true,
		},
	}

	imgBin := sampledata.DecodeImg(sampledata.MetaTileImage)

	coord := tilerender.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 14,
		Size: 2,
	}

	c.Assert(len(tiles), Equals, 4)
	err = cache.Set(coord, tiles, tilesBin)
	c.Assert(err, IsNil)

	coord.X += 1 // 3th tile
	coord.Size = 1
	buf, err := cache.Get(coord)
	c.Assert(err, IsNil)
	buf2 := bytes.NewBuffer(nil)
	err = png.Encode(buf2, tiles[2])
	c.Assert(err, IsNil)
	c.Assert(buf2.Bytes(), Not(IsNil))
	for i, b := range buf2.Bytes() {
		c.Check(buf[i], Equals, b)
	}
}
*/
