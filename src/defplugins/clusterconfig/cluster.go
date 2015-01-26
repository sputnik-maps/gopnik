package clusterconfig

import (
	"plugins"

	json "github.com/orofarne/strict-json"
)

type clusterPluginCfg struct {
	Nodes []string
}

type ClusterConfigPlugin struct {
	clusterPluginCfg
}

func (cp *ClusterConfigPlugin) Configure(cfg json.RawMessage) error {
	err := json.Unmarshal(cfg, &cp.clusterPluginCfg)
	if err != nil {
		return err
	}
	return err
}

func (cp *ClusterConfigPlugin) GetRenders() ([]string, error) {
	return cp.Nodes, nil
}

type ClusterConfigPluginFactory struct {
}

func (cpf *ClusterConfigPluginFactory) Name() string {
	return "ClusterConfigPlugin"
}

func (cpf *ClusterConfigPluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(ClusterConfigPlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(ClusterConfigPluginFactory))
}
