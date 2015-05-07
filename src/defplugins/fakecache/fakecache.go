package fakecache

import (
	"errors"
	"time"

	json "github.com/orofarne/strict-json"

	"gopnik"
	"plugins"
	"sampledata"
)

type fakeCachePluginConf struct {
	UseStubImage bool
	GetSleep     string
	SetSleep     string
	GetError     string
	SetError     string
}

type FakeCachePlugin struct {
	config   fakeCachePluginConf
	img      []byte
	getSleep time.Duration
	setSleep time.Duration
	setError error
	getError error
}

func (self *FakeCachePlugin) Configure(cfg json.RawMessage) error {
	err := json.Unmarshal(cfg, &self.config)
	if err != nil {
		return err
	}
	if self.config.UseStubImage {
		self.img, err = sampledata.Asset("1_0_0.png")
		if err != nil {
			return err
		}
	}

	if self.config.GetSleep != "" {
		var err error
		self.getSleep, err = time.ParseDuration(self.config.GetSleep)
		if err != nil {
			return nil
		}
	}
	if self.config.SetSleep != "" {
		var err error
		self.setSleep, err = time.ParseDuration(self.config.SetSleep)
		if err != nil {
			return nil
		}
	}

	if self.config.SetError != "" {
		self.setError = errors.New(self.config.SetError)
	}
	if self.config.GetError != "" {
		self.getError = errors.New(self.config.GetError)
	}

	return nil
}

func (self *FakeCachePlugin) Get(gopnik.TileCoord) ([]byte, error) {
	time.Sleep(self.getSleep)
	if self.getError != nil {
		return nil, self.getError
	}
	return self.img, nil
}

func (self *FakeCachePlugin) Set(gopnik.TileCoord, []gopnik.Tile) error {
	time.Sleep(self.setSleep)
	return self.setError
}

type FakeCachePluginFactory struct {
}

func (self *FakeCachePluginFactory) Name() string {
	return "FakeCachePlugin"
}

func (self *FakeCachePluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(FakeCachePlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(FakeCachePluginFactory))
}
