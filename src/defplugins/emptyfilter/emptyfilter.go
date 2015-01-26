package emptyfilter

import (
	"gopnik"
	"plugins"

	json "github.com/orofarne/strict-json"
)

type EmptyFilterPlugin struct {
}

func (self *EmptyFilterPlugin) Filter(coord gopnik.TileCoord) (gopnik.TileCoord, error) {
	return coord, nil
}

type EmptyFilterPluginFactory struct {
}

func (self *EmptyFilterPluginFactory) Name() string {
	return "EmptyFilterPlugin"
}

func (self *EmptyFilterPluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	return &EmptyFilterPlugin{}, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(EmptyFilterPluginFactory))
}
