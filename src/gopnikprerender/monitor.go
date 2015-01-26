package main

import (
	"sort"
	"sync"
	"time"
)

type monitorPoint struct {
	Timestamp int64
	Value     float64
}

type monitor struct {
	data   map[string][]monitorPoint
	dataMu sync.RWMutex
}

func newMonitor() *monitor {
	mon := new(monitor)
	mon.data = make(map[string][]monitorPoint)
	return mon
}

func (mon *monitor) AddPoint(metric string, ts time.Time, val float64) {
	mon.dataMu.Lock()
	defer mon.dataMu.Unlock()
	mon.data[metric] = append(mon.data[metric], monitorPoint{ts.Unix(), val})
}

func (mon *monitor) Points(metric string) []monitorPoint {
	mon.dataMu.RLock()
	defer mon.dataMu.RUnlock()
	data, _ := mon.data[metric]
	return data
}

func (mon *monitor) Metrics() []string {
	mon.dataMu.RLock()
	defer mon.dataMu.RUnlock()

	var res []string
	for k, _ := range mon.data {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}
