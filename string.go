package vesper

import (
	"strings"
)

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

// AsStringValue - return the native string representation of the object, if possible
func AsStringValue(obj *Object) (string, error) {
	if !IsString(obj) {
		return "", Error(ArgumentErrorKey, StringType, obj)
	}
	return obj.text, nil
}

// ToString - convert the object to a string, if possible
func ToString(a *Object) (*Object, error) {
	switch a.Type {
	case NullType:
		return a, nil
	case CharacterType:
		return String(string([]rune{rune(a.fval)})), nil
	case StringType:
		return a, nil
	case BlobType:
		return String(string(BlobValue(a))), nil
	case SymbolType, KeywordType, TypeType:
		return String(a.text), nil
	case NumberType, BooleanType:
		return String(a.String()), nil
	case ArrayType:
		var chars []rune
		for _, c := range a.elements {
			if !IsCharacter(c) {
				return nil, Error(ArgumentErrorKey, "to-string: array element is not a <character>: ", c)
			}
			chars = append(chars, rune(c.fval))
		}
		return String(string(chars)), nil
	case ListType:
		var chars []rune
		for a != EmptyList {
			c := Car(a)
			if !IsCharacter(c) {
				return nil, Error(ArgumentErrorKey, "to-string: list element is not a <character>: ", c)
			}
			chars = append(chars, rune(c.fval))
			a = a.cdr
		}
		return String(string(chars)), nil
	default:
		return nil, Error(ArgumentErrorKey, "to-string: cannot convert argument to <string>: ", a)
	}
}

// StringLength - return the string length
func StringLength(s string) int {
	return len([]rune(s))
}

// EncodeString - return the encoded form of a string value
func EncodeString(s string) string {
	var buf []rune
	buf = append(buf, '"')
	for _, c := range s {
		// TODO: Coalesce append calls
		switch c {
		case '"':
			buf = append(buf, '\\')
			buf = append(buf, '"')
		case '\\':
			buf = append(buf, '\\')
			buf = append(buf, '\\')
		case '\n':
			buf = append(buf, '\\')
			buf = append(buf, 'n')
		case '\t':
			buf = append(buf, '\\')
			buf = append(buf, 't')
		case '\f':
			buf = append(buf, '\\')
			buf = append(buf, 'f')
		case '\b':
			buf = append(buf, '\\')
			buf = append(buf, 'b')
		case '\r':
			buf = append(buf, '\\')
			buf = append(buf, 'r')
		default:
			buf = append(buf, c)
		}
	}
	buf = append(buf, '"')
	return string(buf)
}

// Character - return a new <character> object
func Character(c rune) *Object {
	return &Object{
		Type: CharacterType,
		fval: float64(c),
	}
}

// ToCharacter - convert object to a <character> object, if possible
func ToCharacter(c *Object) (*Object, error) {
	switch c.Type {
	case CharacterType:
		return c, nil
	case StringType:
		if len(c.text) == 1 {
			for _, r := range c.text {
				return Character(r), nil
			}
		}
	case NumberType:
		r := rune(int(c.fval))
		return Character(r), nil
	}
	return nil, Error(ArgumentErrorKey, "Cannot convert to <character>: ", c)
}

// AsRuneValue - return the native rune representation of the character object, if possible
func AsRuneValue(c *Object) (rune, error) {
	if !IsCharacter(c) {
		return 0, Error(ArgumentErrorKey, "Not a <character>", c)
	}
	return rune(c.fval), nil
}

// StringCharacters - return a slice of <character> objects that represent the string
func StringCharacters(s *Object) []*Object {
	chars := make([]*Object, 0, len([]rune(s.text)))
	for _, c := range s.text {
		chars = append(chars, Character(c))
	}
	return chars
}

// StringRef - return the <character> object at the specified string index
func StringRef(s *Object, idx int) *Object {
	// utf8 requires a scan
	for i, r := range s.text {
		if i == idx {
			return Character(r)
		}
	}
	return Null
}

func stringToArray(s *Object) *Object {
	return Array(StringCharacters(s)...)
}

func stringToList(s *Object) *Object {
	return List(StringCharacters(s)...)
}

// StringSplit splits a string by the delimiters
func StringSplit(obj *Object, delims *Object) (*Object, error) {
	if !IsString(obj) {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 1, got ", obj)
	}
	if !IsString(delims) {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 2, got ", delims)
	}
	lst := EmptyList
	tail := EmptyList
	for _, s := range strings.Split(obj.text, delims.text) {
		if lst == EmptyList {
			lst = List(String(s))
			tail = lst
		} else {
			tail.cdr = List(String(s))
			tail = tail.cdr
		}
	}
	return lst, nil
}

// StringJoin joins a string with the given delimters
func StringJoin(seq *Object, delims *Object) (*Object, error) {
	if !IsString(delims) {
		return nil, Error(ArgumentErrorKey, "join expected a <string> for argument 2, got ", delims)
	}
	switch seq.Type {
	case ListType:
		result := ""
		for seq != EmptyList {
			o := seq.car
			if o != EmptyString && o != Null {
				if result != "" {
					result += delims.text
				}
				result += o.String()
			}
			seq = seq.cdr
		}
		return String(result), nil
	case ArrayType:
		result := ""
		elements := seq.elements
		count := len(elements)
		for i := 0; i < count; i++ {
			o := elements[i]
			if o != EmptyString && o != Null {
				if result != "" {
					result += delims.text
				}
				result += o.String()
			}
		}
		return String(result), nil
	default:
		return nil, Error(ArgumentErrorKey, "join expected a <list> or <array> for argument 1, got a ", seq.Type)
	}
}
