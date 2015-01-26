package emptyfilter

import (
	"testing"

	"gopnik"
)

func TestEmptyFilterCast(t *testing.T) {
	fp := new(EmptyFilterPlugin)
	var v gopnik.FilterPluginInterface
	v = fp
	_ = v
}
