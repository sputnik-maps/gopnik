package bboxtagfilter

import (
	"bbox"
	"gopnik"
	"plugins"

	json "github.com/orofarne/strict-json"
)

type filterItem struct {
	BBoxes        []bbox.BBox
	Drop          []string
	Add           []string
	DropOtherwise []string
	AddOtherwise  []string
}

type bBoxTagFilterPluginConfig struct {
	Rules []filterItem
}

type BBoxTagFilterPlugin struct {
	rules []filterItem
}

func strArrDrop(arr []string, elem string) []string {
	for i := 0; i < len(arr); i++ {
		if arr[i] == elem {
			// Remove i-th element
			arr = arr[:i+copy(arr[i:], arr[i+1:])]
		}
	}
	return arr
}

func strArrAdd(arr []string, elem string) []string {
	// Search if elem is also present in the slice
	for i := 0; i < len(arr); i++ {
		if arr[i] == elem {
			return arr
		}
	}
	return append(arr, elem)
}

func (self *BBoxTagFilterPlugin) applyFilterItem(coord *gopnik.TileCoord, fi *filterItem, inBb bool) {
	if inBb {
		for _, tag := range fi.Drop {
			coord.Tags = strArrDrop(coord.Tags, tag)
		}
		for _, tag := range fi.Add {
			coord.Tags = strArrAdd(coord.Tags, tag)
		}
	} else {
		for _, tag := range fi.DropOtherwise {
			coord.Tags = strArrDrop(coord.Tags, tag)
		}
		for _, tag := range fi.AddOtherwise {
			coord.Tags = strArrAdd(coord.Tags, tag)
		}
	}
}

func (self *BBoxTagFilterPlugin) Filter(coord gopnik.TileCoord) (gopnik.TileCoord, error) {
L:
	for _, rule := range self.rules {
		// Check bboxes
		for _, bb := range rule.BBoxes {
			if bb.Crosses(coord) {
				self.applyFilterItem(&coord, &rule, true)
				continue L
			}
		}
		self.applyFilterItem(&coord, &rule, false)
	}
	return coord, nil
}

func (self *BBoxTagFilterPlugin) Configure(cfg json.RawMessage) error {
	var config bBoxTagFilterPluginConfig
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}
	self.rules = config.Rules
	return nil
}

type BBoxTagFilterPluginFactory struct {
}

func (self *BBoxTagFilterPluginFactory) Name() string {
	return "BBoxTagFilterPlugin"
}

func (self *BBoxTagFilterPluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	res := new(BBoxTagFilterPlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(BBoxTagFilterPluginFactory))
}
