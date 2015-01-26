package fakecache

import (
	"testing"

	"gopnik"
)

func TestFilecacheCast(t *testing.T) {
	fp := new(FakeCachePlugin)
	var v gopnik.CachePluginInterface
	v = fp
	_ = v
}
