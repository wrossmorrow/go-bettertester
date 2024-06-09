package bettertester

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	pb "github.com/wrossmorrow/bettertester/gen/proto/go/config/v1"
)

type Call struct {
	Name       string
	Before     SimpleSet
	After      SimpleSet
	Proto      RequestProto
	Assertions []Assertion
	Store      map[string]string
}

func NewCall(name string) Call {
	return NewCallWithRequestProto(name, DefaultRequestProto())
}

func NewCallWithRequestProto(name string, proto *RequestProto) Call {
	return NewCallWithRequestProtoAndAssertions(name, proto, make([]Assertion, 0))
}

func NewCallWithRequestProtoAndAssertions(name string, proto *RequestProto, assertions []Assertion) Call {
	return Call{
		Name:       name,
		Before:     NewSimpleSet(),
		After:      NewSimpleSet(),
		Proto:      *proto,
		Assertions: assertions,
		Store:      make(map[string]string),
	}
}

func NewCallFromProtobuf(p *pb.CallSpec) Call {
	rp := NewRequestProtoFromProtobuf(p.GetProto())
	at := make([]Assertion, 0)
	if p.GetExpectedStatus() > 0 {
		at = append(at, &StatusAssertion{Expected: int(p.GetExpectedStatus())})
	}
	for _, a := range p.GetAssertions() {
		at = append(at, NewAssertionFromProtobuf(a))
	}
	call := NewCallWithRequestProtoAndAssertions(p.GetName(), rp, at)
	if len(p.GetAfter()) > 0 {
		call.After.AddAll(p.GetAfter())
	}
	if len(p.GetStore()) > 0 {
		for _, v := range p.GetStore() {
			call.Store[v.Path] = v.As
		}
	}
	return call
}

func (c *Call) GetReferences() CallReferences {
	refs := CallReferences{
		Request:    c.Proto.GetReferences(),
		Assertions: make(map[int][]string),
		After:      c.After.AsSlice(),
	}
	for i, a := range c.Assertions {
		r := a.GetReferences()
		if len(r) > 0 {
			refs.Assertions[i] = r
		}
	}
	return refs
}

func (c *Call) IsBefore(d *Call) {
	c.Before.Add(d.Name)
	d.After.Add(c.Name)
}

func (c *Call) IsAfter(d *Call) {
	c.After.Add(d.Name)
	d.Before.Add(c.Name)
}

func (c *Call) Requires() []string {
	return c.After.AsSlice()
}

func (c *Call) Preceeds() []string {
	return c.Before.AsSlice()
}

func (c *Call) Printable() string {
	return fmt.Sprintf("%v -> %s -> %v", c.After.AsSlice(), c.Name, c.Before.AsSlice())
}

func (c *Call) Execute(ctx *ExecutionContext) *CallResult {
	req, err := c.getRequest(ctx)
	if err != nil {
		return NewCallResultFromError(err)
	}
	resp, err := ctx.Client.Do(req)
	return NewCallResultFromRequestAndResponse(req, resp, err)
}

func (c *Call) getRequest(ctx *ExecutionContext) (*http.Request, error) {
	host, err := ctx.ResolveString(c.Proto.Host)
	if err != nil {
		return nil, err
	}
	path, err := ctx.ResolveString(c.Proto.Path)
	if err != nil {
		return nil, err
	}
	headers, err := ctx.ResolveMap(c.Proto.Headers)
	if err != nil {
		return nil, err
	}
	query, err := ctx.ResolveMap(c.Proto.Parameters)
	if err != nil {
		return nil, err
	}
	var body []byte
	if c.Proto.Body != nil {
		body, err = c.Proto.Body.Resolve(ctx)
		if err != nil {
			return nil, err
		}
		return &http.Request{
			Method: c.Proto.Method,
			URL: &url.URL{
				Scheme:   c.Proto.Scheme,
				Host:     fmt.Sprintf("%s:%d", host, c.Proto.GetPort()),
				Path:     path,
				RawQuery: encodeParams(query),
			},
			Header: headers,
			Body:   io.NopCloser(bytes.NewReader(body)),
		}, nil
	}
	return &http.Request{
		Method: c.Proto.Method,
		URL: &url.URL{
			Scheme:   c.Proto.Scheme,
			Host:     fmt.Sprintf("%s:%d", host, c.Proto.GetPort()),
			Path:     path,
			RawQuery: encodeParams(query),
		},
		Header: headers,
	}, nil
}

func encodeParams(params map[string][]string) string {
	values := url.Values{}
	for k, vs := range params {
		for _, v := range vs {
			values.Add(k, v)
		}
	}
	return values.Encode()
}
