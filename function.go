package kana

import "fmt"

// Apply is a primitive instruction to apply a function to a list of arguments
var Apply = &Object{Type: FunctionType}

// CallCC is a primitive instruction to executable (restore) a continuation
var CallCC = &Object{Type: FunctionType}

// Spawn is a primitive instruction to apply a function to a list of arguments
var Spawn = &Object{Type: FunctionType}

// IsFunction returns true if the object is a function
func IsFunction(obj *Object) bool {
	return obj.Type == FunctionType
}

// Closure creates a new closure in the given frame
func Closure(code *Code, frame *frame) *Object {
	return &Object{
		Type:  FunctionType,
		code:  code,
		frame: frame,
	}
}

func functionToString(f *Object) string {
	if f.primitive != nil {
		return "#[function " + f.primitive.name + "]"
	}
	if f.code != nil {
		n := f.code.name
		if n == "" {
			return fmt.Sprintf("#[function]")
		}
		return fmt.Sprintf("#[function %s]", n)
	}
	if f.continuation != nil {
		return "#[continuation]"
	}
	if f == Apply {
		return "#[function apply]"
	}
	if f == CallCC {
		return "#[function callcc]"
	}
	if f == Spawn {
		return "#[function spawn]"
	}
	panic("Bad function")
}
