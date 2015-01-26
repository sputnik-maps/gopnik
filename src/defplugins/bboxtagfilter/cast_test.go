package bboxtagfilter

import (
	"testing"

	"gopnik"
)

func TestBBoxTagFilterCast(t *testing.T) {
	fp := new(BBoxTagFilterPlugin)
	var v gopnik.FilterPluginInterface
	v = fp
	_ = v
}
