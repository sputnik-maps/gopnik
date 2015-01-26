package plugins

import (
	"testing"

	. "gopkg.in/check.v1"

	json "github.com/orofarne/strict-json"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type PluginSuite struct{}

var _ = Suite(&PluginSuite{})

type FakePlugin struct {
	X int
}

type FakePluginFactory struct {
}

func (f *FakePluginFactory) Name() string {
	return "fake_plugin"
}

func (f *FakePluginFactory) New(cfg json.RawMessage) (interface{}, error) {
	res := new(FakePlugin)
	err := json.Unmarshal(cfg, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *PluginSuite) TestPluginStore(c *C) {
	ps := NewPluginStore()
	err := ps.AddPlugin(new(FakePluginFactory))
	c.Assert(err, IsNil)
	p, err := ps.Create("fake_plugin", []byte(`{"X": 10}`))
	c.Assert(err, IsNil)
	c.Assert(p, NotNil)
	fp, ok := p.(*FakePlugin)
	c.Assert(ok, Equals, true)
	c.Check(fp.X, Equals, 10)
}
