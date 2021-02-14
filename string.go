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
