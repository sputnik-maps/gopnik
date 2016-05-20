package filecache_test

import (
	"defplugins/simplefilecache"
	"fmt"
	"gopnik"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type fileNameTestCase struct {
	template      string
	expectedFile  string
	expectedError bool
}

var (
	fileNameTests = []fileNameTestCase{
		{
			template:     "{{.Zoom}}_{{.X}}_{{.Y}}",
			expectedFile: "3_1_2.png",
		},
		{
			template:     "{{.Zoom}}/{{.X}}-{{.Y}}",
			expectedFile: "3/1-2.png",
		},
		{
			template:     "{{.Zoom}}/{{.X}}/{{.Y}}",
			expectedFile: "3/1/2.png",
		},
		{
			template:     "{{.Size}}",
			expectedFile: "4.png",
		},
		{
			template:      "{{.NoSuchVar}}",
			expectedFile:  "",
			expectedError: true,
		},
	}
)

//Tests that cache can use different layouts
//when generating tiles.
func TestFileName(t *testing.T) {
	dirs := make([]string, 0)
	defer cleanup(dirs)
	for _, test := range fileNameTests {
		tilesCache, err := ioutil.TempDir("", "tiles")
		if err != nil {
			t.Fatalf("Error configuring cache: %s", err)
		}

		dirs = append(dirs, tilesCache)

		tilesCache = strings.Replace(tilesCache, "\\", "\\\\", -1)
		conf := fmt.Sprintf(`{
		"root": "%s",
		"fileName": "%s"
		}`, tilesCache, test.template)
		cache := &filecache.SimpleFileCachePlugin{}
		err = cache.Configure([]byte(conf))
		if err != nil {
			t.Fatal(err)
		}

		err = cache.Set(gopnik.TileCoord{
			X:    1,
			Y:    2,
			Zoom: 3,
			Size: 4,
		}, []gopnik.Tile{{Image: []byte("123")}})
		if err != nil && !test.expectedError {
			t.Fatal(err)
		}
		if err == nil && test.expectedError {
			t.Fatal("Expected error when generated tiles")
		}

		if !test.expectedError {
			data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", tilesCache, test.expectedFile))
			if err != nil {
				t.Fatal(err)
			}

			if string(data) != "123" {
				t.Fatalf("Expected 123 in file but got %v", data)
			}
		}
	}
}

func cleanup(dirs []string) error {
	var err error
	for _, d := range dirs {
		dErr := os.RemoveAll(d)
		if dErr != nil {
			err = dErr
		}
	}
	return err
}
