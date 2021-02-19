package vesper

// IsArray returns true if the object is an array
func IsArray(obj *Object) bool {
	return obj.Type == ArrayType
}

// MakeArray - create a new <array> object of the specified size, with all elements initialized to
// the specified value
func MakeArray(size int, init *Object) *Object {
	elements := make([]*Object, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return ArrayFromElementsNoCopy(elements)
}

// Array - create a new <array> object from the given element objects.
func Array(elements ...*Object) *Object {
	return ArrayFromElements(elements, len(elements))
}

// ArrayFromElements - return a new <array> object from the given slice of elements. The slice is copied.
func ArrayFromElements(elements []*Object, count int) *Object {
	el := make([]*Object, count)
	copy(el, elements[0:count])
	return ArrayFromElementsNoCopy(el)
}

// ArrayFromElementsNoCopy - create a new <array> object from the given slice of elements. The slice is NOT copied.
func ArrayFromElementsNoCopy(elements []*Object) *Object {
	return &Object{
		Type:     ArrayType,
		elements: elements,
	}
}

// CopyArray - return a copy of the <array> object
func CopyArray(a *Object) *Object {
	return ArrayFromElements(a.elements, len(a.elements))
}
