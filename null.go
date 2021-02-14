package vesper

// NullType the type of the null object
var NullType = Intern("<null>")

// Null is a singleton representing nothing. It is distinct from an empty list.
var Null = &Object{Type: NullType}

// IsNull returns true if the object is the null object
func IsNull(obj *Object) bool {
	return obj == Null
}
