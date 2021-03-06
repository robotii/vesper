package vesper

var defaultSymtab = initSymbolTable()

// Intern - internalize the name into the global symbol table
func (vm *VM) Intern(name string) *Object {
	sym, ok := vm.Symbols[name]
	if !ok {
		sym = &Object{text: name}
		if IsValidKeywordName(name) {
			sym.Type = KeywordType
		} else if IsValidTypeName(name) {
			sym.Type = TypeType
		} else if IsValidSymbolName(name) {
			sym.Type = SymbolType
		} else {
			panic("invalid symbol/type/keyword name passed to intern: " + name)
		}
		vm.Symbols[name] = sym
	}
	return sym
}

// IsValidSymbolName returns true if the string is a valid symbol
func IsValidSymbolName(name string) bool {
	return len(name) > 0
}

// IsValidTypeName returns true if the string is a typename
func IsValidTypeName(s string) bool {
	n := len(s)
	return n > 2 && s[0] == '<' && s[n-1] == '>'
}

// IsValidKeywordName returns true if the string is a keyword
func IsValidKeywordName(s string) bool {
	n := len(s)
	return n > 1 && s[n-1] == ':'
}

// ToKeyword converts the object to a keyword, if possible
func (vm *VM) ToKeyword(obj *Object) (*Object, error) {
	switch obj.Type {
	case KeywordType:
		return obj, nil
	case TypeType:
		return vm.Intern(obj.text[1:len(obj.text)-1] + ":"), nil
	case SymbolType:
		return vm.Intern(obj.text + ":"), nil
	case StringType:
		if IsValidKeywordName(obj.text) {
			return vm.Intern(obj.text), nil
		} else if IsValidSymbolName(obj.text) {
			return vm.Intern(obj.text + ":"), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-keyword expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.Type)
}

func typeNameString(s string) string {
	return s[1 : len(s)-1]
}

// TypeName returns the name of the type
func (vm *VM) TypeName(t *Object) (*Object, error) {
	if !IsType(t) {
		return nil, Error(ArgumentErrorKey, "type-name expected a <type>, got a ", t.Type)
	}
	return vm.Intern(typeNameString(t.text)), nil
}

// KeywordName returns the keyword as a string
func (vm *VM) KeywordName(t *Object) (*Object, error) {
	if !IsKeyword(t) {
		return nil, Error(ArgumentErrorKey, "keyword-name expected a <keyword>, got a ", t.Type)
	}
	return vm.unkeyworded(t)
}

func keywordNameString(s string) string {
	return s[:len(s)-1]
}

func unkeywordedString(k *Object) string {
	if IsKeyword(k) {
		return keywordNameString(k.text)
	}
	return k.text
}

func (vm *VM) unkeyworded(obj *Object) (*Object, error) {
	if IsSymbol(obj) {
		return obj, nil
	}
	if IsKeyword(obj) {
		return vm.Intern(keywordNameString(obj.text)), nil
	}
	return nil, Error(ArgumentErrorKey, "Expected <keyword> or <symbol>, got ", obj.Type)
}

// ToSymbol converts the object to a symbol, or returns an error if
// the object cannot be converted.
func (vm *VM) ToSymbol(obj *Object) (*Object, error) {
	switch obj.Type {
	case KeywordType:
		return vm.Intern(keywordNameString(obj.text)), nil
	case TypeType:
		return vm.Intern(typeNameString(obj.text)), nil
	case SymbolType:
		return obj, nil
	case StringType:
		if IsValidSymbolName(obj.text) {
			return vm.Intern(obj.text), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-symbol expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.Type)
}

func initSymbolTable() map[string]*Object {
	syms := make(map[string]*Object, 0)
	TypeType = &Object{text: "<type>"}
	TypeType.Type = TypeType //mutate to bootstrap type type
	syms[TypeType.text] = TypeType

	KeywordType = &Object{Type: TypeType, text: "<keyword>"}
	syms[KeywordType.text] = KeywordType

	SymbolType = &Object{Type: TypeType, text: "<symbol>"}
	syms[SymbolType.text] = SymbolType

	return syms
}

// Symbol creates a symbol from the given objects
func (vm *VM) Symbol(names []*Object) (*Object, error) {
	size := len(names)
	if size < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch o.Type {
		case StringType, SymbolType:
			s = o.text
		default:
			return nil, Error(ArgumentErrorKey, "symbol name component invalid: ", o)
		}
		name += s
	}
	return vm.Intern(name), nil
}
