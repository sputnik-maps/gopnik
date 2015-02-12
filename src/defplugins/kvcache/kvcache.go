package couchcache

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"

	"app"
	"gopnik"
	"plugins"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

type u8Color [3]uint8

func (col u8Color) RGBA() (r, g, b, a uint32) {
	r = uint32(col[0])
	r |= r << 8
	g = uint32(col[1])
	g |= g << 8
	b = uint32(col[2])
	b |= b << 8
	a = 0xffff
	return
}

var colorBlack = u8Color{0, 0, 0}

type kvstoreCachePluginConf struct {
	Backend             app.PluginConfig
	UseMultilevel       bool
	UseSecondLevelCache bool
	Prefix              string
}

type KVStorePlugin struct {
	config  kvstoreCachePluginConf
	store   gopnik.KVStore
	cache2L map[u8Color][]byte
}

func (self *KVStorePlugin) Configure(cfg json.RawMessage) error {
	err := json.Unmarshal(cfg, &self.config)
	if err != nil {
		return err
	}

	plug, err := plugins.DefaultPluginStore.Create(
		self.config.Backend.Plugin, self.config.Backend.PluginConfig)
	if err != nil {
		return err
	}
	var ok bool
	self.store, ok = plug.(gopnik.KVStore)
	if !ok {
		return fmt.Errorf("Invalid KV plugin")
	}

	if self.config.UseSecondLevelCache {
		self.cache2L = make(map[u8Color][]byte)
	}

	return nil
}

func (self *KVStorePlugin) key(coord gopnik.TileCoord, level int) string {
	if self.config.Prefix == "" {
		return fmt.Sprintf("%v:%v:%v:%v:%v",
			level, coord.Size, coord.Zoom, coord.X, coord.Y)
	} else {
		return fmt.Sprintf("%v:%v:%v:%v:%v:%v",
			self.config.Prefix, level, coord.Size, coord.Zoom, coord.X, coord.Y)
	}
}

func (self *KVStorePlugin) parseSecondLevel(metacoord, coord gopnik.TileCoord, data []byte) ([]byte, error) {
	var col u8Color
	colorSize := uint64(binary.Size(col))
	index := (coord.Y-metacoord.Y)*metacoord.Size + (coord.X - metacoord.X)
	offset := colorSize * index
	buf := bytes.NewReader(data)
	_, err := buf.Seek(int64(offset), 0)
	if err != nil {
		return nil, fmt.Errorf("Invalid second level cache: %v", err)
	}
	err = binary.Read(buf, binary.BigEndian, &col)
	if err != nil {
		return nil, fmt.Errorf("Invalid second level cache (invalid color): %v", err)
	}
	if col[0] == colorBlack[0] && col[1] == colorBlack[1] && col[2] == colorBlack[2] {
		// Empty!
		return nil, nil
	}

	// Check cache
	if self.config.UseSecondLevelCache {
		img := self.cache2L[col]
		if img != nil {
			return img, nil
		}
	}

	// Generate image
	bounds := image.Rect(0, 0, 256, 256)
	img := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.Set(x, y, col)
		}
	}
	outbuf := bytes.NewBuffer(nil)
	err = png.Encode(outbuf, img)
	if err != nil {
		return nil, err
	}
	imgData := outbuf.Bytes()

	// Save image to cache
	if self.config.UseSecondLevelCache {
		self.cache2L[col] = imgData
	}

	return imgData, nil
}

func (self *KVStorePlugin) Get(coord gopnik.TileCoord) ([]byte, error) {
	// Request from kvstore
	N := 1
	if self.config.UseMultilevel {
		N = 2
	}

	for i := 1; i <= N; i++ {
		var k_coord gopnik.TileCoord
		switch i {
		case 2:
			k_coord = app.App.Metatiler().TileToMetatile(&coord)
		default:
			k_coord = coord
		}
		key := self.key(k_coord, i)
		log.Debug("Request tile by key '%v' from kvstore...", key)
		data, err := self.store.Get(key)
		if data == nil {
			//Key not found
			continue
		}
		if err != nil {
			return nil, err
		}
		log.Debug("Key '%v' [level=%v] found: %v bytes", key, i, len(data))
		if data != nil {
			switch i {
			case 2:
				return self.parseSecondLevel(k_coord, coord, data)
			default:
				return data, err
			}
		}
	}
	return nil, nil
}

func (self *KVStorePlugin) setData(coord gopnik.TileCoord, data []byte, level int) error {
	tileKey := self.key(coord, level)
	return self.store.Set(tileKey, data)
}

func (self *KVStorePlugin) setSecondLevelData(coord gopnik.TileCoord, tiles []gopnik.Tile) error {
	buf := bytes.NewBuffer(nil)
	for _, elem := range tiles {
		if elem.SingleColor != nil {
			r, g, b, _ := elem.SingleColor.RGBA()
			col := u8Color{uint8(r), uint8(g), uint8(b)}
			err := binary.Write(buf, binary.BigEndian, &col)
			if err != nil {
				return err
			}
		} else {
			err := binary.Write(buf, binary.BigEndian, colorBlack)
			if err != nil {
				return err
			}
		}
	}
	return self.setData(coord, buf.Bytes(), 2)
}

func (self *KVStorePlugin) Set(coord gopnik.TileCoord, tiles []gopnik.Tile) error {
	var err error

	c := coord
	c.Size = 1

	for c.Y < coord.Y+coord.Size {
		c.X = coord.X
		for c.X < coord.X+coord.Size {
			index := int((c.Y-coord.Y)*coord.Size + (c.X - coord.X))
			if !self.config.UseMultilevel || tiles[index].SingleColor == nil {
				if err = self.setData(c, tiles[index].Image, 1); err != nil {
					return err
				}
			}
			c.X++
		}
		c.Y++
	}

	// Save secondLevel
	for _, elem := range tiles {
		if elem.SingleColor != nil {
			err = self.setSecondLevelData(coord, tiles)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

type KVStorePluginFactory struct {
}

func (cpf *KVStorePluginFactory) Name() string {
	return "KVStorePlugin"
}

func (cpf *KVStorePluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	var res = new(KVStorePlugin)
	err := res.Configure(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func init() {
	plugins.DefaultPluginStore.AddPlugin(new(KVStorePluginFactory))
}
