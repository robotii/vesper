package vesper

import "fmt"

// Apply is a primitive instruction to apply a function to a list of arguments
var Apply = &Object{Type: FunctionType}

// CallCC is a primitive instruction to executable (restore) a continuation
var CallCC = &Object{Type: FunctionType}

// GoFunc is a primitive instruction to create a goroutine
var GoFunc = &Object{Type: FunctionType}

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
			return "#[function]"
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
	if f == GoFunc {
		return "#[function go]"
	}
	panic("Bad function")
}

func functionSignature(f *Object) string {
	if f.primitive != nil {
		return f.primitive.signature
	}
	if f.code != nil {
		return f.code.signature()
	}
	if f.continuation != nil {
		return "(<function>) <any>"
	}
	if f == Apply {
		return "(<any>*) <list>"
	}
	if f == CallCC {
		return "(<function>) <any>"
	}
	if f == GoFunc {
		return "(<function> <any>*) <null>"
	}
	panic("Bad function")
}

func functionSignatureFromTypes(result *Object, args []*Object, rest *Object) string {
	sig := "("
	for i, t := range args {
		if !IsType(t) {
			panic("not a type: " + t.String())
		}
		if i > 0 {
			sig += " "
		}
		sig += t.text
	}
	if rest != nil {
		if !IsType(rest) {
			panic("not a type: " + rest.String())
		}
		if sig != "(" {
			sig += " "
		}
		sig += rest.text + "*"
	}
	sig += ") "
	if !IsType(result) {
		panic("not a type: " + result.String())
	}
	sig += result.text
	return sig
}
