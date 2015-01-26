package memorykv

import (
	"gopnik"
	"testing"
)

func TestMemoryKVCast(t *testing.T) {
	fp := new(MemoryKV)
	var v gopnik.KVStore
	v = fp
	_ = v
}
