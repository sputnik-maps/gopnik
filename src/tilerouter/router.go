	"io/ioutil"
var ATTEMPTS = 2

	for i := 0; i < ATTEMPTS; i++ {
		addr := tr.selector.SelectRender(coord)
		if addr == "" {
			img, err = nil, fmt.Errorf("No available renders")
			time.Sleep(10 * time.Second)
			continue
		}
		renderUrl := fmt.Sprintf("http://%s/%v/%v/%v.png",
			addr, coord.Zoom, coord.X, coord.Y)
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

	return
