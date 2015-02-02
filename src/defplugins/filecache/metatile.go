package filecache

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"app"
	"gopnik"
)

/*
#define META_MAGIC "META"
#define META_MAGIC_COMPRESSED "METZ"

struct entry {
	int offset;
	int size;
};

struct meta_layout {
	char magic[4];
	int count; // METATILE ^ 2
	int x, y, z; // lowest x,y of this metatile, plus z
	struct entry index[]; // count entries
	// Followed by the tile data
	// The index offsets are measured from the start of the file
};
*/

const MAXCOUNT = 1000
const MAXENTRYSIZE = 100000

type metaEntry struct {
	Offset int32
	Size   int32
}

type metaLayout struct {
	Magic   []byte
	Count   int32
	X, Y, Z int32
	Index   []metaEntry
}

func encodeHeader(w io.Writer, ml *metaLayout) error {
	endian := binary.LittleEndian
	var err error
	if err = binary.Write(w, endian, ml.Magic); err != nil {
		return err
	}
	if err = binary.Write(w, endian, ml.Count); err != nil {
		return err
	}
	if err = binary.Write(w, endian, ml.X); err != nil {
		return err
	}
	if err = binary.Write(w, endian, ml.Y); err != nil {
		return err
	}
	if err = binary.Write(w, endian, ml.Z); err != nil {
		return err
	}
	for _, ent := range ml.Index {
		if err = binary.Write(w, endian, ent); err != nil {
			return err
		}
	}
	return nil
}

func EncodeMetatile(w io.Writer, coord gopnik.TileCoord, tiles []gopnik.Tile) error {
	mask := int32(coord.Size) - 1
	ml := &metaLayout{
		Magic: []byte{'M', 'E', 'T', 'A'},
		Count: int32(len(tiles)),
		X:     int32(coord.X) & ^mask,
		Y:     int32(coord.Y) & ^mask,
		Z:     int32(coord.Zoom),
	}
	offset := int32(20 + 8*len(tiles)) // sizeof(metaLayout)
	for _, tile := range tiles {
		ml.Index = append(ml.Index, metaEntry{
			Offset: offset,
			Size:   int32(len(tile.Image)),
		})
		offset += int32(len(tile.Image))
	}
	if err := encodeHeader(w, ml); err != nil {
		return nil
	}
	for _, tile := range tiles {
		if _, err := w.Write(tile.Image); err != nil {
			return nil
		}
	}
	return nil
}

func decodeMetatileHeader(r io.Reader) (*metaLayout, error) {
	endian := binary.LittleEndian
	ml := new(metaLayout)

	ml.Magic = make([]byte, 4)
	err := binary.Read(r, endian, &ml.Magic)
	if err != nil {
		return nil, err
	}
	if ml.Magic[0] != 'M' || ml.Magic[1] != 'E' || ml.Magic[2] != 'T' || ml.Magic[3] != 'A' {
		return nil, fmt.Errorf("Invalid Magic field: %v", ml.Magic)
	}

	if err = binary.Read(r, endian, &ml.Count); err != nil {
		return nil, err
	}
	if ml.Count > MAXCOUNT {
		return nil, fmt.Errorf("Count > MAXCOUNT (Count = %v)", ml.Count)
	}

	if err = binary.Read(r, endian, &ml.X); err != nil {
		return nil, err
	}
	if err = binary.Read(r, endian, &ml.Y); err != nil {
		return nil, err
	}
	if err = binary.Read(r, endian, &ml.Z); err != nil {
		return nil, err
	}

	for i := int32(0); i < ml.Count; i++ {
		var entry metaEntry
		if err = binary.Read(r, endian, &entry); err != nil {
			return nil, err
		}
		ml.Index = append(ml.Index, entry)
	}

	return ml, nil
}

func GetRawTileFromMetatile(r io.ReadSeeker, coord gopnik.TileCoord) ([]byte, error) {
	ml, err := decodeMetatileHeader(r)
	if err != nil {
		return nil, err
	}

	size := int32(math.Sqrt(float64(ml.Count)))
	index := (int32(coord.Y)-ml.Y)*size + (int32(coord.X) - ml.X)
	if index >= ml.Count {
		return nil, fmt.Errorf("Invalid index %v/%v", index, ml.Count)
	}
	entry := ml.Index[index]
	if entry.Size > MAXENTRYSIZE {
		return nil, fmt.Errorf("entry size > MAXENTRYSIZE (size: %v)", entry.Size)
	}
	r.Seek(int64(entry.Offset), 0)
	buf := make([]byte, entry.Size)
	l, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	if int32(l) != entry.Size {
		return nil, fmt.Errorf("Invalid tile seze: %v != %v", l, entry.Size)
	}
	return buf, nil
}

func MetatileHashPath(outPath string, coord gopnik.TileCoord) (dirName, fName string) {
	metacoord := app.App.Metatiler().TileToMetatile(&coord)

	hash := make([]byte, 5)
	x := uint32(metacoord.X)
	y := uint32(metacoord.Y)
	// Each meta tile winds up in its own file, with several in each leaf directory
	// the .meta tile name is beasd on the sub-tile at (0,0)

	for i := 0; i < 5; i++ {
		hash[i] = byte(((x & 0x0f) << 4) | (y & 0x0f))
		x >>= 4
		y >>= 4
	}

	dirName = fmt.Sprintf("%s/%v/%v/%v/%v/%v", outPath, metacoord.Zoom, hash[4], hash[3], hash[2], hash[1])
	fName = fmt.Sprintf("%v.meta", hash[0])
	return
}
