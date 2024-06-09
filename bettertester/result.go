package bettertester

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type CallResult struct {
	Request  Request
	Response Response
	Metadata Metadata
	Err      error
}

func (r *CallResult) GetHeaderContainer(phase string) HeaderContainer {
	switch phase {
	case "request":
		return &r.Request
	case "response":
		return &r.Response
	default:
		return nil
	}
}

func (r *CallResult) GetBodyContainer(phase string) BodyContainer {
	switch phase {
	case "request":
		return &r.Request
	case "response":
		return &r.Response
	default:
		return nil
	}
}

func NewCallResultFromError(err error) *CallResult {
	return &CallResult{
		Request:  Request{},
		Response: Response{},
		Err:      err,
	}
}

func NewCallResultFromRequestAndError(req *http.Request, err error) *CallResult {
	return &CallResult{
		Request:  requestFromHttpRequest(req),
		Response: Response{},
		Err:      err,
	}
}

func NewCallResultFromRequestAndResponse(req *http.Request, resp *http.Response, err error) *CallResult {
	if err != nil {
		return NewCallResultFromRequestAndError(req, err)
	}
	if resp == nil || resp.StatusCode == 0 {
		return &CallResult{
			Request:  requestFromHttpRequest(req),
			Response: Response{},
			Err:      fmt.Errorf("no response"),
		}
	}
	r, err := responseFromHttpResponse(resp)
	return &CallResult{
		Request:  requestFromHttpRequest(req),
		Response: r,
		Err:      err,
	}
}

func (r *CallResult) IsError() bool {
	return r.Err != nil
}

func (r *CallResult) SetError(err error) {
	r.Err = err
}

func (r *CallResult) GetError() error {
	return r.Err
}

func (r *CallResult) GetResponseBodyJsonPath(path []string) (interface{}, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(r.Response.Body, &data)
	if err != nil {
		return nil, err
	}
	res, err := findInGenericMap(data, path)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func findInGenericMap(data interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return data, nil
	}
	p := path[0]
	if len(path) == 1 {
		switch d := data.(type) {
		case []interface{}:
			i, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("path not found (missing list index)")
			}
			if i < 0 {
				return nil, fmt.Errorf("path not found (invalid list index)")
			}
			if i >= len(d) {
				return nil, fmt.Errorf("path not found (list index out of bounds)")
			}
			return d[i], nil
		case map[string]interface{}:
			if v, ok := d[p]; !ok {
				return nil, fmt.Errorf("path not found (missing key)")
			} else {
				return v, nil
			}
		default:
			return nil, fmt.Errorf("path not found")
		}
	}
	switch d := data.(type) {
	case []interface{}:
		i, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("path not found (missing list index)")
		}
		if i < 0 {
			return nil, fmt.Errorf("path not found (invalid list index)")
		}
		if i >= len(d) {
			return nil, fmt.Errorf("path not found (list index out of bounds)")
		}
		return findInGenericMap(d[i], path[1:])
	case map[string]interface{}:
		if v, ok := d[p]; !ok {
			return nil, fmt.Errorf("path not found (missing key)")
		} else {
			return findInGenericMap(v, path[1:])
		}
	default:
		return nil, fmt.Errorf("path not found (object expected, got %s)", printableType(d))
	}
}

func printableType(data interface{}) string {
	switch data.(type) {
	case string:
		return "string"
	case int:
		return "int"
	case float64:
		return "float64"
	case bool:
		return "bool"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}
