package vesper

type continuation struct {
	ops   []int
	stack []*Object
	pc    int
}

// Continuation creates a continuation object
func Continuation(frame *frame, ops []int, pc int, stack []*Object) *Object {
	newStack := make([]*Object, len(stack))
	copy(newStack, stack)
	return &Object{
		Type:  FunctionType,
		frame: frame,
		continuation: &continuation{
			ops:   ops,
			stack: newStack,
			pc:    pc,
		},
	}
}
