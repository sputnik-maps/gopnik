package plugins

import (
	"fmt"

	json "github.com/orofarne/strict-json"
)

type PluginFactory interface {
	Name() string
	New(json.RawMessage) (interface{}, error)
}

type PluginStore struct {
	plugins map[string]PluginFactory
}

var DefaultPluginStore *PluginStore = NewPluginStore()

func NewPluginStore() *PluginStore {
	self := &PluginStore{}

	self.plugins = make(map[string]PluginFactory)

	return self
}

func (self *PluginStore) Create(pluginName string, pluginConfig json.RawMessage) (interface{}, error) {
	plugFactory, found := self.plugins[pluginName]
	if !found {
		return nil, fmt.Errorf("Plugin %s not found", pluginName)
	}
	plug, err := plugFactory.New(pluginConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create plugin '%s': %v", pluginName, err)
	}
	return plug, nil
}

func (self *PluginStore) AddPlugin(plugin PluginFactory) error {
	name := plugin.Name()
	if _, ok := self.plugins[name]; ok {
		return fmt.Errorf("Plugin with name '%s' exists", name)
	}
	self.plugins[name] = plugin
	return nil
}

func (self *PluginStore) String() string {
	var resArr []string
	for k, _ := range self.plugins {
		resArr = append(resArr, k)
	}
	return fmt.Sprintf("%v", resArr)
}
