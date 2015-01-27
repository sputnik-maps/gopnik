package plugins_enabled

import (
	"encoding/json"
	"testing"

	"app"
	"gopnik"
	"plugins"
	"sampledata"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type StorePluginSuite struct{}

var _ = Suite(&StorePluginSuite{})

func (s *StorePluginSuite) EmptyTest(c *C) {
	c.Assert(true, Equals, true)
}

func (s *StorePluginSuite) GetSetTest(c *C, cfgStr string) {
	var cfg app.PluginConfig
	err := json.Unmarshal([]byte(cfgStr), &cfg)
	c.Assert(err, IsNil)

	var tiles []gopnik.Tile
	for _, fname := range []string{"1_0_0.png", "1_1_0.png", "1_0_1.png", "1_1_1.png"} {
		data, err := sampledata.Asset(fname)
		c.Assert(err, IsNil)
		tile, err := gopnik.NewTile(data)
		c.Assert(err, IsNil)
		c.Assert(tile, Not(IsNil))
		tiles = append(tiles, *tile)
	}

	plgn, err := plugins.DefaultPluginStore.Create(cfg.Plugin, cfg.PluginConfig)
	c.Assert(err, IsNil)
	store, ok := plgn.(gopnik.CachePluginInterface)
	c.Assert(ok, Equals, true)

	err = store.Set(gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 2,
	}, tiles)
	// Set error should be nil
	c.Assert(err, IsNil)

	t01, err := store.Get(gopnik.TileCoord{
		X:    0,
		Y:    1,
		Zoom: 1,
		Size: 1,
	})
	// Get error should be nil
	c.Assert(err, IsNil)

	t01Orig, err := sampledata.Asset("1_0_1.png")
	c.Assert(err, IsNil)
	sampledata.CheckImage(c, t01, t01Orig)
}
