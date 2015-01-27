package memorykv

import (
	"plugins"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

type MemoryKV struct {
	store map[string][]byte
}

type MemoryKVFactory struct {
}

func (self *MemoryKVFactory) Name() string {
	return "MemoryKV"
}

func (self *MemoryKVFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(MemoryKV)
	res.store = make(map[string][]byte)
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(MemoryKVFactory))
}

func (self *MemoryKV) Get(key string) (data []byte, err error) {
	log.Debug("Request data by key '%v' from memory...", key)
	data = self.store[key]
	return
}

func (self *MemoryKV) Set(key string, value []byte) (err error) {
	log.Debug("Save data by key '%v' to memory", key)
	self.store[key] = value
	return
}

func (self *MemoryKV) Delete(key string) (err error) {
	log.Debug("Delete data by key '%v' to memory", key)
	delete(self.store, key)
	return
}
