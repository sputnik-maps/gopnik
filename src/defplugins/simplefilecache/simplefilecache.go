package filecache

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopnik"
	"plugins"

	json "github.com/orofarne/strict-json"
)

type simpleFileCacheConfig struct {
	Root string
}

type SimpleFileCachePlugin struct {
	root string
}

func (self *SimpleFileCachePlugin) Configure(cfg json.RawMessage) error {
	conf := simpleFileCacheConfig{
		Root: "/tmp/tiles",
	}
	err := json.Unmarshal(cfg, &conf)
	if err != nil {
		return err
	}
	self.root = conf.Root
	err = os.MkdirAll(self.root, 0777)
	return err
}

func (self *SimpleFileCachePlugin) getTileFName(coord gopnik.TileCoord) string {
	return fmt.Sprintf("%v/%v_%v_%v.png", self.root, coord.Zoom, coord.X, coord.Y)
}

func (self *SimpleFileCachePlugin) Get(coord gopnik.TileCoord) ([]byte, error) {
	data, err := ioutil.ReadFile(self.getTileFName(coord))
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	return data, err
}

func (self *SimpleFileCachePlugin) Set(coord gopnik.TileCoord, tiles []gopnik.Tile) error {
	for i, tile := range tiles {
		c := gopnik.TileCoord{
			X:    coord.X + (uint64(i) % coord.Size),
			Y:    coord.Y + (uint64(i) / coord.Size),
			Zoom: coord.Zoom,
			Size: coord.Size,
		}
		err := ioutil.WriteFile(self.getTileFName(c), tile.Image, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

type SimpleFileCachePluginFactory struct {
}

func (self *SimpleFileCachePluginFactory) Name() string {
	return "SimpleFileCachePlugin"
}

func (self *SimpleFileCachePluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(SimpleFileCachePlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(SimpleFileCachePluginFactory))
}
