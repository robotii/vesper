package vesper

import (
	"strings"
)

// IsArray returns true if the object is an array
func IsArray(obj *Object) bool {
	return obj.Type == ArrayType
}

// ArrayEqual - return true of the two arrays are equal,
// i.e. the same length and all the elements are also equal
func ArrayEqual(v1 *Object, v2 *Object) bool {
	el1 := v1.elements
	el2 := v2.elements
	count := len(el1)
	if count != len(el2) {
		return false
	}
	for i := 0; i < count; i++ {
		if !Equal(el1[i], el2[i]) {
			return false
		}
	}
	return true
}

func arrayToString(a *Object) string {
	el := a.elements
	var buf strings.Builder
	buf.WriteString("[")
	count := len(el)
	if count > 0 {
		buf.WriteString(el[0].String())
		for i := 1; i < count; i++ {
			buf.WriteString(" ")
			buf.WriteString(el[i].String())
		}
	}
	buf.WriteString("]")
	return buf.String()
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

// ToArray - convert the object to an <array>, if possible
func ToArray(obj *Object) (*Object, error) {
	switch obj.Type {
	case ArrayType:
		return obj, nil
	case ListType:
		return listToArray(obj), nil
	case StructType:
		return structToArray(obj), nil
	case StringType:
		return stringToArray(obj), nil
	}
	return nil, Error(ArgumentErrorKey, "to-array expected <array>, <list>, <struct>, or <string>, got a ", obj.Type)
}
