package clusterconfig

import (
	"testing"

	json "github.com/orofarne/strict-json"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type ClusterSuite struct{}

var _ = Suite(&ClusterSuite{})

func strToRawMessage(str string) (res json.RawMessage) {
	err := json.Unmarshal([]byte(str), &res)
	if err != nil {
		panic(err)
	}
	return
}

func (s *ClusterSuite) TestConfigure(c *C) {
	cc := new(ClusterConfigPlugin)
	cfg := strToRawMessage(`{"Nodes": ["localhost:8090", "www.openstreetmap.org"]}`)

	err := cc.Configure(cfg)
	c.Assert(err, IsNil)

	renders, err := cc.GetRenders()
	c.Assert(err, IsNil)
	c.Check(renders, DeepEquals, []string{"localhost:8090", "www.openstreetmap.org"})
}
