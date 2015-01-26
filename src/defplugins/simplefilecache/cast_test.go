package filecache

import (
	"testing"

	"gopnik"
)

func TestSimpleFileCacheCast(t *testing.T) {
	fp := new(SimpleFileCachePlugin)
	var v gopnik.CachePluginInterface
	v = fp
	_ = v
}
