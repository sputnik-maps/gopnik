package tilerouter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"gopnik"
)

var ATTEMPTS = 2

type TileRouter struct {
	timeout     time.Duration
	pingPeriod  time.Duration
	client      http.Client
	rendersLock sync.RWMutex
	selector    *RenderSelector
}

func NewTileRouter(renders []string, timeout time.Duration, pingPeriod time.Duration) (*TileRouter, error) {
	tr := &TileRouter{
		timeout:    timeout,
		pingPeriod: pingPeriod,
		client: http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: timeout,
			},
		},
		rendersLock: sync.RWMutex{},
	}

	tr.UpdateRenders(renders)

	return tr, nil
}

func (tr *TileRouter) UpdateRenders(renders []string) {
	selector, err := NewRenderSelector(renders, tr.pingPeriod, tr.timeout)
	if err != nil {
		log.Error("Failed to recreate RenderSelector: %v", err)
	}

	tr.rendersLock.Lock()
	defer tr.rendersLock.Unlock()
	if tr.selector != nil {
		tr.selector.Stop()
	}
	tr.selector = selector
}

func (tr *TileRouter) Tile(coord gopnik.TileCoord) (img []byte, err error) {
	for i := 0; i < ATTEMPTS; i++ {
		addr := tr.selector.SelectRender(coord)
		if addr == "" {
			img, err = nil, fmt.Errorf("No available renders")
			time.Sleep(10 * time.Second)
			continue
		}
		renderUrl := fmt.Sprintf("http://%s/%v/%v/%v.png",
			addr, coord.Zoom, coord.X, coord.Y)
		for i, tag := range coord.Tags {
			if i == 0 {
				renderUrl += "?"
			} else {
				renderUrl += "&"
			}
			renderUrl += "tag="
			renderUrl += url.QueryEscape(tag)
		}
		resp, er := tr.client.Get(renderUrl)
		if er != nil {
			tr.selector.SetStatus(addr, Offline)
			img, err = nil, fmt.Errorf("HTTP GET error: %v", er)
			continue
		}
		defer resp.Body.Close()
		img, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			tr.selector.SetStatus(addr, Offline)
			img, err = nil, fmt.Errorf("HTTP read error: %v", err)
			continue
		}
		return
	}

	return
}
