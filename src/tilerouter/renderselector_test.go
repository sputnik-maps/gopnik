package tilerouter

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"gopnik"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type RenderSelectorSuite struct{}

var _ = Suite(&RenderSelectorSuite{})

type fakeRender struct {
}

func (srv *fakeRender) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.String(), "/status") {
		w.Write([]byte{'O', 'k'})
		return
	}
	http.Error(w, "Invalid request", 400)
}

func runFakeRender(addr string) {
	var render fakeRender
	s := &http.Server{
		Addr:           addr,
		Handler:        &render,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

func (s *RenderSelectorSuite) TestRoute(c *C) {
	renders := []string{"localhost:9001", "localhost:9002", "localhost:9003"}

	for _, r := range renders {
		go runFakeRender(r)
	}
	time.Sleep(time.Millisecond)

	rs, err := NewRenderSelector(renders, time.Second, 30*time.Second)
	c.Assert(err, IsNil)
	defer rs.Stop()

	coord := gopnik.TileCoord{
		X:    0,
		Y:    0,
		Zoom: 1,
		Size: 1,
	}
	back1 := rs.SelectRender(coord)
	coord.X = 3
	back2 := rs.SelectRender(coord)
	coord.Y = 4
	back3 := rs.SelectRender(coord)
	coord.Zoom = 5
	back4 := rs.SelectRender(coord)

	c.Check(
		back1 != back2 || back1 != back3 || back1 != back4,
		Equals, true)
}
