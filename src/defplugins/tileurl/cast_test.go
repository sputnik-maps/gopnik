package tileurl

import (
	"testing"

	"gopnik"
)

func TestFilecacheCast(t *testing.T) {
	fp := new(TileUrlPlugin)
	var v gopnik.CachePluginInterface
	v = fp
	_ = v
}
