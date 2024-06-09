package bettertester

import (
	"fmt"
	"net/http"
	"strconv"
)

type Request struct {
	Method     string
	Scheme     string
	Host       string
	Port       int
	Path       string
	Headers    Headers
	Parameters map[string][]string
	Body       []byte
}

func (r *Request) GetHeaders() Headers {
	return r.Headers
}

func (r *Request) GetHeaderValues(key string) ([]string, bool) {
	return r.Headers.Get(key)
}

func (r *Request) GetHeaderValue(key string, index int) (string, error) {
	if values, ok := r.Headers.Get(key); ok {
		if index < 0 || index >= len(values) {
			return "", fmt.Errorf("index out of bounds")
		}
		return values[index], nil
	}
	return "", fmt.Errorf("header not found")
}

func (r *Request) GetHeaderPath(path []string) (string, error) {
	v, err := findInGenericMap(r.Headers, path)
	if err != nil {
		return "", err
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", nil
}

func (r *Request) GetBody() []byte {
	return r.Body
}

func (r *Request) GetBodyPath(path []string) (interface{}, error) {
	return "", nil
}

func requestFromHttpRequest(req *http.Request) Request {
	port, err := strconv.Atoi(req.URL.Port())
	if err != nil {
		if req.URL.Scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}
	return Request{
		Method:     req.Method,
		Scheme:     req.URL.Scheme,
		Host:       req.URL.Host,
		Port:       port,
		Path:       req.URL.Path,
		Headers:    NewHeadersFromHttpHeaders(req.Header),
		Parameters: req.URL.Query(),
		Body:       nil, // TODO: Read body
	}
}
