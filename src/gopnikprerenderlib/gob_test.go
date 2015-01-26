package gopnikprerenderlib

import (
	"bytes"
	"encoding/gob"
	"testing"

	"gopnik"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct {
	network *bytes.Buffer
	enc     *gob.Encoder
	dec     *gob.Decoder
}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpTest(c *C) {
	s.network = bytes.NewBuffer(nil)
	s.enc = gob.NewEncoder(s.network)
	s.dec = gob.NewDecoder(s.network)
}

func (s *MySuite) TestEncodeTaskCoord(c *C) {
	task1 := RTask{
		Type: RenderTask,
		Coord: &gopnik.TileCoord{
			X:    1,
			Y:    2,
			Zoom: 3,
			Size: 4,
		},
	}
	err := s.enc.Encode(&task1)
	c.Assert(err, IsNil)

	var task2 RTask
	err = s.dec.Decode(&task2)
	c.Assert(err, IsNil)

	c.Check(task2.Type, Equals, task1.Type)
	c.Check(task2.Config, IsNil)
	c.Check(task2.Coord, NotNil)
	// Not shared memory
	c.Check(task2.Coord, Not(Equals), task1.Coord)
	c.Check(task2.Coord.Equals(task1.Coord), Equals, true)
	task1.Coord.X += 10
	c.Check(task2.Coord.Equals(task1.Coord), Not(Equals), true)
}
