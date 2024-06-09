package bettertester

import (
	"fmt"

	pb "github.com/wrossmorrow/bettertester/gen/proto/go/config/v1"
)

type Assertion interface {
	GetReferences() []string
	Assert(*Response) error
}

func NewAssertionFromProtobuf(a *pb.Assertion) Assertion {
	switch a.GetAssertion().(type) {
	case *pb.Assertion_Status:
		return &StatusAssertion{Expected: int(a.GetStatus().GetCode())}
	case *pb.Assertion_Exists:
		return &ExistsAssertion{Path: a.GetExists().GetPath()}
	case *pb.Assertion_Contains:
		return &ContainsAssertion{Path: a.GetContains().GetPath(), Value: a.GetContains().GetValue()}
	case *pb.Assertion_Equals:
		return &EqualsAssertion{Path: a.GetEquals().GetPath(), Value: a.GetEquals().GetValue()}
	case *pb.Assertion_Matches:
		return &MatchesAssertion{Path: a.GetMatches().GetPath(), Pattern: a.GetMatches().GetPattern()}
	default:
		return &NoOpAssertion{}
	}
}

type NoOpAssertion struct {
}

func (a *NoOpAssertion) GetReferences() []string {
	return nil
}

func (a *NoOpAssertion) Assert(r *Response) error {
	return nil
}

type StatusAssertion struct {
	Expected int
}

func (a *StatusAssertion) GetReferences() []string {
	return nil
}

func (a *StatusAssertion) Assert(r *Response) error {
	if r.Status != a.Expected {
		return fmt.Errorf("expected status code %d, got %d", a.Expected, r.Status)
	}
	return nil
}

type ExistsAssertion struct {
	Path string
}

func (a *ExistsAssertion) GetReferences() []string {
	return FindAllRefString(a.Path)
}

func (a *ExistsAssertion) Assert(r *Response) error {
	// lookup path like response.(headers.name(.idx)?|body.path)
	return nil
}

type ContainsAssertion struct {
	Path  string
	Value string
}

func (a *ContainsAssertion) GetReferences() []string {
	return FindAllRefString(a.Value)
}

func (a *ContainsAssertion) Assert(r *Response) error {
	return nil
}

type EqualsAssertion struct {
	Path  string
	Value string // interface?
}

func (a *EqualsAssertion) GetReferences() []string {
	return FindAllRefString(a.Value)
}

func (a *EqualsAssertion) Assert(r *Response) error {
	return nil
}

type MatchesAssertion struct {
	Path    string
	Pattern string
}

func (a *MatchesAssertion) GetReferences() []string {
	return nil
}

func (a *MatchesAssertion) Assert(r *Response) error {
	return nil
}
