package main

import (
	"bytes"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestMinMax(c *C) {
	min, max := minMaxUint64(uint64(1), uint64(3))
	c.Check(min, Equals, uint64(1))
	c.Check(max, Equals, uint64(3))
}

func (s *MySuite) TestMinMax2(c *C) {
	min, max := minMaxUint64(uint64(4), uint64(3))
	c.Check(min, Equals, uint64(3))
	c.Check(max, Equals, uint64(4))
}

func (s *MySuite) TestCSVParsing(c *C) {
	testData := `City,min lon (left upper),min lat (right lower),max lon (right lower),max lat (left upper),Population
Москва,"35,094452","54,190566","40,321198","56,972679","11,980"
Санкт-Петербург,"29,338989","59,592756","30,893555","60,272515","5,028"
Новосибирск,"82,608948","54,789601","83,265381","55,230589","1,524"`

	reader := bytes.NewReader([]byte(testData))
	coords, err := readCSV(reader, []uint64{9})
	c.Assert(err, IsNil)
	c.Check(len(coords), Equals, 6)
}
