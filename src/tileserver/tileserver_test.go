package tileserver

/*
import (
	_ "defplugins/testcache"
	json "github.com/orofarne/strict-json"
	"fmt"
	"imgsplitter"
	"interfaces"
	"io/ioutil"
	"net/http"
	"plugin"
	"sampledata"
	"testing"
	"time"
)

type Config struct {
	CachePlugin string                     //
	Plugins     map[string]json.RawMessage //
}

func TestSimple(t *testing.T) {
	addr := "127.0.0.1:5341"
	cfg := Config{

		CachePlugin: "TestCachePlugin",
		Plugins:     map[string]json.RawMessage{"TestCachePlugin": {}},
	}
	var cp interfaces.CachePluginInterface
	var ok bool

	imgsplitter.SetMetaSize(2)

	if plug, err := plugin.DefaultPluginStore.New(cfg.CachePlugin, cfg.Plugins[cfg.CachePlugin]); err == nil {
		if cp, ok = plug.(interfaces.CachePluginInterface); cp == nil || !ok {
			log.Fatalf("Invalid plugin '%s', '%v', '%v'", cfg.CachePlugin, cp, ok)
		}
	} else {
		log.Fatalf("Failed to create plugin '%s': %v", cfg.CachePlugin, err)
	}

	ts, err := NewTileServer("gopnikslave", sampledata.Stylesheet, 1, 1, cp, 0)
	if err != nil {
		t.Errorf("NewTilesServer error: %v", err)
	}
	s := &http.Server{
		Addr:           addr,
		Handler:        ts,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go s.ListenAndServe()

	resp, err := http.Get(fmt.Sprintf("http://%s/1/0/0.png", addr))
	if err != nil {
		t.Errorf("http.Get error: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("ioutil.ReadAll: %v", err)
	}

	sampledata.CheckTile(t, body, sampledata.TileImage)
}
*/
