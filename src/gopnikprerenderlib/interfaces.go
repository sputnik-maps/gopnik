package gopnikprerenderlib

import (
	"time"

	"app"
	"gopnik"
)

const (
	Hello = iota
	RenderTask
	Stat
	Error
	Config
	Monitoring
	GetTask
)

type RHello struct {
	SaverPool int
	app.RenderPoolsConfig
}

type RConfig struct {
	SaverThreads int
	app.RenderPoolsConfig
}

type RStat struct {
	RenderTime time.Duration
	SaveTime   time.Duration
}

type RTask struct {
	Type   int
	Coord  *gopnik.TileCoord
	Config *RConfig
}

type RResponse struct {
	Type  int
	Coord *gopnik.TileCoord
	Stat  *RStat
	Error string
	Hello *RHello
	Mon   string
}
