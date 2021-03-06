package vesper

// BooleanType is the type of true and false
var BooleanType = defaultVM.Intern("<boolean>")

// True is the singleton boolean true value
var True = &Object{Type: BooleanType, fval: 1}

// False is the singleton boolean false value
var False = &Object{Type: BooleanType, fval: 0}

// IsBoolean returns true if the object type is boolean
func IsBoolean(obj *Object) bool {
	return obj.Type == BooleanType
}

// BoolValue - return native bool value of the object
func BoolValue(obj *Object) bool {
	return obj == True
}
