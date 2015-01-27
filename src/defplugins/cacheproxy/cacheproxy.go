package cacheproxy

import (
	"fmt"

	"gopnik"
	"plugins"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

type cacheProxyPluginBackend struct {
	Name     string
	Plugin   gopnik.CachePluginInterface
	MinZoom  uint64
	MaxZoom  uint64
	Tags     []string
	ReadOnly bool
}

type cacheProxyPluginBackendConf struct {
	Plugin       string
	PluginConfig json.RawMessage
	MinZoom      uint64
	MaxZoom      uint64
	Tags         []string
	ReadOnly     bool
}

type cacheProxyPluginConf struct {
	Backends      []cacheProxyPluginBackendConf
	WriteToFirst  bool
	ReadFromFirst bool
	WriteToAny    bool
}

type CacheProxyPlugin struct {
	config        cacheProxyPluginConf
	backends      []cacheProxyPluginBackend
	readFromFirst bool
}

func (self *CacheProxyPlugin) Configure(cfg json.RawMessage) error {
	// Parse config
	err := json.Unmarshal(cfg, &self.config)
	if err != nil {
		return err
	}

	// Configure backends
	for _, pConf := range self.config.Backends {
		if plug, err := plugins.DefaultPluginStore.Create(pConf.Plugin, pConf.PluginConfig); err == nil {
			if cPlug, correctPlug := plug.(gopnik.CachePluginInterface); correctPlug {
				self.backends = append(self.backends, cacheProxyPluginBackend{
					Name:     pConf.Plugin,
					Plugin:   cPlug,
					MinZoom:  pConf.MinZoom,
					MaxZoom:  pConf.MaxZoom,
					Tags:     pConf.Tags,
					ReadOnly: pConf.ReadOnly,
				})
			} else {
				return fmt.Errorf("Plugin '%s' is not CachePluginInterface", pConf.Plugin)
			}
		} else {
			return fmt.Errorf("Plugin '%s' not configured: '%v'", pConf.Plugin, err)
		}
	}

	return nil
}

func (self *CacheProxyPlugin) backendIsResonable(backend *cacheProxyPluginBackend, coord *gopnik.TileCoord) bool {
	if backend.Tags != nil {
	TL:
		for _, cfgT := range backend.Tags {
			for _, inT := range coord.Tags {
				if inT == cfgT {
					continue TL
				}
			}
			return false
		}
	}
	if coord.Zoom < backend.MinZoom || coord.Zoom > backend.MaxZoom {
		return false
	}
	return true
}

func (self *CacheProxyPlugin) Get(coord gopnik.TileCoord) ([]byte, error) {
	var errs []error
	var n = 0

BL:
	for _, backend := range self.backends {
		if !self.backendIsResonable(&backend, &coord) {
			log.Debug("Backend %v is unresonable", backend)
			continue BL
		}
		log.Debug("Backend %v is resonable", backend)

		data, getErr := backend.Plugin.Get(coord)
		if getErr != nil {
			errs = append(errs, fmt.Errorf("%v Get error: %v", backend.Name, getErr))
			log.Error("CacheProxyPlugin backend `%v` error: %v", backend.Name, getErr)
			if self.readFromFirst {
				break
			}
			continue
		}
		if data != nil || self.readFromFirst {
			return data, nil
		}
		n++

		if self.config.ReadFromFirst {
			break
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("CacheProxyPlugin errors: %v", errs)
	}
	if n == 0 {
		return nil, fmt.Errorf("No backends for zoom %v", coord.Zoom)
	}
	return nil, nil // Nothig was found
}

func (self *CacheProxyPlugin) Set(coord gopnik.TileCoord, tiles []gopnik.Tile) error {
	errChan := make(chan error)

	var n = 0
	for _, backPlug := range self.backends {
		if !backPlug.ReadOnly {
			if !self.backendIsResonable(&backPlug, &coord) {
				continue
			}
			n++

			go func(plug gopnik.CachePluginInterface) {
				errChan <- plug.Set(coord, tiles)
			}(backPlug.Plugin)

			if self.config.WriteToFirst {
				break
			}
		}
	}

	if n == 0 {
		return fmt.Errorf("No backend to write")
	}

	var errs []error
	for i := 0; i < n; i++ {
		err := <-errChan
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 || (self.config.WriteToAny && len(errs) < n) {
		return nil
	}

	return fmt.Errorf("CacheProxy errors: %v", errs)
}

type CacheProxyPluginFactory struct {
}

func (self *CacheProxyPluginFactory) Name() string {
	return "CacheProxyPlugin"
}

func (cpf *CacheProxyPluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(CacheProxyPlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(CacheProxyPluginFactory))
}
