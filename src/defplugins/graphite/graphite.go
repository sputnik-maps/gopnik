package graphiteplugin

import (
	"plugins"

	"github.com/orofarne/hmetrics2-graphite"
	json "github.com/orofarne/strict-json"
)

type graphitePluginCfg struct {
	Host   string
	Port   int
	Prefix string
}

type GraphitePlugin struct {
	cfg graphitePluginCfg
}

func (self *GraphitePlugin) Configure(cfg json.RawMessage) error {
	err := json.Unmarshal(cfg, &self.cfg)
	if err != nil {
		return err
	}
	return err
}

func (self *GraphitePlugin) Exporter() (func(map[string]float64), error) {
	return hmetrics2graphite.Exporter(self.cfg.Host, self.cfg.Port, self.cfg.Prefix)
}

type GraphitePluginFactory struct {
}

func (self *GraphitePluginFactory) Name() string {
	return "GraphitePlugin"
}

func (self *GraphitePluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(GraphitePlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(GraphitePluginFactory))
}
