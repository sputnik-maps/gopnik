package couchcache

import (
	"fmt"
)

type mockCouchbase struct {
	Cache map[string][]byte
}

func newMockCouchbase() *mockCouchbase {
	c := new(mockCouchbase)
	c.Cache = make(map[string][]byte)
	return c
}

func (c *mockCouchbase) GetRaw(k string) ([]byte, error) {
	val, found := c.Cache[k]
	if !found {
		return nil, fmt.Errorf("MCResponse status=KEY_ENOENT")
	}
	return val, nil
}
func (c *mockCouchbase) SetRaw(k string, exp int, v []byte) error {
	c.Cache[k] = v
	return nil
}

func (c *mockCouchbase) AddRaw(k string, exp int, v []byte) (added bool, err error) {
	_, found := c.Cache[k]
	c.Cache[k] = v
	return !found, nil
}
