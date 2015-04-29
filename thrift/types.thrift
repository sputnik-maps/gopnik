namespace go gopnikrpc.types
namespace cpp gopnikrpc.types

struct Coord {
	1: i64 x
	2: i64 y
	3: i16 zoom
	4: i16 size
	5: set<string> tags
}

struct Color {
	1: i32 r
	2: i32 g
	3: i32 b
	4: i32 a
}

struct Tile {
	1: optional binary image
	2: optional Color color
}
