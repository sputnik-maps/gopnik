package couchcache

import (
	"testing"

	"gopnik"
)

func TestCouchbaseKVCast(t *testing.T) {
	fp := new(KVStorePlugin)
	var v gopnik.CachePluginInterface
	v = fp
	_ = v
}
