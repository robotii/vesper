package vesper

import "fmt"

// Code - compiled Vesper bytecode
type Code struct {
	name     string
	ops      []int
	argc     int
	defaults []*Object
	keys     []*Object
}

// IsCode returns true if the object is a code object
func IsCode(obj *Object) bool {
	return obj.Type == CodeType
}

// MakeCode creates a new code object
func MakeCode(argc int, defaults []*Object, keys []*Object, name string) *Object {
	return &Object{
		Type: CodeType,
		code: &Code{
			name:     name,
			ops:      nil,
			argc:     argc,
			defaults: defaults, //nil for normal procs, empty for rest, and non-empty for optional/keyword
			keys:     keys,
		},
	}
}
func (code *Code) String() string {
	// TODO: Better string representation
	return fmt.Sprintf("(function (%d %v %s) %v)", code.argc, code.defaults, code.keys, code.ops)
}
