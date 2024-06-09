package bettertester

import (
	"encoding/json"

	pb "github.com/wrossmorrow/bettertester/gen/proto/go/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type BodyContainer interface {
	GetBody() []byte
	GetBodyPath(path []string) (interface{}, error)
}

type ResolvableBody interface {
	GetReferences() []string
	Resolve(ctx *ExecutionContext) ([]byte, error)
}

type BodyBinary struct {
	Bytes []byte
}

func NewBodyBinary(b []byte) *BodyBinary {
	return &BodyBinary{Bytes: b}
}

func (b *BodyBinary) GetReferences() []string {
	return nil
}

func (b *BodyBinary) Resolve(ctx *ExecutionContext) ([]byte, error) {
	return b.Bytes, nil
}

type BodyText struct {
	Text       string
	References SimpleSet
}

func NewBodyText(s string) *BodyText {
	return &BodyText{Text: s, References: NewSimpleSetFrom(FindAllRefString(s))}
}

func (b *BodyText) GetReferences() []string {
	return b.References.AsSlice()
}

func (b *BodyText) Resolve(ctx *ExecutionContext) ([]byte, error) {
	// Iterate over stored references
	// Get values from the context
	// Replace the references with the values
	return []byte(b.Text), nil
}

// we expect this to be flat, map[string](bool|int|float|string)
type BodyForm struct {
	Form       map[string]interface{}
	References SimpleSet
}

func NewBodyForm(d map[string]interface{}) *BodyForm {
	return &BodyForm{Form: d, References: NewSimpleSetFrom(getRefs(d))}
}

func (b *BodyForm) GetReferences() []string {
	return b.References.AsSlice()
}

func (b *BodyForm) Resolve(ctx *ExecutionContext) ([]byte, error) {
	// Iterate over stored references
	// Get values from the context
	// Replace the references with the values
	// How to "map" references to "locations"? Flattening? Serialize -> Replace?
	bytes, err := json.Marshal(b.Form)
	// replace references with values...
	return bytes, err
}

type BodyJson struct {
	Json       map[string]interface{}
	References SimpleSet
}

func NewBodyJson(d map[string]interface{}) *BodyJson {
	return &BodyJson{Json: d, References: NewSimpleSetFrom(getRefs(d))}
}

func (b *BodyJson) GetReferences() []string {
	return b.References.AsSlice()
}

func (b *BodyJson) Resolve(ctx *ExecutionContext) ([]byte, error) {
	// Iterate over stored references
	// Get values from the context
	// Replace the references with the values
	// How to "map" references to "locations"? Flattening? Serialize -> Replace?
	bytes, err := json.Marshal(b.Json)
	// replace references with values...
	return bytes, err
}

func NewBodyFromProtobuf(p *pb.CallBody) ResolvableBody {
	if p == nil {
		return nil
	}
	switch p.GetBody().(type) {
	// case *pb.CallBody_Binary:
	// 	return &BodyBinary{}
	case *pb.CallBody_Text:
		return NewBodyText(p.GetText())
	case *pb.CallBody_Form:
		j, err := StructToInterface(p.GetForm())
		if err != nil {
			return nil
		}
		return NewBodyForm(j.(map[string]interface{}))
	case *pb.CallBody_Json:
		j, err := StructToInterface(p.GetJson())
		if err != nil {
			return nil
		}
		return NewBodyJson(j.(map[string]interface{}))
	default:
		return nil
	}
}

func StructToInterface(s proto.Message) (interface{}, error) {
	b, err := protojson.Marshal(s)
	if err != nil {
		return nil, err
	}
	var m interface{}
	err = json.Unmarshal(b, &m)
	return m, err
}

func getRefs(data interface{}) []string {
	switch d := data.(type) {
	case map[string]interface{}:
		refs := make([]string, 0)
		for _, v := range d {
			refs = append(refs, getRefs(v)...)
		}
		return refs
	case []interface{}:
		refs := make([]string, 0)
		for _, v := range d {
			refs = append(refs, getRefs(v)...)
		}
		return refs
	case string:
		return FindAllRefString(d)
	case int:
		return nil
	case float64:
		return nil
	case bool:
		return nil
	default:
		return nil
	}
}
