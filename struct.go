package vesper

import (
	"strings"
)

// IsValidStructKey - return true of the object is a valid <struct> key.
func IsValidStructKey(o *Object) bool {
	switch o.Type {
	case StringType, SymbolType, KeywordType, TypeType:
		return true
	}
	return false
}

// EmptyStruct - a <struct> with no bindings
var EmptyStruct = MakeStruct(0)

// MakeStruct - create an empty <struct> object with the specified capacity
func MakeStruct(capacity int) *Object {
	return &Object{
		Type:     StructType,
		bindings: make(map[*Object]*Object, capacity),
	}
}

// Struct - create a new <struct> object from the arguments, which can be other structs, or key/value pairs
func Struct(fieldvals []*Object) (*Object, error) {
	strct := &Object{
		Type:     StructType,
		bindings: make(map[*Object]*Object),
	}
	count := len(fieldvals)
	i := 0
	var bindings map[*Object]*Object
	for i < count {
		o := Value(fieldvals[i])
		i++
		switch o.Type {
		case StructType: // not a valid key, just copy bindings from it
			if bindings == nil {
				bindings = make(map[*Object]*Object, len(o.bindings))
			}
			for k, v := range o.bindings {
				bindings[k] = v
			}
		case StringType, SymbolType, KeywordType, TypeType:
			if i == count {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in arglist: ", o)
			}
			if bindings == nil {
				bindings = make(map[*Object]*Object)
			}
			bindings[o] = fieldvals[i]
			i++
		default:
			return nil, Error(ArgumentErrorKey, "Bad struct key: ", o)
		}
	}
	if bindings == nil {
		strct.bindings = make(map[*Object]*Object)
	} else {
		strct.bindings = bindings
	}
	return strct, nil
}

// StructLength - return the length (field count) of the <struct> object
func StructLength(strct *Object) int {
	return len(strct.bindings)
}

// Get - return the value for the key of the object. The Value() function is first called to
// handle typed instances of <struct>.
// This is called by the VM, when a keyword is used as a function.
func Get(obj *Object, key *Object) (*Object, error) {
	s := Value(obj)
	if s.Type != StructType {
		return nil, Error(ArgumentErrorKey, "get expected a <struct> argument, got a ", obj.Type)
	}
	return structGet(s, key), nil
}

func structGet(s *Object, key *Object) *Object {
	switch key.Type {
	case KeywordType, SymbolType, TypeType, StringType:
		result, ok := s.bindings[key]
		if ok {
			return result
		}
	}
	return Null
}

// Has returns whether the struct has the given key.
// Returns an error if the object is not a struct.
func Has(obj *Object, key *Object) (bool, error) {
	tmp, err := Get(obj, key)
	if err != nil || IsNull(tmp) {
		return false, err
	}
	return true, nil
}

// Put adds the given object to the struct with the given key
func Put(obj *Object, key *Object, val *Object) {
	obj.bindings[key] = val
}

// Unput deletes the given key from the struct
func Unput(obj *Object, key *Object) {
	delete(obj.bindings, key)
}

func sliceContains(slice []*Object, obj *Object) bool {
	for _, o := range slice {
		if o == obj {
			return true
		}
	}
	return false
}

func slicePut(bindings []*Object, key *Object, val *Object) []*Object {
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		if key == bindings[i] {
			bindings[i+1] = val
			return bindings
		}
	}
	return append(bindings, key, val)
}

func (vm *VM) validateKeywordArgList(args *Object, keys []*Object) (*Object, error) {
	tmp, err := vm.validateKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return ListFromValues(tmp), nil
}

func (vm *VM) validateKeywordArgBindings(args *Object, keys []*Object) ([]*Object, error) {
	count := ListLength(args)
	bindings := make([]*Object, 0, count)
	for args != EmptyList {
		key := Car(args)
		switch key.Type {
		case SymbolType:
			key = vm.Intern(key.text + ":")
			fallthrough
		case KeywordType:
			if !sliceContains(keys, key) {
				return nil, Error(ArgumentErrorKey, key, " bad keyword parameter. Allowed keys: ", keys)
			}
			args = args.cdr
			if args == EmptyList {
				return nil, Error(ArgumentErrorKey, key, " mismatched keyword/value pair in parameter")
			}
			bindings = slicePut(bindings, key, Car(args))
		case StructType:
			for sym, v := range key.bindings {
				if sliceContains(keys, sym) {
					bindings = slicePut(bindings, sym, v)
				}
			}
		default:
			return nil, Error(ArgumentErrorKey, "Not a keyword: ", key)
		}
		args = args.cdr
	}
	return bindings, nil
}

// StructEqual returns true if the object is equal to the argument
func StructEqual(s1 *Object, s2 *Object) bool {
	bindings1 := s1.bindings
	size := len(bindings1)
	bindings2 := s2.bindings
	if size == len(bindings2) {
		for k, v := range bindings1 {
			v2, ok := bindings2[k]
			if !ok {
				return false
			}
			if !Equal(v, v2) {
				return false
			}
		}
		return true
	}
	return false
}

func structToString(s *Object) string {
	var buf strings.Builder
	buf.WriteString("{")
	first := true
	for k, v := range s.bindings {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(k.String())
		buf.WriteString(" ")
		buf.WriteString(v.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func structToList(s *Object) (*Object, error) {
	result := EmptyList
	tail := EmptyList
	for k, v := range s.bindings {
		tmp := List(k, v)
		if result == EmptyList {
			result = List(tmp)
			tail = result
		} else {
			tail.cdr = List(tmp)
			tail = tail.cdr
		}
	}
	return result, nil
}

func structToArray(s *Object) *Object {
	size := len(s.bindings)
	el := make([]*Object, size)
	j := 0
	for k, v := range s.bindings {
		el[j] = Array(k, v)
		j++
	}
	return ArrayFromElements(el, size)
}

// StructKeys returns the keys of the struct
func StructKeys(s *Object) *Object {
	return structKeyList(s)
}

// StructValues returns a list of the values of the struct
func StructValues(s *Object) *Object {
	return structValueList(s)
}

func structKeyList(s *Object) *Object {
	result := EmptyList
	tail := EmptyList
	for key := range s.bindings {
		if result == EmptyList {
			result = List(key)
			tail = result
		} else {
			tail.cdr = List(key)
			tail = tail.cdr
		}
	}
	return result
}

func structValueList(s *Object) *Object {
	result := EmptyList
	tail := EmptyList
	for _, v := range s.bindings {
		if result == EmptyList {
			result = List(v)
			tail = result
		} else {
			tail.cdr = List(v)
			tail = tail.cdr
		}
	}
	return result
}

func listToStruct(lst *Object) (*Object, error) {
	strct := &Object{
		Type:     StructType,
		bindings: make(map[*Object]*Object),
	}
	for lst != EmptyList {
		k := lst.car
		lst = lst.cdr
		switch k.Type {
		case ListType:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			Put(strct, k.car, k.cdr.car)
		case ArrayType:
			elements := k.elements
			n := len(elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", elements[0])
			}
			Put(strct, elements[0], elements[1])
		default:
			if !IsValidStructKey(k) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if lst == EmptyList {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in list: ", k)
			}
			Put(strct, k, lst.car)
			lst = lst.cdr
		}
	}
	return strct, nil
}

func arrayToStruct(a *Object) (*Object, error) {
	count := len(a.elements)
	strct := &Object{
		Type:     StructType,
		bindings: make(map[*Object]*Object, count),
	}
	i := 0
	for i < count {
		k := a.elements[i]
		i++
		switch k.Type {
		case ListType:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			Put(strct, k.car, k.cdr.car)
		case ArrayType:
			elements := k.elements
			n := len(elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.elements[0])
			}
			Put(strct, elements[0], elements[1])
		case StringType, SymbolType, KeywordType, TypeType:
		default:
			if !IsValidStructKey(k) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if i == count {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in array: ", k)
			}
			Put(strct, k, a.elements[i])
			i++
		}
	}
	return strct, nil
}

// ToStruct converts an object to a struct
func ToStruct(obj *Object) (*Object, error) {
	val := Value(obj)
	switch val.Type {
	case StructType:
		return val, nil
	case ListType:
		return listToStruct(val)
	case ArrayType:
		return arrayToStruct(val)
	}
	return nil, Error(ArgumentErrorKey, "to-struct cannot accept argument of type ", obj.Type)
}
