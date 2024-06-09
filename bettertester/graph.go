package bettertester

import (
	"fmt"
	"os"
	"strings"

	pb "github.com/wrossmorrow/bettertester/gen/proto/go/config/v1"
	"google.golang.org/protobuf/proto"
)

type Executable interface {
	Execute(ctx *ExecutionContext) error
}

type CallGraph struct {
	Calls map[string]*Call
}

func NewCallGraph() CallGraph {
	return CallGraph{
		Calls: make(map[string]*Call),
	}
}

func NewCallGraphFromConfigFile(filename string) *CallGraph {
	cfg, err := ReadConfig(filename)
	if err != nil {
		panic(err)
	}
	return NewCallGraphFromProtobufConfig(cfg)
}

func NewCallGraphFromProtobufConfig(cfg *pb.Config) *CallGraph {
	defaults := cfg.GetDefaults()
	constants, err := NewConstantsFromProtobufConfig(cfg)
	if err != nil {
		panic(err)
	}
	callgraph := NewCallGraph()
	stored := NewStoredReferences()
	for _, c := range cfg.GetCalls() {
		d := proto.Clone(defaults).(*pb.CallSpec)
		proto.Merge(d, c)
		call := NewCallFromProtobuf(d)
		for _, n := range call.Store {
			err = stored.AddNewReference(n, c.Name)
			if err != nil {
				panic(err)
			}
		}
		callgraph.AddCall(&call)
	}

	err = callgraph.ResolveReferences(constants, stored)
	if err != nil {
		panic(err)
	}

	return &callgraph
}

func (cg *CallGraph) AddCall(c *Call) error {
	_, ok := cg.Calls[c.Name]
	if ok {
		return fmt.Errorf("Call \"%s\" already exists", c.Name)
	}
	cg.Calls[c.Name] = c
	return nil
}

func (cg *CallGraph) AddNewCallNamed(name string) (*Call, error) {
	_, ok := cg.Calls[name]
	if ok {
		return nil, fmt.Errorf("Call \"%s\" already exists", name)
	}
	c := NewCall(name)
	cg.Calls[name] = &c
	return &c, nil
}

func (cg *CallGraph) AddNewCallNamedWithRequestProto(name string, p *RequestProto) (*Call, error) {
	_, ok := cg.Calls[name]
	if ok {
		return nil, fmt.Errorf("Call \"%s\" already exists", name)
	}
	c := NewCallWithRequestProto(name, p)
	cg.Calls[name] = &c
	return &c, nil
}

func (cg *CallGraph) GetCall(name string) (*Call, error) {
	c, ok := cg.Calls[name]
	if !ok {
		return nil, fmt.Errorf("call \"%s\" does not exist", name)
	}
	return c, nil
}

func (cg *CallGraph) AddCallAfter(name, after string) error {
	a, err := cg.GetCall(after)
	if err != nil {
		return err
	}
	c, err := cg.AddNewCallNamed(name)
	if err != nil {
		return err
	}
	c.IsAfter(a)
	return nil
}

func (cg *CallGraph) AddCallBefore(name, before string) error {
	b, err := cg.GetCall(before)
	if err != nil {
		return err
	}
	c, err := cg.AddNewCallNamed(name)
	if err != nil {
		return err
	}
	c.IsBefore(b)
	return nil
}

func (cg *CallGraph) ResolveReferences(constants *Constants, stored *StoredReferences) error {
	for _, c := range cg.Calls {
		r := c.GetReferences()
		for _, ref := range r.Flatten() {
			if strings.HasPrefix(ref, "env:") {
				e := strings.TrimPrefix(ref, "env:")
				_, ok := os.LookupEnv(e)
				if !ok {
					return fmt.Errorf("referenced environment variable \"%s\" not present", e)
				}
			} else if strings.HasPrefix(ref, "const:") {
				r := strings.TrimPrefix(ref, "const:")
				if !constants.Exists(r) {
					return fmt.Errorf("constant reference \"%s\" not found", r)
				}
			} else if strings.HasPrefix(ref, "stored:") {
				r := strings.TrimPrefix(ref, "stored:")
				b, ok := stored.GetReference(r)
				if !ok {
					return fmt.Errorf("stored reference \"%s\" not found", r)
				}
				c.IsAfter(cg.Calls[b])
			} else if strings.HasPrefix(ref, "call:") {
				r := strings.TrimPrefix(ref, "call:")
				n := strings.Split(r, ".")[0] // valid?
				_, ok := cg.Calls[n]
				if !ok {
					return fmt.Errorf("call reference \"%s\" not found", n)
				}
				c.IsAfter(cg.Calls[n])
			} else {
				return fmt.Errorf("unknown reference type \"%s\"", ref)
			}

		}
	}
	return nil
}

func (cg *CallGraph) AddCallEdge(before, after string) error {
	b, err := cg.GetCall(before)
	if err != nil {
		return err
	}
	a, err := cg.GetCall(after)
	if err != nil {
		return err
	}
	a.IsAfter(b)
	return nil
}

func (cg *CallGraph) Len() int {
	return len(cg.Calls)
}

func (cg *CallGraph) Print() {
	for _, c := range cg.Calls {
		fmt.Printf("%s\n", c.Printable())
	}
}

func (cg *CallGraph) Execute(ctx *ExecutionContext) error {
	C := NewStack()             // call stack, for calls that are ready to be executed
	F := make(map[string]*Call) // named list of calls that are finished executing
	for _, c := range cg.Calls {
		if c.After.IsEmpty() {
			C.Push(c)
		}
	}
	if C.Len() == 0 {
		return fmt.Errorf("no start point found")
	}
	ctx.Start() // allocate a new context/results holder basically
	for C.Len() > 0 {
		c := C.Pop().(*Call)
		n := c.Name
		r := ctx.Execute(c)
		fmt.Printf("Executed %s (%s) %d\n", n, c.Printable(), r.Response.Status)
		F[n] = c
		if r.IsError() {
			return r.GetError()
		}
		for _, nb := range c.Preceeds() {
			b := cg.Calls[nb]
			ready := true
			for _, a := range b.Requires() {
				if _, ok := F[a]; !ok {
					ready = false
				}
			}
			if ready {
				C.Push(b)
			}
		}
	}
	if len(F) < cg.Len() {
		m := make([]string, 0)
		for _, c := range cg.Calls {
			if _, ok := F[c.Name]; !ok {
				m = append(m, c.Name)
			}
		}
		return fmt.Errorf("incomplete call graph; did not reach %v", m)
	}
	return nil
}
