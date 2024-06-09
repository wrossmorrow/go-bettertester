package bettertester

type Stack struct {
	Items []interface{}
}

func NewStack() Stack {
	return Stack{
		Items: make([]interface{}, 0),
	}
}

func (s *Stack) Len() int {
	return len(s.Items)
}

func (s *Stack) Push(item interface{}) {
	s.Items = append(s.Items, item)
}

func (s *Stack) Pop() interface{} {
	if len(s.Items) == 0 {
		return nil
	}
	item := s.Items[0]
	s.Items = s.Items[1:len(s.Items)]
	return item
}

func (s *Stack) Printable() []string {
	items := make([]string, 0)
	for _, i := range s.Items {
		items = append(items, i.(*Call).Name)
	}
	return items
}
