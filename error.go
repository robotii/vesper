package vesper

import (
	"fmt"
	"strings"
)

var (
	// ErrorKey - used for generic errors
	ErrorKey = defaultVM.Intern("error:")
	// ArgumentErrorKey used for argument errors
	ArgumentErrorKey = defaultVM.Intern("argument-error:")
	// SyntaxErrorKey used for syntax errors
	SyntaxErrorKey = defaultVM.Intern("syntax-error:")
	// MacroErrorKey used for macro errors
	MacroErrorKey = defaultVM.Intern("macro-error:")
	// IOErrorKey used for IO errors
	IOErrorKey = defaultVM.Intern("io-error:")
	// InterruptKey used for interrupts that were captured
	InterruptKey = defaultVM.Intern("interrupt:")
	// InternalErrorKey used for internal errors
	InternalErrorKey = defaultVM.Intern("internal-error:")
)

// Error creates a new Error from the arguments. The first is an actual Vesper keyword object,
// the rest are interpreted as/converted to strings
func Error(errkey *Object, args ...interface{}) error {
	var buf strings.Builder
	for _, o := range args {
		if l, ok := o.(*Object); ok {
			buf.WriteString(fmt.Sprintf("%v", Write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	if errkey.Type != KeywordType {
		errkey = ErrorKey
	}
	return MakeError(errkey, String(buf.String()))
}

// MakeError creates an error object with the given data
func MakeError(elements ...*Object) *Object {
	data := Array(elements...)
	return &Object{Type: ErrorType, car: data}
}

// IsError returns true if the object is an error
func IsError(o interface{}) bool {
	if o == nil {
		return false
	}
	if err, ok := o.(*Object); ok {
		if err.Type == ErrorType {
			return true
		}
	}
	return false
}

// ErrorData returns the data associated with the error
func ErrorData(err *Object) *Object {
	return err.car
}

// Error converts the error to a string
func (lob *Object) Error() string {
	if lob.Type == ErrorType {
		s := lob.car.String()
		if lob.text != "" {
			s += " [in " + lob.text + "]"
		}
		return s
	}
	return lob.String()
}
