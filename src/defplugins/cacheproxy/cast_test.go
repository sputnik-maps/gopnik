package cacheproxy

import (
	"testing"

	"gopnik"
)

func TestCacheProxyCast(t *testing.T) {
	fp := new(CacheProxyPlugin)
	var v gopnik.CachePluginInterface
	v = fp
	_ = v
}
