package vesper

// EmptyString an empty string
var EmptyString = String("")

// String - create a new string object
func String(s string) *Object {
	return &Object{
		Type: StringType,
		text: s,
	}
}

// IsString returns true if the object is a string
func IsString(obj *Object) bool {
	return obj.Type == StringType
}

