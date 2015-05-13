package tileurl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	json "github.com/orofarne/strict-json"

	"gopnik"
	"plugins"
)

type tileUrlPluginConf struct {
	Url string // Can contains {x}, {y}, {z}
}

type TileUrlPlugin struct {
	url string
}

func (self *TileUrlPlugin) Configure(cfg json.RawMessage) error {
	config := tileUrlPluginConf{}

	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}

	self.url = config.Url

	return nil
}

func (self *TileUrlPlugin) getUrl(coord gopnik.TileCoord) (url string) {
	url = self.url
	url = strings.Replace(url, "{x}", strconv.FormatUint(coord.X, 10), -1)
	url = strings.Replace(url, "{y}", strconv.FormatUint(coord.Y, 10), -1)
	url = strings.Replace(url, "{z}", strconv.FormatUint(coord.Zoom, 10), -1)
	return
}

func (self *TileUrlPlugin) Get(coord gopnik.TileCoord) ([]byte, error) {
	url := self.getUrl(coord)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "image/png" {
		return nil, fmt.Errorf("Invalid Content-Type: '%v'", contentType)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (self *TileUrlPlugin) Set(gopnik.TileCoord, []gopnik.Tile) error {
	return errors.New("Set operation is not allowed")
}

type TileUrlPluginFactory struct {
}

func (self *TileUrlPluginFactory) Name() string {
	return "TileUrlPlugin"
}

func (self *TileUrlPluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(TileUrlPlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(TileUrlPluginFactory))
}
