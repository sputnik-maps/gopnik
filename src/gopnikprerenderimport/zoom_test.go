package main

import (
	. "gopkg.in/check.v1"
)

type MySuite2 struct{}

var _ = Suite(&MySuite2{})

func (s *MySuite2) TestZoom1(c *C) {
	zoom := "9"
	res, err := parseZoom(zoom)
	c.Assert(err, IsNil)
	c.Check(len(res), Equals, 1)
	c.Check(res[0], Equals, uint64(9))
}

func (s *MySuite2) TestZoom2(c *C) {
	zoom := "1-5"
	res, err := parseZoom(zoom)
	c.Assert(err, IsNil)
	c.Check(len(res), Equals, 5)
	for i := 0; i < len(res); i++ {
		c.Check(res[i], Equals, uint64(i+1))
	}
}

func (s *MySuite2) TestZoom3(c *C) {
	zoom := "1,2,3"
	res, err := parseZoom(zoom)
	c.Assert(err, IsNil)
	c.Check(len(res), Equals, 3)
	for i := 0; i < len(res); i++ {
		c.Check(res[i], Equals, uint64(i+1))
	}
}

func (s *MySuite2) TestZoom4(c *C) {
	zoom := "1-5,6,7"
	res, err := parseZoom(zoom)
	c.Assert(err, IsNil)
	c.Check(len(res), Equals, 7)
	for i := 0; i < len(res); i++ {
		c.Check(res[i], Equals, uint64(i+1))
	}
}
