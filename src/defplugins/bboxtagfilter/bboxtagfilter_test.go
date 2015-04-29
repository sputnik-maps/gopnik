package bboxtagfilter

import (
	"testing"

	"gopnik"

	json "github.com/orofarne/strict-json"
	"github.com/stretchr/testify/require"
)

const TEST_CONFIG_1 = `
{
	"Rules": [
		{
			"BBoxes": [
				{
					"MinZoom": 1,
					"MaxZoom": 19,
					"MaxLat": 56.972679,
					"MaxLon": 40.321198,
					"MinLat": 54.190566,
					"MinLon": 35.094452,
					"#City": "Москва"
				 }
			],
			"Add": ["tag1"],
			"DropOtherwise": ["tag2"]
		}
	]
}
`

var TEST_COORD_INSIDE_MOSCOW = gopnik.TileCoord{
	X:    39614,
	Y:    20483,
	Zoom: 16,
	Size: 1,
}

var TEST_COORD_OUTSIDE_OF_MOSCOW = gopnik.TileCoord{
	X:    40212,
	Y:    20421,
	Zoom: 16,
	Size: 1,
}

var TEST_COORD_CONTAINS_MOSCOW = gopnik.TileCoord{
	X:    16,
	Y:    8,
	Zoom: 5,
	Size: 8,
}

func createTestBBoxTagFilter(config string) gopnik.FilterPluginInterface {

	factory := BBoxTagFilterPluginFactory{}
	filterI, err := factory.New(json.RawMessage(config))
	if err != nil {
		panic(err)
	}
	return filterI.(gopnik.FilterPluginInterface)
}

func Test_BBox_Tag_Filter_Can_Be_Cteated(t *testing.T) {
	config := `{ "Rules": [] }`
	filter := createTestBBoxTagFilter(config)

	require.NotNil(t, filter)
}

func Test_BBox_Tag_Filter_Should_Add_Tag_To_Inner_Metatile(t *testing.T) {
	filter := createTestBBoxTagFilter(TEST_CONFIG_1)

	bbox, err := filter.Filter(TEST_COORD_INSIDE_MOSCOW)

	require.Nil(t, err)
	require.Equal(t, len(bbox.Tags), 1)
	require.Equal(t, bbox.Tags[0], "tag1")
}

func Test_BBox_Tag_Filter_Should_Add_Tag_To_Metatile_That_Contains_BBox(t *testing.T) {
	filter := createTestBBoxTagFilter(TEST_CONFIG_1)

	bbox, err := filter.Filter(TEST_COORD_CONTAINS_MOSCOW)

	require.Nil(t, err)
	require.Equal(t, len(bbox.Tags), 1)
	require.Equal(t, bbox.Tags[0], "tag1")
}

func Test_BBox_Tag_Filter_Should_Not_Add_Tag_To_Outer_Metatile(t *testing.T) {
	filter := createTestBBoxTagFilter(TEST_CONFIG_1)

	bbox, err := filter.Filter(TEST_COORD_OUTSIDE_OF_MOSCOW)

	require.Nil(t, err)
	require.Equal(t, len(bbox.Tags), 0)
}

func Test_BBox_Tag_Filter_Should_Drop_Tag_From_Outer_Metatile(t *testing.T) {
	filter := createTestBBoxTagFilter(TEST_CONFIG_1)
	coord := gopnik.TileCoord{
		X:    TEST_COORD_OUTSIDE_OF_MOSCOW.X,
		Y:    TEST_COORD_OUTSIDE_OF_MOSCOW.Y,
		Zoom: TEST_COORD_OUTSIDE_OF_MOSCOW.Zoom,
		Size: TEST_COORD_OUTSIDE_OF_MOSCOW.Size,
		Tags: []string{"tag2"},
	}

	bbox, err := filter.Filter(coord)

	require.Nil(t, err)
	require.Equal(t, len(bbox.Tags), 0)
}

func Test_BBox_Tag_Filter_Should_Not_Drop_Tag_From_Outer_Metatile(t *testing.T) {
	filter := createTestBBoxTagFilter(TEST_CONFIG_1)
	coord := gopnik.TileCoord{
		X:    TEST_COORD_OUTSIDE_OF_MOSCOW.X,
		Y:    TEST_COORD_OUTSIDE_OF_MOSCOW.Y,
		Zoom: TEST_COORD_OUTSIDE_OF_MOSCOW.Zoom,
		Size: TEST_COORD_OUTSIDE_OF_MOSCOW.Size,
		Tags: []string{"tag3"},
	}

	bbox, err := filter.Filter(coord)

	require.Nil(t, err)
	require.Equal(t, len(bbox.Tags), 1)
	require.Equal(t, bbox.Tags[0], "tag3")
}
