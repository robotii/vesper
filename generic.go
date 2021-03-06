package vesper

var (
	// GenfnsSymbol used to define generic Function
	GenfnsSymbol = defaultVM.Intern("*genfns*")
	// MethodsKeyword used to define methods
	MethodsKeyword = defaultVM.Intern("methods:")
)

func (vm *VM) methodSignature(formalArgs *Object) (*Object, error) {
	sig := ""
	for formalArgs != EmptyList {
		s := formalArgs.car //might be a symbol, might be a list
		tname := ""
		if s.Type == ListType { //specialized
			t := Cadr(s)
			if t.Type != TypeType {
				return nil, Error(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
			}
			tname = t.text
		} else if s.Type == SymbolType { //unspecialized
			tname = "<any>"
		} else {
			return nil, Error(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
		}
		sig += tname
		formalArgs = formalArgs.cdr
	}
	return vm.Intern(sig), nil
}

func arglistSignature(args []*Object) string {
	sig := ""
	for _, arg := range args {
		sig += arg.Type.text
	}
	return sig
}

func signatureCombos(argtypes []*Object) []string {
	switch len(argtypes) {
	case 0:
		return []string{}
	case 1:
		return []string{argtypes[0].text, AnyType.text}
	default:
		//get the combinations of the tail, and concat both the type and <any> onto each of those combos
		rest := signatureCombos(argtypes[1:]) // ["<number>" "<any>"]
		result := make([]string, 0, len(rest)*2)
		this := argtypes[0]
		for _, s := range rest {
			result = append(result, this.text+s)
		}
		for _, s := range rest {
			result = append(result, AnyType.text+s)
		}
		return result
	}
}

var cachedSigs = make(map[string][]*Object)

func (vm *VM) arglistSignatures(args []*Object) []*Object {
	key := arglistSignature(args)
	sigs, ok := cachedSigs[key]
	if !ok {
		var argtypes []*Object
		for _, arg := range args {
			argtypes = append(argtypes, arg.Type)
		}
		stringSigs := signatureCombos(argtypes)
		sigs = make([]*Object, 0, len(stringSigs))
		for _, sig := range stringSigs {
			sigs = append(sigs, vm.Intern(sig))
		}
		cachedSigs[key] = sigs
	}
	return sigs
}

func (vm *VM) getfn(sym *Object, args []*Object) (*Object, error) {
	sigs := vm.arglistSignatures(args)
	gfs := GetGlobal(GenfnsSymbol)
	if gfs != nil && gfs.Type == StructType {
		gf := structGet(gfs, sym)
		if gf == Null {
			return nil, Error(ErrorKey, "Not a generic function: ", sym)
		}
		methods := structGet(Value(gf), MethodsKeyword)
		if methods.Type == StructType {
			for _, sig := range sigs {
				fun := structGet(methods, sig)
				if fun != Null {
					return fun, nil
				}
			}
		}
	}
	return nil, Error(ErrorKey, "Generic function ", sym, ", has no matching method for: ", args)
}
