package main

import (
	"gopnik"
	"time"
)

type tileReport struct {
	Coord      gopnik.TileCoord
	Error      error
	RenderTime time.Duration
	SaveTime   time.Duration
}
