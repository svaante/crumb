package crumb

type SimpleStack struct {
    slice []string
    head int
}

func NewSimpleStack(slice []string) *SimpleStack {
    return &SimpleStack{
        slice: slice,
        head: 0,
    }
}

func (s *SimpleStack) Peek() string {
    return s.slice[s.head]
}

func (s *SimpleStack) Pop() string {
    head := s.head
    s.head += 1
    return s.slice[head]
}

func (s *SimpleStack) Prepend(slice []string) {
    s.slice = append(slice, s.slice[s.head:]...)
    s.head = 0
}

func (s *SimpleStack) Empty() []string {
    head := s.head
    s.head = len(s.slice)
    return s.slice[head:]
}

func (s *SimpleStack) Size() int {
    return len(s.slice) - s.head
}

