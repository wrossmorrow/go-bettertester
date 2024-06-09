package bettertester

import (
	"fmt"
	"regexp"
)

var (
	ref = regexp.MustCompile(`\$\{((env|const|stored|call):([a-zA-Z0-9-_.]+)*([a-zA-Z0-9-_]+))\}`)
)

type FlattenableReferences interface {
	Flatten() []string
}

type StoredReferences struct {
	Values map[string]string
}

func NewStoredReferences() *StoredReferences {
	return &StoredReferences{
		Values: make(map[string]string),
	}
}

func (s *StoredReferences) AddReference(ref, name string) {
	s.Values[ref] = name
}

func (s *StoredReferences) AddNewReference(ref, name string) error {
	_, ok := s.Values[ref]
	if ok {
		return fmt.Errorf("reference \"%s\" already exists", ref)
	}
	s.AddReference(ref, name)
	return nil
}

func (s *StoredReferences) GetReference(ref string) (string, bool) {
	v, ok := s.Values[ref]
	return v, ok
}

func (s *StoredReferences) GetInverted() map[string]string {
	inv := make(map[string]string)
	for k, v := range s.Values {
		inv[v] = k
	}
	return inv
}

func (s *StoredReferences) GetReferences() map[string]string {
	return s.Values
}

type RequestReferences struct {
	Host       []string
	Path       []string
	Headers    map[string][]string
	Parameters map[string][]string
	Body       []string
}

type CallReferences struct {
	Request    RequestReferences
	Assertions map[int][]string
	After      []string
}

func (r *RequestReferences) Flatten() []string {
	refs := NewSimpleSet()
	if r.Host != nil {
		refs.AddAll(r.Host)
	}
	if r.Path != nil {
		refs.AddAll(r.Path)
	}
	if r.Headers != nil {
		for _, v := range r.Headers {
			refs.AddAll(v)
		}
	}
	if r.Parameters != nil {
		for _, v := range r.Parameters {
			refs.AddAll(v)
		}
	}
	if r.Body != nil {
		refs.AddAll(r.Body)
	}
	return refs.AsSlice()
}

func (r *CallReferences) Flatten() []string {
	refs := NewSimpleSet()
	refs.AddAll(r.Request.Flatten())
	for _, v := range r.Assertions {
		refs.AddAll(v)
	}
	if r.After != nil {
		for _, v := range r.After {
			refs.Add(fmt.Sprintf("call:%s", v))
		}
	}
	return refs.AsSlice()
}

func FindAllRefString(s string) []string {
	refs := NewSimpleSet()
	matches := ref.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		refs.Add(m[1])
	}
	return refs.AsSlice()
}

func FindAllRefsStringSlice(s []string) []string {
	if s == nil {
		return nil
	}
	matches := NewSimpleSet()
	for _, v := range s {
		m := FindAllRefString(v)
		if m != nil {
			matches.AddAll(m)
		}
	}
	return matches.AsSlice()
}

func FindAllRefsStringMap(d map[string]string) map[string][]string {
	if d == nil {
		return nil
	}
	matches := make(map[string][]string)
	for k, v := range d {
		m := FindAllRefString(v)
		if m != nil {
			matches[k] = m
		}
	}
	return matches
}

func FindAllRefsStringMapSlice(d map[string][]string) map[string][]string {
	if d == nil {
		return nil
	}
	matches := make(map[string][]string)
	for k, v := range d {
		m := FindAllRefsStringSlice(v)
		if len(m) > 0 {
			matches[k] = m
		}
	}
	return matches
}
