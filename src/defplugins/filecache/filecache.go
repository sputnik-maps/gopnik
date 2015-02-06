package filecache

import (
	"fmt"
	"os"
	"sync"

	"gopnik"
	"plugins"

	json "github.com/orofarne/strict-json"
)

type fileCacheConfig struct {
	Root    string
	UseHash bool
}

type FileCachePlugin struct {
	cfg fileCacheConfig
	mu  sync.RWMutex
}

func (self *FileCachePlugin) Configure(cfg json.RawMessage) error {
	self.cfg.Root = "/tmp/tiles"
	self.cfg.UseHash = true
	return json.Unmarshal(cfg, &self.cfg)
}

func (self *FileCachePlugin) getMetatilePath(coord gopnik.TileCoord) (dirName, fName string) {
	dirName, fName = MetatileHashPath(self.cfg.Root, coord)
	return
}

func (self *FileCachePlugin) Get(coord gopnik.TileCoord) ([]byte, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	dirName, fName := self.getMetatilePath(coord)
	file, err := os.Open(fmt.Sprintf("%v/%v", dirName, fName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return GetRawTileFromMetatile(file, coord)
}

func (self *FileCachePlugin) Set(coord gopnik.TileCoord, tiles []gopnik.Tile) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	dirName, fName := self.getMetatilePath(coord)
	err := os.MkdirAll(dirName, 0777)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(fmt.Sprintf("%v/%v", dirName, fName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	return EncodeMetatile(file, coord, tiles)
}

type FileCachePluginFactory struct {
}

func (self *FileCachePluginFactory) Name() string {
	return "FileCachePlugin"
}

func (self *FileCachePluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(FileCachePlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(FileCachePluginFactory))
}
