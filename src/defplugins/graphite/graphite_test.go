package graphiteplugin

import (
	"testing"

	"gopnik"
)

func TestCacheProxyCast(t *testing.T) {
	gr := new(GraphitePlugin)
	var v gopnik.MonitoringPluginInterface
	v = gr
	_ = v
}
