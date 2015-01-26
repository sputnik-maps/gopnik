package gopnik

type Metatiler struct {
	metaSize uint64
	tileSize uint64
}

func NewMetatiler(metaSize uint64, tileSize uint64) *Metatiler {
	return &Metatiler{
		metaSize: metaSize,
		tileSize: tileSize,
	}
}

func (self *Metatiler) TileToMetatile(coord *TileCoord) TileCoord {
	mSize := self.MetaSize(coord.Zoom)
	mask := mSize - 1
	return TileCoord{
		X:    coord.X & ^mask,
		Y:    coord.Y & ^mask,
		Zoom: coord.Zoom,
		Size: mSize,
		Tags: coord.Tags,
	}
}

func (self *Metatiler) MetaSize(zoom uint64) uint64 {
	nT := uint64(1) << zoom
	if nT < self.metaSize {
		return nT
	} else {
		return self.metaSize
	}
}

func (self *Metatiler) TileSize() uint64 {
	return self.tileSize
}

func (self *Metatiler) NTiles(zoom uint64) uint64 {
	mS := self.MetaSize(zoom)
	return mS * mS
}
