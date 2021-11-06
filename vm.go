package vesper

import (
	"fmt"
	"os"
	"time"
)

// VM - the Vesper VM
type VM struct {
	StackSize    int
	Symbols      map[string]*Object
	MacroMap     map[*Object]*Macro
	ConstantsMap map[*Object]int
	Constants    []*Object
	Extensions   []Extension
}

const defaultStackSize = 1000

var interrupted = false
var interrupts chan os.Signal

func checkInterrupt() bool {
	if interrupts != nil {
		select {
		case msg := <-interrupts:
			if msg != nil {
				interrupted = true
				return true
			}
		default:
			return false
		}
	}
	return false
}

// NewVM creates a new VM
func NewVM() *VM {
	return CloneVM(defaultVM)
}

// CloneVM creates a clone of the original VM
func CloneVM(copy *VM) *VM {
	return &VM{
		StackSize:    copy.StackSize,
		Symbols:      copyEnv(copy.Symbols),
		MacroMap:     copyMacros(copy.MacroMap),
		ConstantsMap: copyConstantMap(copy.ConstantsMap),
		Constants:    copyConstants(copy.Constants),
	}
}

func addContext(env *frame, err error) error {
	if e, ok := err.(*Object); ok {
		if env.code != nil {
			if env.code.name != "throw" {
				e.text = env.code.name
			} else if env.previous != nil {
				if env.previous.code != nil {
					e.text = env.previous.code.name
				}
			}
		}
	}
	return err
}

func (vm *VM) keywordCall(fun *Object, argc int, pc int, stack []*Object, sp int) (int, int, error) {
	if argc != 1 {
		return 0, 0, Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
	}
	v, err := Get(stack[sp], fun)
	if err != nil {
		return 0, 0, err
	}
	stack[sp] = v
	return pc, sp, nil
}

func argcError(name string, min int, max int, provided int) error {
	s := "1 argument"
	if min == max {
		if min != 1 {
			s = fmt.Sprintf("%d arguments", min)
		}
	} else if max < 0 {
		s = fmt.Sprintf("%d or more arguments", min)
	} else {
		s = fmt.Sprintf("%d to %d arguments", min, max)
	}
	return Error(ArgumentErrorKey, fmt.Sprintf("%s expected %s, got %d", name, s, provided))
}

func (vm *VM) callPrimitive(prim *primitive, argv []*Object) (*Object, error) {
	if prim.defaults != nil {
		return vm.callPrimitiveWithDefaults(prim, argv)
	}
	argc := len(argv)
	if argc != prim.argc {
		return nil, argcError(prim.name, prim.argc, prim.argc, argc)
	}
	for i, arg := range argv {
		t := prim.args[i]
		if t != AnyType && arg.Type != t {
			return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].text, i+1, argv[i].Type.text))
		}
	}
	return prim.fun(argv)
}

func (vm *VM) callPrimitiveWithDefaults(prim *primitive, argv []*Object) (*Object, error) {
	provided := len(argv)
	minargc := prim.argc
	if len(prim.defaults) == 0 {
		rest := prim.rest
		if provided < minargc {
			return nil, argcError(prim.name, minargc, -1, provided)
		}
		for i := 0; i < minargc; i++ {
			t := prim.args[i]
			arg := argv[i]
			if t != AnyType && arg.Type != t {
				return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].text, i+1, argv[i].Type.text))
			}
		}
		if rest != AnyType {
			for i := minargc; i < provided; i++ {
				arg := argv[i]
				if arg.Type != rest {
					return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, rest.text, i+1, argv[i].Type.text))
				}
			}
		}
		return prim.fun(argv)
	}
	maxargc := len(prim.args)
	if provided < minargc {
		return nil, argcError(prim.name, minargc, maxargc, provided)
	}
	newargs := make([]*Object, maxargc)
	if prim.keys != nil {
		j := 0
		copy(newargs, argv[:minargc])
		for i := minargc; i < maxargc; i++ {
			newargs[i] = prim.defaults[j]
			j++
		}
		j = minargc
		ndefaults := len(prim.defaults)
		for j < provided {
			k := argv[j]
			j++
			if j == provided {
				return nil, Error(ArgumentErrorKey, "mismatched keyword/value pair in argument list")
			}
			if k.Type != KeywordType {
				return nil, Error(ArgumentErrorKey, "expected keyword, got a "+k.Type.text)
			}
			gotit := false
			for i := 0; i < ndefaults; i++ {
				if prim.keys[i] == k {
					gotit = true
					newargs[i+minargc] = argv[j]
					j++
					break
				}
			}
			if !gotit {
				return nil, Error(ArgumentErrorKey, prim.name, " accepts ", prim.keys, " as keyword arg(s), not ", k)
			}
		}
		argv = newargs
	} else {
		if provided > maxargc {
			return nil, argcError(prim.name, minargc, maxargc, provided)
		}
		copy(newargs, argv)
		j := 0
		for i := provided; i < maxargc; i++ {
			newargs[i] = prim.defaults[j]
			j++
		}
		argv = newargs
	}
	for i, arg := range argv {
		t := prim.args[i]
		if t != AnyType && arg.Type != t {
			return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].text, i+1, argv[i].Type.text))
		}
	}
	return prim.fun(argv)
}

func (vm *VM) funcall(fun *Object, argc int, ops []int, savedPc int, stack []*Object, sp int, env *frame) ([]int, int, int, *frame, error) {
opCallAgain:
	if fun.Type == FunctionType {
		if fun.code != nil {
			if interrupted || checkInterrupt() {
				return nil, 0, 0, nil, addContext(env, Error(InterruptKey))
			}
			if fun.code.defaults == nil {
				f := &frame{
					previous: env,
					pc:       savedPc,
					ops:      ops,
					locals:   fun.frame,
					code:     fun.code,
				}
				expectedArgc := fun.code.argc
				if argc != expectedArgc {
					return nil, 0, 0, nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
				}
				if argc <= 5 {
					f.elements = f.firstfive[:argc]
				} else {
					f.elements = make([]*Object, argc)
				}
				endSp := sp + argc
				copy(f.elements, stack[sp:endSp])
				return fun.code.ops, 0, endSp, f, nil
			}
			f, err := vm.buildFrame(env, savedPc, ops, fun, argc, stack, sp)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp += argc
			env = f
			ops = fun.code.ops
			return ops, 0, sp, env, err
		}
		if fun.primitive != nil {
			val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = val
			return ops, savedPc, sp, env, err
		}
		if fun == Apply {
			if argc < 2 {
				err := Error(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
				return vm.catch(err, stack, env)
			}
			fun = stack[sp]
			args := stack[sp+argc-1]
			if !IsList(args) {
				err := Error(ArgumentErrorKey, "apply expected a <list> as its final argument")
				return vm.catch(err, stack, env)
			}
			arglist := args
			for i := argc - 2; i > 0; i-- {
				arglist = Cons(stack[sp+i], arglist)
			}
			sp += argc
			argc = ListLength(arglist)
			i := 0
			sp -= argc
			for arglist != EmptyList {
				stack[sp+i] = arglist.car
				i++
				arglist = arglist.cdr
			}
			goto opCallAgain
		}
		if fun == CallCC {
			if argc != 1 {
				err := Error(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
			}
			fun = stack[sp]
			stack[sp] = Continuation(env, ops, savedPc, stack[sp+1:])
			goto opCallAgain
		}
		if fun.continuation != nil {
			if argc != 1 {
				err := Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
			}
			arg := stack[sp]
			sp = len(stack) - len(fun.continuation.stack)
			segment := stack[sp:]
			copy(segment, fun.continuation.stack)
			sp--
			stack[sp] = arg
			return fun.continuation.ops, fun.continuation.pc, sp, fun.frame, nil
		}
		if fun == GoFunc {
			err := vm.spawn(stack[sp], argc-1, stack, sp+1)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = Null
			return ops, savedPc, sp, env, err
		}
		panic("unsupported instruction")
	}
	if fun.Type == KeywordType {
		if argc != 1 {
			err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
			return vm.catch(err, stack, env)
		}
		v, err := Get(stack[sp], fun)
		if err != nil {
			return vm.catch(err, stack, env)
		}
		stack[sp] = v
		return ops, savedPc, sp, env, err
	}
	err := Error(ArgumentErrorKey, "Not a function: ", fun)
	return vm.catch(err, stack, env)
}

func (vm *VM) tailcall(fun *Object, argc int, ops []int, stack []*Object, sp int, env *frame) ([]int, int, int, *frame, error) {
opTailCallAgain:
	if fun.Type == FunctionType {
		if fun.code != nil {
			if fun.code.defaults == nil && fun.code == env.code { //self-tail-call - we can reuse the frame.
				expectedArgc := fun.code.argc
				if argc != expectedArgc {
					return nil, 0, 0, nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
				}
				endSp := sp + argc
				copy(env.elements, stack[sp:endSp])
				return fun.code.ops, 0, endSp, env, nil
			}
			f, err := vm.buildFrame(env.previous, env.pc, env.ops, fun, argc, stack, sp)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp += argc
			return fun.code.ops, 0, sp, f, nil
		}
		if fun.primitive != nil {
			val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = val
			return env.ops, env.pc, sp, env.previous, nil
		}
		if fun == Apply {
			if argc < 2 {
				err := Error(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
				return vm.catch(err, stack, env)
			}
			fun = stack[sp]
			args := stack[sp+argc-1]
			if !IsList(args) {
				err := Error(ArgumentErrorKey, "apply expected its last argument to be a <list>")
				return vm.catch(err, stack, env)
			}
			arglist := args
			for i := argc - 2; i > 0; i-- {
				arglist = Cons(stack[sp+i], arglist)
			}
			sp += argc
			argc = ListLength(arglist)
			i := 0
			sp -= argc
			for arglist != EmptyList {
				stack[sp+i] = arglist.car
				i++
				arglist = arglist.cdr
			}
			goto opTailCallAgain
		}
		if fun.continuation != nil {
			if argc != 1 {
				err := Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
			}
			arg := stack[sp]
			sp = len(stack) - len(fun.continuation.stack)
			segment := stack[sp:]
			copy(segment, fun.continuation.stack)
			sp--
			stack[sp] = arg
			return fun.continuation.ops, fun.continuation.pc, sp, fun.frame, nil
		}
		if fun == CallCC {
			if argc != 1 {
				err := Error(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
			}
			fun = stack[sp]
			stack[sp] = Continuation(env.previous, env.ops, env.pc, stack[sp:])
			goto opTailCallAgain
		}
		if fun == GoFunc {
			err := vm.spawn(stack[sp], argc-1, stack, sp+1)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = Null
			return env.ops, env.pc, sp, env.previous, nil
		}
		panic("Bad function")
	}
	if fun.Type == KeywordType {
		if argc != 1 {
			err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
			return vm.catch(err, stack, env)
		}
		v, err := Get(stack[sp], fun)
		if err != nil {
			return vm.catch(err, stack, env)
		}
		stack[sp] = v
		return env.ops, env.pc, sp, env.previous, nil
	}
	err := Error(ArgumentErrorKey, "Not a function:", fun)
	return vm.catch(err, stack, env)
}

func (vm *VM) keywordTailcall(fun *Object, argc int, ops []int, stack []*Object, sp int, env *frame) ([]int, int, int, *frame, error) {
	if argc != 1 {
		err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
		return vm.catch(err, stack, env)
	}
	v, err := Get(stack[sp], fun)
	if err != nil {
		return vm.catch(err, stack, env)
	}
	stack[sp] = v
	return env.ops, env.pc, sp, env.previous, nil
}

func (vm *VM) execCompileTime(code *Code, arg *Object) (*Object, error) {
	args := []*Object{arg}
	prev := verbose
	verbose = false
	res, err := vm.Execute(code, args)
	verbose = prev
	return res, err
}

func (vm *VM) catch(err error, stack []*Object, env *frame) ([]int, int, int, *frame, error) {
	errobj, ok := err.(*Object)
	if !ok {
		errobj = MakeError(ErrorKey, String(err.Error()))
	}
	handler := GetGlobal(vm.Intern("*top-handler*"))
	if handler != nil && handler.Type == FunctionType {
		if handler.code != nil {
			if handler.code.argc == 1 {
				sp := len(stack) - 1
				stack[sp] = errobj
				return vm.funcall(handler, 1, nil, 0, stack, sp, nil)
			}
		}
	}
	return nil, 0, 0, nil, addContext(env, err)
}

func (vm *VM) spawn(fun *Object, argc int, stack []*Object, sp int) error {
	if fun.Type == FunctionType {
		if fun.code != nil {
			env, err := vm.buildFrame(nil, 0, nil, fun, argc, stack, sp)
			if err != nil {
				return err
			}
			go func(code *Code, env *frame) {
				_, err := vm.exec(code, env)
				if err != nil {
					// TODO: How do we handle errors?
					println("; [*** error in goroutine '", code.name, "': ", err, "]")
				} else if verbose {
					println("; [goroutine '", code.name, "' exited cleanly]")
				}
			}(fun.code, env)
			return nil
		}
		// apply, go and callcc cannot be called directly in a goroutine
		// TODO: Should we allow this?
	}
	return Error(ArgumentErrorKey, "Bad function for go: ", fun)
}

// Execute runs the given code with the given arguments
func (vm *VM) Execute(code *Code, args []*Object) (*Object, error) {
	if len(args) != code.argc {
		return nil, Error(ArgumentErrorKey, "Wrong number of arguments")
	}
	env := &frame{
		elements: make([]*Object, len(args)),
		code:     code,
	}
	copy(env.elements, args)
	startTime := time.Now()
	result, err := vm.exec(code, env)
	dur := time.Since(startTime)
	if err != nil {
		return nil, err
	}
	if result == nil {
		panic("result should never be nil if no error")
	}
	if verbose {
		println("; executed in ", dur.String())
		if !interactive {
			println("; => ", result.String())
		}
	}
	return result, err
}

func (vm *VM) exec(code *Code, env *frame) (*Object, error) {
	stack := make([]*Object, vm.StackSize)
	sp := vm.StackSize
	ops := code.ops
	pc := 0
	var err error
	for {
		op := ops[pc]
		switch op {
		case opCall:
			argc := ops[pc+1]
			fun := stack[sp]
			if fun.primitive != nil {
				nextSp := sp + argc
				val, err := vm.callPrimitive(fun.primitive, stack[sp+1:nextSp+1])
				if err != nil {
					ops, pc, _, env, err = vm.catch(err, stack, env)
					if err != nil {
						return nil, err
					}
				}
				stack[nextSp] = val
				sp = nextSp
				pc += 2
			} else if fun.Type == FunctionType {
				ops, pc, sp, env, err = vm.funcall(fun, argc, ops, pc+2, stack, sp+1, env)
				if err != nil {
					return nil, err
				}
			} else if fun.Type == KeywordType {
				pc, sp, err = vm.keywordCall(fun, argc, pc+2, stack, sp+1)
				if err != nil {
					ops, pc, sp, env, err = vm.catch(err, stack, env)
					if err != nil {
						return nil, err
					}
				}
			} else {
				ops, pc, sp, env, err = vm.catch(Error(ArgumentErrorKey, "Not callable: ", fun), stack, env)
				if err != nil {
					return nil, err
				}
			}

		case opGlobal:
			sym := vm.Constants[ops[pc+1]]
			sp--
			// Check for undefined globals
			if sym == nil || sym.car == nil {
				stack[sp] = Null
			} else {
				stack[sp] = sym.car
			}
			pc += 2

		case opLocal:
			tmpEnv := env
			i := ops[pc+1]
			for i > 0 {
				tmpEnv = tmpEnv.locals
				i--
			}
			j := ops[pc+2]
			val := tmpEnv.elements[j]
			sp--
			stack[sp] = val
			pc += 3

		case opJumpFalse:
			b := stack[sp]
			sp++
			if b == False {
				pc += ops[pc+1]
			} else {
				pc += 2
			}

		case opPop:
			sp++
			pc++

		case opTailCall:
			fun := stack[sp]
			argc := ops[pc+1]
			if fun.primitive != nil {
				nextSp := sp + argc
				val, err := vm.callPrimitive(fun.primitive, stack[sp+1:nextSp+1])
				if err != nil {
					_, _, _, env, err = vm.catch(err, stack, env)
					if err != nil {
						return nil, err
					}
				}
				stack[nextSp] = val
				sp = nextSp
				ops = env.ops
				pc = env.pc
				env = env.previous
				if env == nil {
					return stack[sp], nil
				}
			} else if fun.Type == FunctionType {
				ops, pc, sp, env, err = vm.tailcall(fun, argc, ops, stack, sp+1, env)
				if err != nil {
					return nil, err
				}
				if env == nil {
					return stack[sp], nil
				}
			} else if fun.Type == KeywordType {
				ops, pc, sp, env, err = vm.keywordTailcall(fun, argc, ops, stack, sp+1, env)
				if err != nil {
					ops, pc, sp, env, err = vm.catch(err, stack, env)
					if err != nil {
						return nil, err
					}
				} else {
					if env == nil {
						return stack[sp], nil
					}
				}
			} else {
				ops, pc, sp, env, err = vm.catch(Error(ArgumentErrorKey, "Not callable: ", fun), stack, env)
				if err != nil {
					return nil, err
				}
			}

		case opLiteral:
			sp--
			stack[sp] = vm.Constants[ops[pc+1]]
			pc += 2

		case opSetLocal:
			tmpEnv := env
			i := ops[pc+1]
			for i > 0 {
				tmpEnv = tmpEnv.locals
				i--
			}
			j := ops[pc+2]
			tmpEnv.elements[j] = stack[sp]
			pc += 3

		case opClosure:
			sp--
			stack[sp] = Closure(vm.Constants[ops[pc+1]].code, env)
			pc = pc + 2

		case opReturn:
			if env.previous == nil {
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			env = env.previous

		case opJump:
			pc += ops[pc+1]

		case opDefGlobal:
			sym := vm.Constants[ops[pc+1]]
			vm.defGlobal(sym, stack[sp])
			pc += 2

		case opUndefGlobal:
			sym := vm.Constants[ops[pc+1]]
			undefGlobal(sym)
			pc += 2

		case opDefMacro:
			sym := vm.Constants[ops[pc+1]]
			vm.defMacro(sym, stack[sp])
			stack[sp] = sym
			pc += 2

		case opUse:
			sym := vm.Constants[ops[pc+1]]
			err := vm.Use(sym)
			if err != nil {
				ops, pc, sp, env, err = vm.catch(err, stack, env)
				if err != nil {
					return nil, err
				}
			} else {
				sp--
				stack[sp] = sym
				pc += 2
			}

		case opArray:
			vlen := ops[pc+1]
			v := Array(stack[sp : sp+vlen]...)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2

		case opStruct:
			vlen := ops[pc+1]
			v, _ := Struct(stack[sp : sp+vlen])
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2

		case opNone, opCount:
			// Do nothing

		default:
			panic("Bad instruction")
		}
	}
}
