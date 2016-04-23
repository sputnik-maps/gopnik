package filecache

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"gopnik"
	"plugins"

	json "github.com/orofarne/strict-json"
	"bytes"
	"strings"
)

type simpleFileCacheConfig struct {
	Root string
	FileName string
}

type SimpleFileCachePlugin struct {
	root string
	fileNameTemplate *template.Template
}

func (self *SimpleFileCachePlugin) Configure(cfg json.RawMessage) error {
	conf := simpleFileCacheConfig{
		Root: "/tmp/tiles",
		FileName: "{{.Zoom}}_{{.X}}_{{.Y}}",
	}
	err := json.Unmarshal(cfg, &conf)
	if err != nil {
		return err
	}
	self.root = conf.Root
	self.fileNameTemplate, err = template.New("fileName").Parse(conf.FileName)

	if err != nil {
		return err
	}

	err = os.MkdirAll(self.root, 0777)
	return err
}

func (self *SimpleFileCachePlugin) getTileFName(coord gopnik.TileCoord) (string, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v/", self.root)
	err := self.fileNameTemplate.Execute(&buf, coord)
	buf.WriteString(".png")
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (self *SimpleFileCachePlugin) Get(coord gopnik.TileCoord) ([]byte, error) {
	fileName, err := self.getTileFName(coord)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	return data, err
}

func (self *SimpleFileCachePlugin) Set(coord gopnik.TileCoord, tiles []gopnik.Tile) error {
	for i, tile := range tiles {
		c := gopnik.TileCoord{
			X:    coord.X + (uint64(i) % coord.Size),
			Y:    coord.Y + (uint64(i) / coord.Size),
			Zoom: coord.Zoom,
			Size: coord.Size,
		}
		fileName, err := self.getTileFName(c)
		if err != nil {
			return err
		}

		err = os.MkdirAll(fileName[:strings.LastIndex(fileName, "/")], 0666)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(fileName, tile.Image, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

type SimpleFileCachePluginFactory struct {
}

func (self *SimpleFileCachePluginFactory) Name() string {
	return "SimpleFileCachePlugin"
}

func (self *SimpleFileCachePluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(SimpleFileCachePlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(SimpleFileCachePluginFactory))
}
