package bettertester

import (
	"fmt"

	pb "github.com/wrossmorrow/bettertester/gen/proto/go/config/v1"
)

type Constants struct {
	Values map[string]string
}

func NewConstants() Constants {
	return Constants{Values: make(map[string]string)}
}

func NewConstantsFromProtobufConfig(p *pb.Config) (*Constants, error) {
	cs := NewConstants()
	for _, c := range p.GetConstants() {
		if c.GetName() == "" {
			return nil, fmt.Errorf("empty constant name")
		}
		if cs.Exists(c.GetName()) {
			return nil, fmt.Errorf("duplicate constant name: %s", c.GetName())
		}
		cs.Set(c.GetName(), c.GetValue())
	}
	return &cs, nil
}

func (c *Constants) IsEmpty() bool {
	return len(c.Values) == 0
}

func (c *Constants) Exists(name string) bool {
	_, ok := c.Values[name]
	return ok
}

func (c *Constants) Get(name string) string {
	return c.Values[name]
}

func (c *Constants) Set(name, value string) {
	c.Values[name] = value
}
