package vesper

// Object represents all objects in Vesper
type Object struct {
	Type         *Object               // i.e. <string>
	code         *Code                 // non-nil for closure, code
	frame        *frame                // non-nil for closure, continuation
	primitive    *primitive            // non-nil for primitives
	continuation *continuation         // non-nil for continuation
	car          *Object               // non-nil for instances and lists
	cdr          *Object               // non-nil for lists, nil for everything else
	bindings     map[structKey]*Object // non-nil for struct
	elements     []*Object             // non-nil for array
	fval         float64               // number
	text         string                // string, symbol, keyword, type
	Value        interface{}           // the rest of the data for more complex things
}

type stringable interface {
	String() string
}

// TypeType is the metatype, the type of all types
var TypeType *Object

// KeywordType is the type of all keywords
var KeywordType *Object

// SymbolType is the type of all symbols
var SymbolType *Object

// CharacterType is the type of all characters
var CharacterType = Intern("<character>")

// NumberType is the type of all numbers
var NumberType = Intern("<number>")

// StringType is the type of all strings
var StringType = Intern("<string>")

// BlobType is the type of all bytearrays
var BlobType = Intern("<blob>")

// ListType is the type of all lists
var ListType = Intern("<list>")

// ArrayType is the type of all arrays
var ArrayType = Intern("<array>")

// StructType is the type of all structs
var StructType = Intern("<struct>")

// FunctionType is the type of all functions
var FunctionType = Intern("<function>")

// CodeType is the type of compiled code
var CodeType = Intern("<code>")

// ErrorType is the type of all errors
var ErrorType = Intern("<error>")

// AnyType is a pseudo type specifier indicating any type
var AnyType = Intern("<any>")

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
	return ""
}
func IsType(obj *Object) bool {
	return obj.Type == TypeType
}

