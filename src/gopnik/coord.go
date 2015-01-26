package gopnik

type TileCoord struct {
	X    uint64
	Y    uint64
	Zoom uint64
	Size uint64
	Tags []string
}

func (tc *TileCoord) Equals(coord *TileCoord) bool {
	if tc.Zoom == coord.Zoom && tc.X == coord.X && tc.Y == coord.Y && tc.Size == coord.Size {
		return true
	} else {
		return false
	}
}
