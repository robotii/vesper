package kana

// PrimitiveFunction is the native go function signature for all Ell primitive functions
type PrimitiveFunction func(argv []*Object) (*Object, error)

// Primitive - a primitive function, written in Go, callable by VM
type primitive struct { // <function>
	name      string
	fun       PrimitiveFunction
	signature string
	argc      int       // -1 means the primitive itself checks the args (legacy mode)
	args      []*Object // if set, the length must be for total args (both required and optional). The type (or <any>) for each
	rest      *Object   // if set, then any number of this type can follow the normal args. Mutually incompatible with defaults/keys
	defaults  []*Object // if set, then that many optional args beyond argc have these default values
	keys      []*Object // if set, then it must match the size of defaults, and these are the keys
}

// Continuation -
type continuation struct {
	ops   []int
	stack []*Object
	pc    int
}

type frame struct {
	locals   *frame
	previous *frame
	code     *Code
	ops      []int
	elements []*Object
	pc       int
}

type vm struct {
	stackSize int
}

// exec
func (vm *vm) exec(code *Code, env *frame) (*Object, error) {
	return nil, nil
}
