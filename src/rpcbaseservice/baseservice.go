package rpcbaseservice

import (
	"math"
	"sync"

	"github.com/orofarne/hmetrics2"

	"app"
	"program_version"
	"servicestatus"
)

type Service struct {
	data   map[string]float64
	dataMu sync.Mutex
}

func NewService() *Service {
	self := &Service{}
	self.data = make(map[string]float64)

	// Metircs
	hmetrics2.AddHook(func(newData map[string]float64) {
		self.dataMu.Lock()
		defer self.dataMu.Unlock()
		self.data = make(map[string]float64)
		for k, v := range newData {
			if !math.IsNaN(v) && !math.IsInf(v, 0) {
				self.data[k] = v
			}
		}
	})

	return self
}

func (self *Service) Status() (r bool, err error) {
	return servicestatus.IsOk(), nil
}

func (self *Service) Version() (r string, err error) {
	return program_version.GetVersion(), nil
}

func (self *Service) Config() (r string, err error) {
	return app.App.Config(), nil
}

func (self *Service) Stat() (r map[string]float64, err error) {
	self.dataMu.Lock()
	defer self.dataMu.Unlock()
	return self.data, nil
}
