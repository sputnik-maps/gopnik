package gopnikrpcutils

import (
	"gopnik"
	"gopnikrpc/types"
)

func CoordToRPC(coord *gopnik.TileCoord) *types.Coord {
	res := &types.Coord{
		X:    int64(coord.X),
		Y:    int64(coord.Y),
		Zoom: int16(coord.Zoom),
		Size: int16(coord.Size),
	}

	for _, tag := range coord.Tags {
		res.Tags[tag] = true
	}

	return res
}

func CoordFromRPC(coord *types.Coord) *gopnik.TileCoord {
	res := &gopnik.TileCoord{
		Zoom: uint64(coord.Zoom),
		X:    uint64(coord.X),
		Y:    uint64(coord.Y),
		Size: uint64(coord.Size),
	}

	for tag, _ := range coord.Tags {
		res.Tags = append(res.Tags, tag)
	}

	return res
}

func TileToRPC(tile *gopnik.Tile) *types.Tile {
	res := &types.Tile{
		Image: tile.Image,
	}

	if tile.SingleColor != nil {
		r, g, b, a := tile.SingleColor.RGBA()
		res.Color = &types.Color{
			R: int32(r),
			G: int32(g),
			B: int32(b),
			A: int32(a),
		}
	}

	return res
}
