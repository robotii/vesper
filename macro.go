package vesper

import (
	"fmt"
)

// Macro a structure to hold Macro values
type Macro struct {
	Name     *Object
	Expander *Object // a function of one argument
}

// NewMacro - create a new Macro
func NewMacro(name *Object, expander *Object) *Macro {
	return &Macro{name, expander}
}

func (mac *Macro) String() string {
	return fmt.Sprintf("(macro %v %v)", mac.Name, mac.Expander)
}

// Macroexpand - return the expansion of all macros in the object and return the result
func (vm *VM) Macroexpand(expr *Object) (*Object, error) {
	return vm.macroexpandObject(expr)
}

func (vm *VM) macroexpandObject(expr *Object) (*Object, error) {
	if IsList(expr) {
		if expr != EmptyList {
			return vm.macroexpandList(expr)
		}
	}
	return expr, nil
}

func (vm *VM) macroexpandList(expr *Object) (*Object, error) {
	if expr == EmptyList {
		return expr, nil
	}
	lst := expr
	fn := Car(lst)
	head := fn
	if IsSymbol(fn) {
		result, err := vm.expandPrimitive(fn, lst)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		head = fn
	} else if IsList(fn) {
		expanded, err := vm.macroexpandList(fn)
		if err != nil {
			return nil, err
		}
		head = expanded
	}
	tail, err := vm.expandSequence(Cdr(expr))
	if err != nil {
		return nil, err
	}
	return Cons(head, tail), nil
}

func (vm *VM) expand(mac *Macro, expr *Object) (*Object, error) {
	expander := mac.Expander
	if expander.Type == FunctionType {
		if expander.code != nil {
			if expander.code.argc == 1 {
				expanded, err := vm.execCompileTime(expander.code, expr)
				if err == nil {
					if IsList(expanded) {
						return vm.macroexpandObject(expanded)
					}
					return expanded, err
				}
				return nil, err
			}
		} else if expander.primitive != nil {
			args := []*Object{expr}
			expanded, err := expander.primitive.fun(args)
			if err == nil {
				return vm.macroexpandObject(expanded)
			}
			return nil, err
		}
	}
	return nil, Error(MacroErrorKey, "Bad macro expander function: ", expander)
}

func (vm *VM) expandSequence(seq *Object) (*Object, error) {
	var result []*Object
	if seq == nil {
		panic("Whoops: should be (), not nil!")
	}
	for seq != EmptyList {
		item := Car(seq)
		if IsList(item) {
			expanded, err := vm.macroexpandList(item)
			if err != nil {
				return nil, err
			}
			result = append(result, expanded)
		} else {
			result = append(result, item)
		}
		seq = Cdr(seq)
	}
	lst := ListFromValues(result)
	if seq != EmptyList {
		tmp := Cons(seq, EmptyList)
		return Concat(lst, tmp)
	}
	return lst, nil
}

func (vm *VM) expandIf(expr *Object) (*Object, error) {
	i := ListLength(expr)
	if i == 4 {
		tmp, err := vm.expandSequence(Cdr(expr))
		if err != nil {
			return nil, err
		}
		return Cons(Car(expr), tmp), nil
	} else if i == 3 {
		tmp := List(Cadr(expr), Caddr(expr), Null)
		tmp, err := vm.expandSequence(tmp)
		if err != nil {
			return nil, err
		}
		return Cons(Car(expr), tmp), nil
	} else {
		return nil, Error(SyntaxErrorKey, expr)
	}
}

func expandUndef(expr *Object) (*Object, error) {
	if ListLength(expr) != 2 || !IsSymbol(Cadr(expr)) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	return expr, nil
}

// (defn f (x) (+ 1 x))
//  ->
// (def f (fn (x) (+ 1 x)))
func (vm *VM) expandDefn(expr *Object) (*Object, error) {
	exprLen := ListLength(expr)
	if exprLen >= 4 {
		name := Cadr(expr)
		if IsSymbol(name) {
			args := Caddr(expr)
			body, err := vm.expandSequence(Cdddr(expr))
			if err != nil {
				return nil, err
			}
			tmp, err := vm.expandFn(Cons(vm.Intern("fn"), Cons(args, body)))
			if err != nil {
				return nil, err
			}
			return List(vm.Intern("def"), name, tmp), nil
		}
	}
	return nil, Error(SyntaxErrorKey, expr)
}

func (vm *VM) expandDefmacro(expr *Object) (*Object, error) {
	exprLen := ListLength(expr)
	if exprLen >= 4 {
		name := Cadr(expr)
		if IsSymbol(name) {
			args := Caddr(expr)
			body, err := vm.expandSequence(Cdddr(expr))
			if err != nil {
				return nil, err
			}
			//(fn (expr) (apply xxx
			tmp, err := vm.expandFn(Cons(vm.Intern("fn"), Cons(args, body))) //this is the expander with special args\
			if err != nil {
				return nil, err
			}
			sym := vm.Intern("expr")
			tmp, err = vm.expandFn(List(vm.Intern("fn"), List(sym), List(vm.Intern("apply"), tmp, List(vm.Intern("cdr"), sym))))
			if err != nil {
				return nil, err
			}
			return List(vm.Intern("defmacro"), name, tmp), nil
		}
	}
	return nil, Error(SyntaxErrorKey, expr)
}

//(defmacro (defmacro expr)
//  `(defmacro ~(cadr expr) (fn (expr) (apply (fn ~(caddr expr) ~@(cdddr expr)) (cdr expr)))))

func (vm *VM) expandDef(expr *Object) (*Object, error) {
	exprLen := ListLength(expr)
	if exprLen != 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	name := Cadr(expr)
	if !IsSymbol(name) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	if exprLen > 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body := Caddr(expr)
	if !IsList(body) {
		return expr, nil
	}
	val, err := vm.macroexpandList(body)
	if err != nil {
		return nil, err
	}
	return List(Car(expr), name, val), nil
}

func (vm *VM) expandFn(expr *Object) (*Object, error) {
	exprLen := ListLength(expr)
	if exprLen < 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body, err := vm.expandSequence(Cddr(expr))
	if err != nil {
		return nil, err
	}
	bodyLen := ListLength(body)
	if bodyLen > 0 {
		tmp := body
		if IsList(tmp) && Caar(tmp) == vm.Intern("def") || Caar(tmp) == vm.Intern("defmacro") {
			bindings := EmptyList
			for Caar(tmp) == vm.Intern("def") || Caar(tmp) == vm.Intern("defmacro") {
				if Caar(tmp) == vm.Intern("defmacro") {
					return nil, Error(MacroErrorKey, "macros can only be defined at top level")
				}
				def, err := vm.expandDef(Car(tmp))
				if err != nil {
					return nil, err
				}
				bindings = Cons(Cdr(def), bindings)
				tmp = Cdr(tmp)
			}
			bindings = Reverse(bindings)
			tmp = Cons(vm.Intern("letrec"), Cons(bindings, tmp)) //scheme specifies letrec*
			tmp2, err := vm.macroexpandList(tmp)
			return List(Car(expr), Cadr(expr), tmp2), err
		}
	}
	args := Cadr(expr)
	return Cons(Car(expr), Cons(args, body)), nil
}

func (vm *VM) expandSetBang(expr *Object) (*Object, error) {
	exprLen := ListLength(expr)
	if exprLen != 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	var val = Caddr(expr)
	if IsList(val) {
		v, err := vm.macroexpandList(val)
		if err != nil {
			return nil, err
		}
		val = v
	}
	return List(Car(expr), Cadr(expr), val), nil
}

func (vm *VM) expandPrimitive(fn *Object, expr *Object) (*Object, error) {
	switch fn {
	case vm.Intern("quote"):
		return expr, nil
	case vm.Intern("do"):
		return vm.expandSequence(expr)
	case vm.Intern("if"):
		return vm.expandIf(expr)
	case vm.Intern("def"):
		return vm.expandDef(expr)
	case vm.Intern("undef"):
		return expandUndef(expr)
	case vm.Intern("defn"):
		return vm.expandDefn(expr)
	case vm.Intern("defmacro"):
		return vm.expandDefmacro(expr)
	case vm.Intern("fn"):
		return vm.expandFn(expr)
	case vm.Intern("set!"):
		return vm.expandSetBang(expr)
	case vm.Intern("lap"):
		return expr, nil
	case vm.Intern("use"):
		return expr, nil
	default:
		macro := vm.GetMacro(fn)
		if macro != nil {
			tmp, err := vm.expand(macro, expr)
			return tmp, err
		}
		return nil, nil
	}
}

func (vm *VM) crackLetrecBindings(bindings *Object, tail *Object) (*Object, *Object, bool) {
	var names []*Object
	inits := EmptyList
	for bindings != EmptyList {
		if IsList(bindings) {
			tmp := Car(bindings)
			if IsArray(tmp) {
				tmp, _ = ToList(tmp)
			}
			if IsList(tmp) {
				name := Car(tmp)
				if IsSymbol(name) {
					names = append(names, name)
				} else {
					return nil, nil, false
				}
				if IsList(Cdr(tmp)) {
					inits = Cons(Cons(vm.Intern("set!"), tmp), inits)
				} else {
					return nil, nil, false
				}
			} else {
				return nil, nil, false
			}

		} else {
			return nil, nil, false
		}
		bindings = Cdr(bindings)
	}
	inits = Reverse(inits)
	head := inits
	for inits.cdr != EmptyList {
		inits = inits.cdr
	}
	inits.cdr = tail
	return ListFromValues(names), head, true
}

func (vm *VM) expandLetrec(expr *Object) (*Object, error) {
	body := Cddr(expr)
	if body == EmptyList {
		return nil, Error(SyntaxErrorKey, expr)
	}
	bindings := Cadr(expr)
	if IsArray(bindings) {
		bindings, _ = ToList(bindings)
	}
	if !IsList(bindings) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	names, body, ok := vm.crackLetrecBindings(bindings, body)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	code, err := vm.macroexpandList(Cons(vm.Intern("fn"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	values := MakeList(ListLength(names), Null)
	return Cons(code, values), nil
}

func (vm *VM) crackLetBindings(bindings *Object) (*Object, *Object, bool) {
	var names []*Object
	var values []*Object
	for bindings != EmptyList {
		tmp := Car(bindings)
		if IsArray(tmp) {
			tmp, _ = ToList(tmp)
		}
		if IsList(tmp) {
			name := Car(tmp)
			if IsSymbol(name) {
				names = append(names, name)
				tmp2 := Cdr(tmp)
				if tmp2 != EmptyList {
					val, err := vm.macroexpandObject(Car(tmp2))
					if err == nil {
						values = append(values, val)
						bindings = Cdr(bindings)
						continue
					}
				}
			}
		}
		return nil, nil, false
	}
	return ListFromValues(names), ListFromValues(values), true
}

func (vm *VM) expandLet(expr *Object) (*Object, error) {
	// (let () expr ...) -> (do expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((fn (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (fn (label) expr
	if IsSymbol(Cadr(expr)) {
		return vm.expandNamedLet(expr)
	}
	bindings := Cadr(expr)
	if IsArray(bindings) {
		bindings, _ = ToList(bindings)
	}
	if !IsList(bindings) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	names, values, ok := vm.crackLetBindings(bindings)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body := Cddr(expr)
	if body == EmptyList {
		return nil, Error(SyntaxErrorKey, expr)
	}
	code, err := vm.macroexpandList(Cons(vm.Intern("fn"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	return Cons(code, values), nil
}

func (vm *VM) expandNamedLet(expr *Object) (*Object, error) {
	name := Cadr(expr)
	bindings := Caddr(expr)
	if IsArray(bindings) {
		bindings, _ = ToList(bindings)
	}
	if !IsList(bindings) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	names, values, ok := vm.crackLetBindings(bindings)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body := Cdddr(expr)
	tmp := List(vm.Intern("letrec"), List(List(name, Cons(vm.Intern("fn"), Cons(names, body)))), Cons(name, values))
	return vm.macroexpandList(tmp)
}

func (vm *VM) nextCondClause(expr *Object, clauses *Object, count int) (*Object, error) {
	var result *Object
	var err error
	tmpsym := vm.Intern("__tmp__")
	ifsym := vm.Intern("if")
	elsesym := vm.Intern("else")
	letsym := vm.Intern("let")
	dosym := vm.Intern("do")

	clause0 := Car(clauses)
	next := Cdr(clauses)
	clause1 := Car(next)

	if count == 2 {
		if !IsList(clause1) {
			return nil, Error(SyntaxErrorKey, expr)
		}
		if elsesym == Car(clause1) {
			if Cadr(clause0) == vm.Intern("=>") {
				if ListLength(clause0) != 3 {
					return nil, Error(SyntaxErrorKey, expr)
				}
				result = List(letsym, List(List(tmpsym, Car(clause0))), List(ifsym, tmpsym, List(Caddr(clause0), tmpsym), Cons(dosym, Cdr(clause1))))
			} else {
				result = List(ifsym, Car(clause0), Cons(dosym, Cdr(clause0)), Cons(dosym, Cdr(clause1)))
			}
		} else {
			if Cadr(clause1) == vm.Intern("=>") {
				if ListLength(clause1) != 3 {
					return nil, Error(SyntaxErrorKey, expr)
				}
				result = List(letsym, List(List(tmpsym, Car(clause1))), List(ifsym, tmpsym, List(Caddr(clause1), tmpsym), clause1))
			} else {
				result = List(ifsym, Car(clause1), Cons(dosym, Cdr(clause1)))
			}
			if Cadr(clause0) == vm.Intern("=>") {
				if ListLength(clause0) != 3 {
					return nil, Error(SyntaxErrorKey, expr)
				}
				result = List(letsym, List(List(tmpsym, Car(clause0))), List(ifsym, tmpsym, List(Caddr(clause0), tmpsym), result))
			} else {
				result = List(ifsym, Car(clause0), Cons(dosym, Cdr(clause0)), result)
			}
		}
	} else {
		result, err = vm.nextCondClause(expr, next, count-1)
		if err != nil {
			return nil, err
		}
		if Cadr(clause0) == vm.Intern("=>") {
			if ListLength(clause0) != 3 {
				return nil, Error(SyntaxErrorKey, expr)
			}
			result = List(letsym, List(List(tmpsym, Car(clause0))), List(ifsym, tmpsym, List(Caddr(clause0), tmpsym), result))
		} else {
			result = List(ifsym, Car(clause0), Cons(dosym, Cdr(clause0)), result)
		}
	}
	return vm.macroexpandObject(result)
}

func (vm *VM) expandCond(expr *Object) (*Object, error) {
	i := ListLength(expr)
	if i < 2 {
		return nil, Error(SyntaxErrorKey, expr)
	} else if i == 2 {
		tmp := Cadr(expr)
		if Car(tmp) == vm.Intern("else") {
			tmp = Cons(vm.Intern("do"), Cdr(tmp))
		} else {
			expr = Cons(vm.Intern("do"), Cdr(tmp))
			tmp = List(vm.Intern("if"), Car(tmp), expr)
		}
		return vm.macroexpandObject(tmp)
	} else {
		return vm.nextCondClause(expr, Cdr(expr), i-1)
	}
}

func (vm *VM) expandQuasiquote(expr *Object) (*Object, error) {
	if ListLength(expr) != 2 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	return vm.expandQQ(Cadr(expr))
}

func (vm *VM) expandQQ(expr *Object) (*Object, error) {
	switch expr.Type {
	case ListType:
		if expr == EmptyList {
			return expr, nil
		}
		if expr.cdr != EmptyList {
			if expr.car == UnquoteSymbol {
				if expr.cdr.cdr != EmptyList {
					return nil, Error(SyntaxErrorKey, expr)
				}
				return vm.macroexpandObject(expr.cdr.car)
			} else if expr.car == UnquoteSplicingSymbol {
				return nil, Error(MacroErrorKey, "unquote-splicing can only occur in the context of a list ")
			}
		}
		tmp, err := vm.expandQQList(expr)
		if err != nil {
			return nil, err
		}
		return vm.macroexpandObject(tmp)
	case SymbolType:
		return List(vm.Intern("quote"), expr), nil
	default: //all other objects evaluate to themselves
		return expr, nil
	}
}

func (vm *VM) expandQQList(lst *Object) (*Object, error) {
	var tmp *Object
	var err error
	result := List(vm.Intern("concat"))
	tail := result
	for lst != EmptyList {
		item := Car(lst)
		if IsList(item) && item != EmptyList {
			if Car(item) == QuasiquoteSymbol {
				return nil, Error(MacroErrorKey, "nested quasiquote not supported")
			}
			if Car(item) == UnquoteSymbol && ListLength(item) == 2 {
				tmp, err = vm.macroexpandObject(Cadr(item))
				tmp = List(vm.Intern("list"), tmp)
				if err != nil {
					return nil, err
				}
				tail.cdr = List(tmp)
				tail = tail.cdr
			} else if Car(item) == UnquoteSplicingSymbol && ListLength(item) == 2 {
				tmp, err = vm.macroexpandObject(Cadr(item))
				if err != nil {
					return nil, err
				}
				tail.cdr = List(tmp)
				tail = tail.cdr
			} else {
				tmp, err = vm.expandQQList(item)
				if err != nil {
					return nil, err
				}
				tail.cdr = List(List(vm.Intern("list"), tmp))
				tail = tail.cdr
			}
		} else {
			tail.cdr = List(List(vm.Intern("quote"), List(item)))
			tail = tail.cdr
		}
		lst = Cdr(lst)
	}
	return result, nil
}
