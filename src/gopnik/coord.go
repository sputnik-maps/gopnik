package gopnik

type TileCoord struct {
	X    uint64
	Y    uint64
	Zoom uint64
	Size uint64
	Tags []string
}

func (self *TileCoord) Equals(coord *TileCoord) bool {
	if self.Zoom == coord.Zoom && self.X == coord.X && self.Y == coord.Y && self.Size == coord.Size && len(self.Tags) == len(coord.Tags) {
		TL:
			for _, tag := range coord.Tags {
				for _, tag2 := range self.Tags {
					if tag2 == tag {
						continue TL
					}
				}
				return false
			}
		return true
	} else {
		return false
	}
}
