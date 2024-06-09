package bettertester

type SimpleSet struct {
	Values map[string]bool
}

func NewSimpleSet() SimpleSet {
	return SimpleSet{
		Values: make(map[string]bool),
	}
}

func NewSimpleSetFrom(v []string) SimpleSet {
	s := NewSimpleSet()
	s.AddAll(v)
	return s
}

func (s *SimpleSet) IsEmpty() bool {
	return len(s.Values) == 0
}

func (s *SimpleSet) Add(v string) {
	if _, ok := s.Values[v]; !ok {
		s.Values[v] = true
	}
}

func (s *SimpleSet) AddAll(v []string) {
	for _, i := range v {
		s.Add(i)
	}
}

func (s *SimpleSet) Contains(v string) bool {
	_, ok := s.Values[v]
	return ok
}

func (s *SimpleSet) AsSlice() []string {
	set := make([]string, 0, len(s.Values))
	for k := range s.Values {
		set = append(set, k)
	}
	return set
}
