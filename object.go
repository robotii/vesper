package vesper

import (
	"fmt"
	"strconv"
)

// Object represents all objects in Vesper
type Object struct {
	Type         *Object             // i.e. <string>
	code         *Code               // non-nil for closure, code
	frame        *frame              // non-nil for closure, continuation
	primitive    *primitive          // non-nil for primitives
	continuation *continuation       // non-nil for continuation
	car          *Object             // non-nil for instances and lists
	cdr          *Object             // non-nil for lists, nil for everything else
	bindings     map[*Object]*Object // non-nil for struct
	elements     []*Object           // non-nil for array
	fval         float64             // number
	text         string              // string, symbol, keyword, type
	Value        interface{}         // the rest of the data for more complex things
}

type stringable interface {
	String() string
}

// TypeType is the metatype, the type of all types
var TypeType *Object // bootstrapped in initSymbolTable => Intern("<type>")

// KeywordType is the type of all keywords
var KeywordType *Object // bootstrapped in initSymbolTable => Intern("<keyword>")

// SymbolType is the type of all symbols
var SymbolType *Object // bootstrapped in initSymbolTable = Intern("<symbol>")

// CharacterType is the type of all characters
var CharacterType = defaultVM.Intern("<character>")

// NumberType is the type of all numbers
var NumberType = defaultVM.Intern("<number>")

// StringType is the type of all strings
var StringType = defaultVM.Intern("<string>")

// BlobType is the type of all bytearrays
var BlobType = defaultVM.Intern("<blob>")

// ListType is the type of all lists
var ListType = defaultVM.Intern("<list>")

// ArrayType is the type of all arrays
var ArrayType = defaultVM.Intern("<array>")

// StructType is the type of all structs
var StructType = defaultVM.Intern("<struct>")

// FunctionType is the type of all functions
var FunctionType = defaultVM.Intern("<function>")

// CodeType is the type of compiled code
var CodeType = defaultVM.Intern("<code>")

// ErrorType is the type of all errors
var ErrorType = defaultVM.Intern("<error>")

// AnyType is a pseudo type specifier indicating any type
var AnyType = defaultVM.Intern("<any>")

// RuneValue - return native rune value of the object
func RuneValue(obj *Object) rune {
	return rune(obj.fval)
}

// IntValue - return native int value of the object
func IntValue(obj *Object) int {
	return int(obj.fval)
}

// Int64Value - return native int64 value of the object
func Int64Value(obj *Object) int64 {
	return int64(obj.fval)
}

// Float64Value - return native float64 value of the object
func Float64Value(obj *Object) float64 {
	return obj.fval
}

// StringValue - return native string value of the object
func StringValue(obj *Object) string {
	return obj.text
}

// BlobValue - return native []byte value of the object
func BlobValue(obj *Object) []byte {
	b, _ := obj.Value.([]byte)
	return b
}

// NewObject is the constructor for externally defined objects, where the
// value is an interface{}.
func NewObject(variant *Object, value interface{}) *Object {
	return &Object{Type: variant, Value: value}
}

// Identical - return if two objects are identical
func Identical(o1 *Object, o2 *Object) bool {
	return o1 == o2
}

// String returns the string representation of the object
func (lob *Object) String() string {
	switch lob.Type {
	case NullType:
		return "null"
	case BooleanType:
		if lob == True {
			return "true"
		}
		return "false"
	case CharacterType:
		return string([]rune{rune(lob.fval)})
	case NumberType:
		return strconv.FormatFloat(lob.fval, 'f', -1, 64)
	case BlobType:
		return fmt.Sprintf("#[blob %d bytes]", len(BlobValue(lob)))
	case StringType, SymbolType, KeywordType, TypeType:
		return lob.text
	case ListType:
		return listToString(lob)
	case ArrayType:
		return arrayToString(lob)
	case StructType:
		return structToString(lob)
	case FunctionType:
		return functionToString(lob)
	case CodeType:
		return lob.code.String(lob.code.vm)
	case ErrorType:
		return "#<error>" + Write(lob.car)
	case ChannelType:
		return lob.Value.(*channel).String()
	default:
		if lob.Value != nil {
			if s, ok := lob.Value.(stringable); ok {
				return s.String()
			}
			return "#[" + typeNameString(lob.Type.text) + "]"
		}
		return "#" + lob.Type.text + Write(lob.car)
	}
}

// IsCharacter returns true if the object is a character
func IsCharacter(obj *Object) bool {
	return obj.Type == CharacterType
}

// IsNumber returns true if the object is a number
func IsNumber(obj *Object) bool {
	return obj.Type == NumberType
}

// IsList returns true if the object is a list
func IsList(obj *Object) bool {
	return obj.Type == ListType
}

// IsStruct returns true if the object is a struct
func IsStruct(obj *Object) bool {
	return obj.Type == StructType
}

// IsSymbol returns true if the object is a symbol
func IsSymbol(obj *Object) bool {
	return obj.Type == SymbolType
}

// IsKeyword returns true if the object is a keyword
func IsKeyword(obj *Object) bool {
	return obj.Type == KeywordType
}

// IsType returns true if the object is a type
func IsType(obj *Object) bool {
	return obj.Type == TypeType
}

// IsInstance returns true if the object is an instance.
// Since instances have arbitrary Type symbols, all we can check is that the car value is set
func IsInstance(obj *Object) bool {
	return obj.car != nil && obj.cdr == nil
}

// Equal checks if two objects are equal
func Equal(o1 *Object, o2 *Object) bool {
	if o1 == o2 {
		return true
	}
	if o1.Type != o2.Type {
		return false
	}
	switch o1.Type {
	case BooleanType, CharacterType:
		return int(o1.fval) == int(o2.fval)
	case NumberType:
		return NumberEqual(o1.fval, o2.fval)
	case StringType:
		return o1.text == o2.text
	case ListType:
		return ListEqual(o1, o2)
	case ArrayType:
		return ArrayEqual(o1, o2)
	case StructType:
		return StructEqual(o1, o2)
	case SymbolType, KeywordType, TypeType:
		return o1 == o2
	case NullType:
		return true // singleton
	default:
		o1a := Value(o1)
		if o1a != o1 {
			o2a := Value(o2)
			return Equal(o1a, o2a)
		}
		return false
	}
}

// IsPrimitiveType returns true if the object is a primitive
func IsPrimitiveType(tag *Object) bool {
	switch tag {
	case NullType, BooleanType, CharacterType, NumberType,
		StringType, ListType, ArrayType, StructType,
		SymbolType, KeywordType, TypeType, FunctionType:
		return true
	default:
		return false
	}
}

// Instance returns a new instance of the type and value
func Instance(tag *Object, val *Object) (*Object, error) {
	if !IsType(tag) {
		return nil, Error(ArgumentErrorKey, TypeType.text, tag)
	}
	if IsPrimitiveType(tag) {
		return val, nil
	}
	return &Object{
		Type: tag,
		car:  val,
	}, nil
}

// Value returns the result of dereferencing the object
func Value(obj *Object) *Object {
	if obj.cdr == nil && obj.car != nil {
		return obj.car
	}
	return obj
}
