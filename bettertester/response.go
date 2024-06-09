package bettertester

import (
	"fmt"
	"io"
	"net/http"
)

type Response struct {
	Status  int
	Headers Headers
	Body    []byte
}

func (r *Response) GetHeaders() Headers {
	return r.Headers
}

func (r *Response) GetHeaderValues(key string) ([]string, bool) {
	return r.Headers.Get(key)
}

func (r *Response) GetHeaderValue(key string, index int) (string, error) {
	if values, ok := r.Headers.Get(key); ok {
		if index < 0 || index >= len(values) {
			return "", fmt.Errorf("index out of bounds")
		}
		return values[index], nil
	}
	return "", fmt.Errorf("header not found")
}

func (r *Response) GetHeaderPath(path []string) (string, error) {
	v, err := findInGenericMap(r.Headers, path)
	if err != nil {
		return "", err
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("path not found (string expected, got %s)", printableType(v))
}

func (r *Response) GetBody() []byte {
	return r.Body
}

func (r *Response) GetBodyPath(path []string) (interface{}, error) {
	return "", nil
}

func responseFromHttpResponse(resp *http.Response) (Response, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{
			Status:  resp.StatusCode,
			Headers: NewHeadersFromHttpHeaders(resp.Header),
			Body:    nil,
		}, err
	}
	return Response{
		Status:  resp.StatusCode,
		Headers: NewHeadersFromHttpHeaders(resp.Header),
		Body:    body,
	}, err
}
