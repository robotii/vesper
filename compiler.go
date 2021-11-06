package vesper

// Compile - compile the source into a code object.
func (vm *VM) Compile(expr *Object) (*Object, error) {
	target := MakeCode(vm, 0, nil, nil, "")
	err := vm.compileExpr(target, EmptyList, expr, false, false, "")
	if err != nil {
		return nil, err
	}
	target.code.emitReturn()
	return target, nil
}

func calculateLocation(sym *Object, env *Object) (int, int, bool) {
	i := 0
	for env != EmptyList {
		j := 0
		ee := Car(env)
		for ee != EmptyList {
			if Car(ee) == sym {
				return i, j, true
			}
			j++
			ee = Cdr(ee)
		}
		i++
		env = Cdr(env)
	}
	return -1, -1, false
}

func (vm *VM) compileSelfEvalLiteral(target *Object, expr *Object, isTail bool, ignoreResult bool) error {
	if !ignoreResult {
		target.code.emitLiteral(vm.putConstant(expr))
		if isTail {
			target.code.emitReturn()
		}
	}
	return nil
}

func (vm *VM) compileSymbol(target *Object, env *Object, expr *Object, isTail bool, ignoreResult bool) error {
	if vm.GetMacro(expr) != nil {
		return Error(vm.Intern("macro-error"), "Cannot use macro as a value: ", expr)
	}
	if i, j, ok := calculateLocation(expr, env); ok {
		target.code.emitLocal(i, j)
	} else {
		target.code.emitGlobal(vm.putConstant(expr))
	}
	if ignoreResult {
		target.code.emitPop()
	} else if isTail {
		target.code.emitReturn()
	}
	return nil
}

func (vm *VM) compileQuote(target *Object, expr *Object, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen != 2 {
		return Error(SyntaxErrorKey, expr)
	}
	if !ignoreResult {
		target.code.emitLiteral(vm.putConstant(Cadr(expr)))
		if isTail {
			target.code.emitReturn()
		}
	}
	return nil
}

func (vm *VM) compileDef(target *Object, env *Object, lst *Object, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen < 3 {
		return Error(SyntaxErrorKey, lst)
	}
	sym := Cadr(lst)
	val := Caddr(lst)
	err := vm.compileExpr(target, env, val, false, false, sym.String())
	if err == nil {
		target.code.emitDefGlobal(vm.putConstant(sym))
		if ignoreResult {
			target.code.emitPop()
		} else if isTail {
			target.code.emitReturn()
		}
	}
	return err
}

func (vm *VM) compileUndef(target *Object, lst *Object, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen != 2 {
		return Error(SyntaxErrorKey, lst)
	}
	sym := Cadr(lst)
	if !IsSymbol(sym) {
		return Error(SyntaxErrorKey, lst)
	}
	target.code.emitUndefGlobal(vm.putConstant(sym))
	if ignoreResult {
	} else {
		target.code.emitLiteral(vm.putConstant(sym))
		if isTail {
			target.code.emitReturn()
		}
	}
	return nil
}

func (vm *VM) compileMacro(target *Object, env *Object, expr *Object, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen != 3 {
		return Error(SyntaxErrorKey, expr)
	}
	var sym = Cadr(expr)
	if !IsSymbol(sym) {
		return Error(SyntaxErrorKey, expr)
	}
	err := vm.compileExpr(target, env, Caddr(expr), false, false, sym.String())
	if err != nil {
		return err
	}
	target.code.emitDefMacro(vm.putConstant(sym))
	if ignoreResult {
		target.code.emitPop()
	} else if isTail {
		target.code.emitReturn()
	}
	return err
}

func (vm *VM) compileSet(target *Object, env *Object, lst *Object, isTail bool, ignoreResult bool, context string, lstlen int) error {
	if lstlen != 3 {
		return Error(SyntaxErrorKey, lst)
	}
	var sym = Cadr(lst)
	if !IsSymbol(sym) {
		return Error(SyntaxErrorKey, lst)
	}
	err := vm.compileExpr(target, env, Caddr(lst), false, false, context)
	if err != nil {
		return err
	}
	if i, j, ok := calculateLocation(sym, env); ok {
		target.code.emitSetLocal(i, j)
	} else {
		target.code.emitDefGlobal(vm.putConstant(sym)) // fix, should be SetGlobal
	}
	if ignoreResult {
		target.code.emitPop()
	} else if isTail {
		target.code.emitReturn()
	}
	return nil
}

func (vm *VM) compileList(target *Object, env *Object, expr *Object, isTail bool, ignoreResult bool, context string) error {
	if expr == EmptyList {
		if !ignoreResult {
			target.code.emitLiteral(vm.putConstant(expr))
			if isTail {
				target.code.emitReturn()
			}
		}
		return nil
	}
	lst := expr
	lstlen := ListLength(lst)
	if lstlen == 0 {
		return Error(SyntaxErrorKey, lst)
	}
	fn := Car(lst)
	switch fn {
	case vm.Intern("quote"):
		// (quote <datum>)
		return vm.compileQuote(target, expr, isTail, ignoreResult, lstlen)
	case vm.Intern("do"): // a sequence of expressions, for side-effect only
		// (do <expr> ...)
		return vm.compileSequence(target, env, Cdr(lst), isTail, ignoreResult, context)
	case vm.Intern("if"):
		// (if <pred> <Consequent>)
		// (if <pred> <Consequent> <antecedent>)
		if lstlen == 3 || lstlen == 4 {
			return vm.compileIfElse(target, env, Cadr(expr), Caddr(expr), Cdddr(expr), isTail, ignoreResult, context)
		}
		return Error(SyntaxErrorKey, expr)
	case vm.Intern("def"):
		// (def <name> <val>)
		return vm.compileDef(target, env, expr, isTail, ignoreResult, lstlen)
	case vm.Intern("undef"):
		// (undef <name>)
		return vm.compileUndef(target, expr, isTail, ignoreResult, lstlen)
	case vm.Intern("defmacro"):
		// (defmacro <name> (fn args & body))
		return vm.compileMacro(target, env, expr, isTail, ignoreResult, lstlen)
	case vm.Intern("fn"):
		// (fn ()  <expr> ...)
		// (fn (sym ...)  <expr> ...) ;; binds arguments to successive syms
		// (fn (sym ... & rsym)  <expr> ...) ;; all args after the & are collected and bound to rsym
		// (fn (sym ... [sym sym])  <expr> ...) ;; all args up to the array are required, the rest are optional
		// (fn (sym ... [(sym val) sym])  <expr> ...) ;; default values can be provided to optional args
		// (fn (sym ... {sym: def sym: def})  <expr> ...) ;; required args, then keyword args
		// (fn (& sym)  <expr> ...) ;; all args in a list, bound to sym. Same as the following form.
		// (fn sym <expr> ...) ;; all args in a list, bound to sym
		if lstlen < 3 {
			return Error(SyntaxErrorKey, expr)
		}
		body := Cddr(lst)
		args := Cadr(lst)
		return vm.compileFn(target, env, args, body, isTail, ignoreResult, context)
	case vm.Intern("set!"):
		return vm.compileSet(target, env, expr, isTail, ignoreResult, context, lstlen)
	case vm.Intern("code"):
		return target.code.loadOps(vm, Cdr(expr))
	case vm.Intern("use"):
		return vm.compileUse(target, Cdr(lst))
	default:
		fn, args := vm.optimizeFuncall(fn, Cdr(lst))
		return vm.compileFuncall(target, env, fn, args, isTail, ignoreResult, context)
	}
}

func (vm *VM) compileArray(target *Object, env *Object, expr *Object, isTail bool, ignoreResult bool, context string) error {
	// array literal: the elements are evaluated
	vlen := len(expr.elements)
	for i := vlen - 1; i >= 0; i-- {
		obj := expr.elements[i]
		err := vm.compileExpr(target, env, obj, false, false, context)
		if err != nil {
			return err
		}
	}
	if !ignoreResult {
		target.code.emitArray(vlen)
		if isTail {
			target.code.emitReturn()
		}
	}
	return nil
}

func (vm *VM) compileStruct(target *Object, env *Object, expr *Object, isTail bool, ignoreResult bool, context string) error {
	//struct literal: the elements are evaluated
	vlen := len(expr.bindings) * 2
	vals := make([]*Object, 0, vlen)
	for k, v := range expr.bindings {
		vals = append(vals, k, v)
	}
	for i := vlen - 1; i >= 0; i-- {
		obj := vals[i]
		err := vm.compileExpr(target, env, obj, false, false, context)
		if err != nil {
			return err
		}
	}
	if !ignoreResult {
		target.code.emitStruct(vlen)
		if isTail {
			target.code.emitReturn()
		}
	}
	return nil
}

func (vm *VM) compileExpr(target *Object, env *Object, expr *Object, isTail bool, ignoreResult bool, context string) error {
	switch {
	case IsKeyword(expr) || IsType(expr):
		return vm.compileSelfEvalLiteral(target, expr, isTail, ignoreResult)
	case IsSymbol(expr):
		return vm.compileSymbol(target, env, expr, isTail, ignoreResult)
	case IsList(expr):
		return vm.compileList(target, env, expr, isTail, ignoreResult, context)
	case IsArray(expr):
		return vm.compileArray(target, env, expr, isTail, ignoreResult, context)
	case IsStruct(expr):
		return vm.compileStruct(target, env, expr, isTail, ignoreResult, context)
	default:
		if !ignoreResult {
			target.code.emitLiteral(vm.putConstant(expr))
			if isTail {
				target.code.emitReturn()
			}
		}
		return nil
	}
}

func (vm *VM) compileFn(target *Object, env *Object, args *Object, body *Object, isTail bool, ignoreResult bool, context string) error {
	argc := 0
	var syms []*Object
	var defaults []*Object
	var keys []*Object
	tmp := args
	rest := false
	if !IsSymbol(args) {
		if IsArray(tmp) {
			// Allow clojure-style parameter lists for convenience
			// (fn [sym ...] <expr>)
			tmp, _ = ToList(tmp)
		}
		for tmp != nil && tmp.car != nil && tmp != EmptyList {
			a := Car(tmp)
			if IsArray(a) {
				if Cdr(tmp) != EmptyList {
					return Error(SyntaxErrorKey, tmp)
				}
				defaults = make([]*Object, 0, len(a.elements))
				for _, sym := range a.elements {
					def := Null
					if IsList(sym) {
						def = Cadr(sym)
						sym = Car(sym)
					}
					if !IsSymbol(sym) {
						return Error(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					defaults = append(defaults, def)
				}
				tmp = EmptyList
				break
			} else if IsStruct(a) {
				if Cdr(tmp) != EmptyList {
					return Error(SyntaxErrorKey, tmp)
				}
				slen := len(a.bindings)
				defaults = make([]*Object, 0, slen)
				keys = make([]*Object, 0, slen)
				for sym, defValue := range a.bindings {
					if IsList(sym) && Car(sym) == vm.Intern("quote") && Cdr(sym) != EmptyList {
						sym = Cadr(sym)
					} else {
						var err error
						sym, err = vm.unkeyworded(sym)
						if err != nil {
							return Error(SyntaxErrorKey, tmp)
						}
					}
					if !IsSymbol(sym) {
						return Error(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					keys = append(keys, sym)
					defaults = append(defaults, defValue)
				}
				tmp = EmptyList
				break
			} else if !IsSymbol(a) {
				return Error(SyntaxErrorKey, tmp)
			}
			if a == vm.Intern("&") {
				rest = true
			} else {
				if rest {
					syms = append(syms, a)
					defaults = make([]*Object, 0)
					tmp = EmptyList
					break
				}
				argc++
				syms = append(syms, a)
			}
			tmp = Cdr(tmp)
		}
	}
	if tmp != EmptyList {
		if IsSymbol(tmp) {
			syms = append(syms, tmp)
			defaults = make([]*Object, 0)
		} else {
			return Error(SyntaxErrorKey, tmp)
		}
	}
	args = ListFromValues(syms)
	newEnv := Cons(args, env)
	fnCode := MakeCode(vm, argc, defaults, keys, context)
	err := vm.compileSequence(fnCode, newEnv, body, true, false, context)
	if err == nil {
		if !ignoreResult {
			target.code.emitClosure(vm.putConstant(fnCode))
			if isTail {
				target.code.emitReturn()
			}
		}
	}
	return err
}

func (vm *VM) compileSequence(target *Object, env *Object, exprs *Object, isTail bool, ignoreResult bool, context string) error {
	if exprs != EmptyList {
		for Cdr(exprs) != EmptyList {
			err := vm.compileExpr(target, env, Car(exprs), false, true, context)
			if err != nil {
				return err
			}
			exprs = Cdr(exprs)
		}
		return vm.compileExpr(target, env, Car(exprs), isTail, ignoreResult, context)
	}
	return Error(SyntaxErrorKey, Cons(vm.Intern("do"), exprs))
}

func (vm *VM) optimizeFuncall(fn *Object, args *Object) (*Object, *Object) {
	size := ListLength(args)
	if size == 2 {
		switch fn {
		case vm.Intern("+"):
			if Equal(One, Car(args)) {
				return vm.Intern("inc"), Cdr(args)
			} else if Equal(One, Cadr(args)) {
				return vm.Intern("inc"), List(Car(args))
			}
		case vm.Intern("-"):
			if Equal(One, Cadr(args)) {
				return vm.Intern("dec"), List(Car(args))
			}
		}
	}
	return fn, args
}

func (vm *VM) compileFuncall(target *Object, env *Object, fn *Object, args *Object, isTail bool, ignoreResult bool, context string) error {
	argc := ListLength(args)
	if argc < 0 {
		return Error(SyntaxErrorKey, Cons(fn, args))
	}
	err := vm.compileArgs(target, env, args, context)
	if err != nil {
		return err
	}
	err = vm.compileExpr(target, env, fn, false, false, context)
	if err != nil {
		return err
	}
	if isTail {
		target.code.emitTailCall(argc)
	} else {
		target.code.emitCall(argc)
		if ignoreResult {
			target.code.emitPop()
		}
	}
	return nil
}

func (vm *VM) compileArgs(target *Object, env *Object, args *Object, context string) error {
	if args != EmptyList {
		err := vm.compileArgs(target, env, Cdr(args), context)
		if err != nil {
			return err
		}
		return vm.compileExpr(target, env, Car(args), false, false, context)
	}
	return nil
}

func (vm *VM) compileIfElse(target *Object, env *Object, predicate *Object, Consequent *Object, antecedentOptional *Object, isTail bool, ignoreResult bool, context string) error {
	antecedent := Null
	if antecedentOptional != EmptyList {
		antecedent = Car(antecedentOptional)
	}
	err := vm.compileExpr(target, env, predicate, false, false, context)
	if err != nil {
		return err
	}
	loc1 := target.code.emitJumpFalse(0) //returns the location just *after* the jump. setJumpLocation knows this.
	err = vm.compileExpr(target, env, Consequent, isTail, ignoreResult, context)
	if err != nil {
		return err
	}
	loc2 := 0
	if !isTail {
		loc2 = target.code.emitJump(0)
	}
	target.code.setJumpLocation(loc1)
	err = vm.compileExpr(target, env, antecedent, isTail, ignoreResult, context)
	if err == nil {
		if !isTail {
			target.code.setJumpLocation(loc2)
		}
	}
	return err
}

func (vm *VM) compileUse(target *Object, rest *Object) error {
	lstlen := ListLength(rest)
	if lstlen != 1 {
		//to do: other options for use.
		return Error(SyntaxErrorKey, Cons(vm.Intern("use"), rest))
	}
	sym := Car(rest)
	if !IsSymbol(sym) {
		return Error(SyntaxErrorKey, rest)
	}
	symIdx := vm.putConstant(sym)
	target.code.emitUse(symIdx)
	return nil
}
