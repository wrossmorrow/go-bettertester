package bettertester

import (
	"net/http"
	"strings"
)

type CaseInsensitiveMap[T any] struct {
	Data map[string]T
}

func NewCaseInsensitiveMap[T any]() *CaseInsensitiveMap[T] {
	return &CaseInsensitiveMap[T]{
		Data: make(map[string]T),
	}
}

func (m *CaseInsensitiveMap[T]) Set(key string, value T) {
	m.Data[strings.ToLower(key)] = value
}

func (m *CaseInsensitiveMap[T]) Get(key string) (T, bool) {
	v, ok := m.Data[strings.ToLower(key)]
	return v, ok
}

type Headers struct {
	Headers CaseInsensitiveMap[[]string]
}

func NewHeaders() Headers {
	return Headers{
		Headers: *NewCaseInsensitiveMap[[]string](),
	}
}

func NewHeadersFromHttpHeaders(h http.Header) Headers {
	headers := NewHeaders()
	for k, v := range h {
		headers.SetAll(k, v)
	}
	return headers
}

func (h *Headers) Set(key string, value string) {
	h.Headers.Set(key, []string{value})
}

func (h *Headers) SetAll(key string, value []string) {
	h.Headers.Set(key, value)
}

func (h *Headers) Add(key string, value string) {
	if values, ok := h.Headers.Get(key); ok {
		values = append(values, value)
		h.Headers.Set(key, values)
	} else {
		h.Headers.Set(key, []string{value})
	}
}

func (h *Headers) Get(key string) ([]string, bool) {
	return h.Headers.Get(key)
}

type HeaderContainer interface {
	GetHeaders() Headers
	GetHeaderValues(string) ([]string, bool)
	GetHeaderValue(string, int) (string, error)
}
