package vesper

import (
	"strings"
)

// QuoteSymbol represents a quoted expression
var QuoteSymbol = defaultVM.Intern("quote")

// QuasiquoteSymbol represents a quasiquoted expression
var QuasiquoteSymbol = defaultVM.Intern("quasiquote")

// UnquoteSymbol represents an unquoted expression
var UnquoteSymbol = defaultVM.Intern("unquote")

// UnquoteSplicingSymbol represents an unquote-splicing expression
var UnquoteSplicingSymbol = defaultVM.Intern("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

// Cons - create a new list consisting of the first object and the rest of the list
func Cons(car *Object, cdr *Object) *Object {
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

// Caar - return the Car of the Car of the list
func Caar(lst *Object) *Object {
	return Car(Car(lst))
}

// Cadr - return the Car of the Cdr of the list
func Cadr(lst *Object) *Object {
	return Car(Cdr(lst))
}

// Cdar - return the Cdr of the Car of the list
func Cdar(lst *Object) *Object {
	return Car(Cdr(lst))
}

// Cddr - return the Cdr of the Cdr of the list
func Cddr(lst *Object) *Object {
	return Cdr(Cdr(lst))
}

// Cadar - return the Car of the Cdr of the Car of the list
func Cadar(lst *Object) *Object {
	return Car(Cdr(Car(lst)))
}

// Caddr - return the Car of the Cdr of the Cdr of the list
func Caddr(lst *Object) *Object {
	return Car(Cdr(Cdr(lst)))
}

// Cdddr - return the Cdr of the Cdr of the Cdr of the list
func Cdddr(lst *Object) *Object {
	return Cdr(Cdr(Cdr(lst)))
}

// Cadddr - return the Car of the Cdr of the Cdr of the Cdr of the list
func Cadddr(lst *Object) *Object {
	return Car(Cdr(Cdr(Cdr(lst))))
}

// Cddddr - return the Cdr of the Cdr of the Cdr of the Cdr of the list
func Cddddr(lst *Object) *Object {
	return Cdr(Cdr(Cdr(Cdr(lst))))
}

func initEmpty() *Object {
	return &Object{Type: ListType} //car and cdr are both nil
}

// ListEqual returns true if the object is equal to the argument
func ListEqual(lst *Object, a *Object) bool {
	for lst != EmptyList {
		if a == EmptyList {
			return false
		}
		if !Equal(lst.car, a.car) {
			return false
		}
		lst = lst.cdr
		a = a.cdr
	}
	if lst == a {
		return true
	}
	return false
}

func listToString(lst *Object) string {
	var buf strings.Builder
	if lst != EmptyList && lst.cdr != EmptyList && Cddr(lst) == EmptyList {
		switch lst.car {
		case QuoteSymbol:
			buf.WriteString("'")
			buf.WriteString(Cadr(lst).String())
			return buf.String()
		case QuasiquoteSymbol:
			buf.WriteString("`")
			buf.WriteString(Cadr(lst).String())
			return buf.String()
		case UnquoteSymbol:
			buf.WriteString("~")
			buf.WriteString(Cadr(lst).String())
			return buf.String()
		case UnquoteSplicingSymbol:
			buf.WriteString("~@")
			buf.WriteString(Cadr(lst).String())
			return buf.String()
		}
	}
	buf.WriteString("(")
	delim := ""
	for lst != EmptyList {
		buf.WriteString(delim)
		delim = " "
		buf.WriteString(lst.car.String())
		lst = lst.cdr
	}
	buf.WriteString(")")
	return buf.String()
}

// ListLength returns the length of the list
func ListLength(lst *Object) int {
	if lst == EmptyList {
		return 0
	}
	count := 1
	o := lst.cdr
	for o != EmptyList {
		count++
		o = o.cdr
	}
	return count
}

// MakeList creates a list of length count, with all values set to val
func MakeList(count int, val *Object) *Object {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = Cons(val, result)
	}
	return result
}

// ListFromValues creates a list from the given array
func ListFromValues(values []*Object) *Object {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = Cons(v, p)
	}
	return p
}

// List creates a new list from the arguments
func List(values ...*Object) *Object {
	return ListFromValues(values)
}

func listToArray(lst *Object) *Object {
	var elems []*Object
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return ArrayFromElementsNoCopy(elems)
}

// ToList - convert the argument to a List, if possible
func ToList(obj *Object) (*Object, error) {
	switch obj.Type {
	case ListType:
		return obj, nil
	case ArrayType:
		return ListFromValues(obj.elements), nil
	case StructType:
		return structToList(obj)
	case StringType:
		return stringToList(obj), nil
	}
	return nil, Error(ArgumentErrorKey, "to-list cannot accept ", obj.Type)
}

// Reverse reverses a linked list
func Reverse(lst *Object) *Object {
	rev := EmptyList
	for lst != EmptyList {
		rev = Cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

// Flatten flattens a list into a single list
func Flatten(lst *Object) *Object {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		item := lst.car
		switch item.Type {
		case ListType:
			item = Flatten(item)
		case ArrayType:
			litem, _ := ToList(item)
			item = Flatten(litem)
		default:
			item = List(item)
		}
		if tail == EmptyList {
			result = item
			tail = result
		} else {
			tail.cdr = item
		}
		for tail.cdr != EmptyList {
			tail = tail.cdr
		}
		lst = lst.cdr
	}
	return result
}

// Concat joins two lists
func Concat(seq1 *Object, seq2 *Object) (*Object, error) {
	rev := Reverse(seq1)
	if rev == EmptyList {
		return seq2, nil
	}
	lst := seq2
	for rev != EmptyList {
		lst = Cons(rev.car, lst)
		rev = rev.cdr
	}
	return lst, nil
}
