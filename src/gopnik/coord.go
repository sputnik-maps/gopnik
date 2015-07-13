package gopnik

type TileCoord struct {
	X    uint64
	Y    uint64
	Zoom uint64
	Size uint64
	Tags []string
}

func (tc *TileCoord) Equals(coord *TileCoord) bool {
	if tc.Zoom == coord.Zoom && tc.X == coord.X && tc.Y == coord.Y && tc.Size == coord.Size && len(tc.Tags) == len(coord.Tags) {
		for _, tag := range tc.Tags {
			for _, tag2 := range coord.Tags {
				if tag2 != tag {
					return false
				}
			}
		}
		return true
	} else {
		return false
	}
}
