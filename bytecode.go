package vesper

import "fmt"

// Code represents compiled bytecode that can be called
type Code struct {
	name     string
	ops      []int
	argc     int
	defaults []*Object
	keys     []*Object
}

func (code *Code) String() string {
	// TODO: Better string representation
	return fmt.Sprintf("(function (%d %v %s) %v)", code.argc, code.defaults, code.keys, code.ops)
}
