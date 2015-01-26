package filecache

import (
	"testing"

	"gopnik"
)

func TestFilecacheCast(t *testing.T) {
	fp := new(FileCachePlugin)
	var v gopnik.CachePluginInterface
	v = fp
	_ = v
}
