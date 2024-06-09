package bettertester

import (
	pb "github.com/wrossmorrow/bettertester/gen/proto/go/config/v1"
)

type RequestProto struct {
	Method     string              `json:"method"`
	Scheme     string              `json:"scheme"`
	Host       string              `json:"host"`
	Port       int                 `json:"port"`
	Path       string              `json:"path"`
	Headers    map[string][]string `json:"headers"`
	Parameters map[string][]string `json:"parameters"`
	Body       ResolvableBody      // BodyJson?
}

func DefaultRequestProto() *RequestProto {
	return &RequestProto{
		Method:     "GET",
		Scheme:     "https",
		Host:       "google.com",
		Port:       443,
		Path:       "/",
		Headers:    make(map[string][]string),
		Parameters: make(map[string][]string),
		Body:       &BodyText{},
	}
}

func NewRequestProtoFromProtobuf(p *pb.CallProto) *RequestProto {
	return &RequestProto{
		Method:     p.GetMethod(),
		Scheme:     p.GetScheme(),
		Host:       p.GetHost(),
		Port:       int(p.GetPort()),
		Path:       p.GetPath(),
		Headers:    renderHeaders(p.GetHeaders()),
		Parameters: renderParams(p.GetParams()),
		Body:       NewBodyFromProtobuf(p.GetBody()),
	}
}

func (p *RequestProto) GetPort() int {
	if p.Port == 0 {
		if p.Scheme == "http" {
			return 80
		}
		if p.Scheme == "https" {
			return 443
		}
	}
	return p.Port
}

func (p *RequestProto) GetReferences() RequestReferences {
	return RequestReferences{
		Host:       p.GetHostReferences(),
		Path:       p.GetPathReferences(),
		Headers:    p.GetHeadersReferences(),
		Parameters: p.GetParametersReferences(),
		Body:       p.GetBodyReferences(),
	}
}

func (p *RequestProto) GetHostReferences() []string {
	if p.Path == "" {
		return nil
	}
	return FindAllRefString(p.Host)
}

func (p *RequestProto) GetPathReferences() []string {
	if p.Path == "" {
		return nil
	}
	return FindAllRefString(p.Path)
}

func (p *RequestProto) GetHeadersReferences() map[string][]string {
	if p.Headers == nil {
		return nil
	}
	return FindAllRefsStringMapSlice(p.Headers)
}

func (p *RequestProto) GetParametersReferences() map[string][]string {
	if p.Parameters == nil {
		return nil
	}
	return FindAllRefsStringMapSlice(p.Parameters)
}

func (p *RequestProto) GetBodyReferences() []string {
	if p.Body == nil {
		return nil
	}
	return p.Body.GetReferences()
}

func renderHeaders(headers []*pb.CallHeader) map[string][]string {
	m := make(map[string][]string)
	for _, h := range headers {
		if len(h.GetValues()) == 0 {
			if h.GetValue() == "" {
				// log ignore
				continue
			}
			m[h.GetName()] = []string{h.GetValue()}
		} else {
			m[h.GetName()] = h.GetValues()
		}
	}
	return m
}

func renderParams(params []*pb.CallParam) map[string][]string {
	m := make(map[string][]string)
	for _, p := range params {
		if len(p.GetValues()) == 0 {
			if p.GetValue() == "" {
				// log ignore
				continue
			}
			m[p.GetName()] = []string{p.GetValue()}
		} else {
			m[p.GetName()] = p.GetValues()
		}
	}
	return m
}
