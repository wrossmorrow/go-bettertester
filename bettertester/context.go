package bettertester

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ExecutionContext struct {
	// Options?
	Client    *http.Client
	Constants map[string]string
	Store     map[string]*StoredValue
	Results   map[string]*CallResult
}

func NewExecutionContext() *ExecutionContext {
	return NewExecutionContextWithClient(&http.Client{})
}

func NewExecutionContextWithCallTimeout(timeout time.Duration) *ExecutionContext {
	return NewExecutionContextWithClient(&http.Client{
		Timeout: timeout,
	})
}

func NewExecutionContextWithClient(client *http.Client) *ExecutionContext {
	return &ExecutionContext{
		Client:    client,
		Constants: make(map[string]string),
		Store:     make(map[string]*StoredValue),
		Results:   make(map[string]*CallResult),
	}
}

func (ctx *ExecutionContext) Start() {
	ctx.Store = make(map[string]*StoredValue)
	ctx.Results = make(map[string]*CallResult)
}

func (ctx *ExecutionContext) AddResult(name string, result *CallResult) {
	ctx.Results[name] = result
}

func (ctx *ExecutionContext) Finish() {

}

// execute calls, with store and result handling

func (ctx *ExecutionContext) Execute(call *Call) *CallResult {
	result := call.Execute(ctx)
	defer ctx.AddResult(call.Name, result)
	if result.IsError() {
		return result
	}
	for _, a := range call.Assertions {
		err := a.Assert(&result.Response)
		if err != nil {
			result.SetError(err)
			return result
		}
	}
	for k, v := range call.Store {
		parts := strings.Split(k, ".")
		if len(parts) < 2 {
			fmt.Printf("invalid store key: %s\n", k)
			continue
		}
		switch parts[0] {
		case "response":
			switch parts[1] {
			case "status":
				ctx.Store[k] = &StoredValue{
					Name:  k,
					Value: result.Response.Status,
				}
			case "headers":
				name := parts[2]
				hc := result.GetHeaderContainer("response")
				if len(parts) == 3 {
					w, ok := hc.GetHeaderValues(name)
					if ok {
						ctx.Store[k] = &StoredValue{Name: k, Value: w}
					} else {
						ctx.Store[k] = &StoredValue{Name: k, Value: nil}
					}
				} else if len(parts) == 4 {
					i, err := strconv.Atoi(parts[3])
					if err != nil {
						ctx.Store[k] = &StoredValue{Name: k, Value: nil}
					} else {
						w, err := hc.GetHeaderValue(name, i)
						if err != nil {
							ctx.Store[k] = &StoredValue{Name: k, Value: nil}
						} else {
							ctx.Store[k] = &StoredValue{Name: k, Value: w}
						}
					}
				} else {
					ctx.Store[k] = &StoredValue{Name: k, Value: nil}
				}
			case "body":
				if len(parts) == 2 {
					ctx.Store[k] = &StoredValue{Name: k, Value: result.Response.Body}
				} else {
					path := parts[2:]
					r, err := result.GetResponseBodyJsonPath(path)
					if err != nil {
						result.SetError(err)
					} else {
						ctx.Store[v] = &StoredValue{Name: k, Value: r}
						fmt.Printf("storing value: %s == %v\n", k, r)
					}
				}
			default:
			}
		default:
			fmt.Printf("unrecognized store key: %s\n", k)
		}
	}
	return result
}

// resolve placeholders

func (ctx *ExecutionContext) ResolveBytes(b []byte) ([]byte, error) {
	return b, nil
}

func (ctx *ExecutionContext) ResolveMap(m map[string][]string) (map[string][]string, error) {
	result := make(map[string][]string)
	for k, v := range m {
		r, err := ctx.ResolveList(v)
		if err != nil {
			return nil, err
		}
		result[k] = r
	}
	return result, nil
}

func (ctx *ExecutionContext) ResolveList(l []string) ([]string, error) {
	result := make([]string, len(l))
	for i, s := range l {
		r, err := ctx.ResolveString(s)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}

func (ctx *ExecutionContext) ResolveString(s string) (string, error) {
	matches := FindAllRefString(s)
	if len(matches) == 0 {
		return s, nil
	}
	fmt.Printf("resolving \"%s\" -> ", s)
	for _, m := range matches {
		// get value from context, depending on type (env:, const:, store:, call:)
		// and replace all occurrences of the reference with the value
		var v string
		if strings.HasPrefix(m, "env:") {
			v = os.Getenv(strings.TrimPrefix(m, "env:"))
		} else if strings.HasPrefix(m, "const:") {
			v = ctx.Constants[strings.TrimPrefix(m, "const:")]
		} else if strings.HasPrefix(m, "stored:") {
			v = ctx.Store[strings.TrimPrefix(m, "stored:")].Value.(string)
		} else if strings.HasPrefix(m, "call:") {
			// expect <name>.(request.(headers|body)|response.(headers|body)).<path>
			c := strings.TrimPrefix(m, "call:")
			parts := strings.Split(c, ".")
			if len(parts) < 3 {
				return "", fmt.Errorf("invalid call reference: %s", m)
			}
			name := parts[0]
			r, ok := ctx.Results[name]
			if !ok {
				return "", fmt.Errorf("call result not found: %s", name)
			}
			phase := parts[1]
			part := parts[2]
			path := parts[3:]
			switch part {
			case "headers":
				hc := r.GetHeaderContainer(phase)
				if len(path) == 1 {
					w, ok := hc.GetHeaderValues(path[0])
					if !ok {
						return "", fmt.Errorf("header not found: %s", path[0])
					}
					v = strings.Join(w, ",") // TODO: is this always a string?
				} else if len(path) == 2 {
					i, err := strconv.Atoi(path[1])
					if err != nil {
						return "", fmt.Errorf("invalid header path: %s.%s", path[0], path[1])
					}
					w, err := hc.GetHeaderValue(path[0], i)
					if err != nil {
						return "", err
					}
					v = w
				} else {
					return "", fmt.Errorf("invalid header path: %v", path)
				}
			case "body":
				bc := r.GetBodyContainer(phase)
				w, err := bc.GetBodyPath(path)
				if err != nil {
					return "", err
				}
				v = w.(string) // TODO: is this always a string?
			default:
			}
		} else {
			return "", fmt.Errorf("unknown reference type \"%s\"", m)
		}
		s = strings.ReplaceAll(s, fmt.Sprintf("${%s}", m), v)
	}
	fmt.Printf("\"%s\"\n", s)
	return s, nil
}
