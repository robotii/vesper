package kana

// EmptyList - the value of (), terminates a list
var EmptyList = initEmpty()

func initEmpty() *Object {
	return &Object{Type: ListType} // both car and cdr are nil
}

// List - create a new list consisting of the first object and the rest of the list
func List(car *Object, cdr *Object) *Object {
	return &Object{
		Type: ListType,
		car:  car,
		cdr:  cdr,
	}
}

// Car - return the first object in a list
func Car(lst *Object) *Object {
	if lst == EmptyList {
		return Null
	}
	return lst.car
}

// Cdr - return the rest of the list
func Cdr(lst *Object) *Object {
	if lst == EmptyList {
		return lst
	}
	return lst.cdr
}
