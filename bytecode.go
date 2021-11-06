package vesper

import (
	"fmt"
	"strconv"
	"strings"
)

// Code - compiled Vesper bytecode
type Code struct {
	name     string
	ops      []int
	argc     int
	defaults []*Object
	keys     []*Object
	vm       *VM
}

// IsCode returns true if the object is a code object
func IsCode(obj *Object) bool {
	return obj.Type == CodeType
}

// MakeCode creates a new code object
func MakeCode(vm *VM, argc int, defaults []*Object, keys []*Object, name string) *Object {
	return &Object{
		Type: CodeType,
		code: &Code{
			name:     name,
			ops:      nil,
			argc:     argc,
			defaults: defaults, //nil for normal procs, empty for rest, and non-empty for optional/keyword
			keys:     keys,
			vm:       vm,
		},
	}
}

func (code *Code) signature() string {
	tmp := ""
	for i := 0; i < code.argc; i++ {
		tmp += " <any>"
	}
	if code.defaults != nil {
		tmp += " <any>*"
	}
	if tmp != "" {
		return "(" + tmp[1:] + ")"
	}
	return "()"
}

func (code *Code) decompile(vm *VM, pretty bool) string {
	var buf strings.Builder
	code.decompileInto(&buf, vm, "", pretty)
	s := buf.String()
	return strings.Replace(s, "("+FuncSymbol.text+" (\"\" 0 [] [])", "(code", 1)
}

func (code *Code) decompileInto(buf *strings.Builder, vm *VM, indent string, pretty bool) {
	indentAmount := "   "
	offset := 0
	max := len(code.ops)
	prefix := " "
	buf.WriteString(indent + "(" + FuncSymbol.text + " (")
	buf.WriteString(fmt.Sprintf("%q ", code.name))
	buf.WriteString(strconv.Itoa(code.argc))
	if code.defaults != nil {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("%v", code.defaults))
	} else {
		buf.WriteString(" []")
	}
	if code.keys != nil {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("%v", code.keys))
	} else {
		buf.WriteString(" []")
	}
	buf.WriteString(")")
	if pretty {
		indent = indent + indentAmount
		prefix = "\n" + indent
	}
	for offset < max {
		op := code.ops[offset]
		s := prefix + "(" + opsyms[op].text
		switch op {
		case opPop, opReturn, opNone:
			buf.WriteString(s + ")")
			offset++
		case opLiteral, opDefGlobal, opUse, opGlobal, opUndefGlobal, opDefMacro:
			buf.WriteString(s + " " + Write(vm.Constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opCall, opTailCall, opJumpFalse, opJump, opArray, opStruct:
			buf.WriteString(s + " " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opLocal, opSetLocal:
			buf.WriteString(s + " " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case opClosure:
			buf.WriteString(s)
			if pretty {
				buf.WriteString("\n")
			} else {
				buf.WriteString(" ")
			}
			indent2 := ""
			if pretty {
				indent2 = indent + indentAmount
			}
			vm.Constants[code.ops[offset+1]].code.decompileInto(buf, vm, indent2, pretty)
			buf.WriteString(")")
			offset += 2
		default:
			panic(fmt.Sprintf("Bad instruction: %d", code.ops[offset]))
		}
	}
	buf.WriteString(")")
}

func (code *Code) String(vm *VM) string {
	return code.decompile(vm, true)
}

func (code *Code) loadOps(vm *VM, lst *Object) error {
	for lst != EmptyList {
		instr := Car(lst)
		op := Car(instr)
		switch op {
		case ClosureSymbol:
			lstFunc := Cadr(instr)
			if Car(lstFunc) != FuncSymbol {
				return Error(SyntaxErrorKey, instr)
			}
			lstFunc = Cdr(lstFunc)
			funcParams := Car(lstFunc)
			var argc int
			var name string
			var defaults []*Object
			var keys []*Object
			var err error
			if IsSymbol(funcParams) {
				//legacy form, just the argc
				argc, err = AsIntValue(funcParams)
				if err != nil {
					return err
				}
				if argc < 0 {
					argc = -argc - 1
					defaults = make([]*Object, 0)
				}
			} else if IsList(funcParams) && ListLength(funcParams) == 4 {
				tmp := funcParams
				a := Car(tmp)
				tmp = Cdr(tmp)
				name, err = AsStringValue(a)
				if err != nil {
					return Error(SyntaxErrorKey, funcParams)
				}
				a = Car(tmp)
				tmp = Cdr(tmp)
				argc, err = AsIntValue(a)
				if err != nil {
					return Error(SyntaxErrorKey, funcParams)
				}
				a = Car(tmp)
				tmp = Cdr(tmp)
				if IsArray(a) {
					defaults = a.elements
				}
				a = Car(tmp)
				if IsArray(a) {
					keys = a.elements
				}
			} else {
				return Error(SyntaxErrorKey, funcParams)
			}
			fun := MakeCode(vm, argc, defaults, keys, name)
			_ = fun.code.loadOps(vm, Cdr(lstFunc))
			code.emitClosure(vm.putConstant(fun))
		case LiteralSymbol:
			code.emitLiteral(vm.putConstant(Cadr(instr)))
		case LocalSymbol:
			i, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			j, err := AsIntValue(Caddr(instr))
			if err != nil {
				return err
			}
			code.emitLocal(i, j)
		case SetlocalSymbol:
			i, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			j, err := AsIntValue(Caddr(instr))
			if err != nil {
				return err
			}
			code.emitSetLocal(i, j)
		case GlobalSymbol:
			sym := Cadr(instr)
			if IsSymbol(sym) {
				code.emitGlobal(vm.putConstant(sym))
			} else {
				return Error(GlobalSymbol, " argument 1 not a symbol: ", sym)
			}
		case UndefineSymbol:
			code.emitUndefGlobal(vm.putConstant(Cadr(instr)))
		case JumpSymbol:
			loc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitJump(loc)
		case JumpfalseSymbol:
			loc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitJumpFalse(loc)
		case CallSymbol:
			argc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitCall(argc)
		case TailcallSymbol:
			argc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitTailCall(argc)
		case ReturnSymbol:
			code.emitReturn()
		case PopSymbol:
			code.emitPop()
		case DefglobalSymbol:
			code.emitDefGlobal(vm.putConstant(Cadr(instr)))
		case DefmacroSymbol:
			code.emitDefMacro(vm.putConstant(Cadr(instr)))
		case UseSymbol:
			code.emitUse(vm.putConstant(Cadr(instr)))
		default:
			panic(fmt.Sprintf("Bad instruction: %v", op))
		}
		lst = Cdr(lst)
	}
	return nil
}

func (code *Code) emitLiteral(symIdx int) {
	code.ops = append(code.ops, opLiteral, symIdx)
}

func (code *Code) emitGlobal(symIdx int) {
	code.ops = append(code.ops, opGlobal, symIdx)
}
func (code *Code) emitCall(argc int) {
	code.ops = append(code.ops, opCall, argc)
}
func (code *Code) emitReturn() {
	code.ops = append(code.ops, opReturn)
}
func (code *Code) emitTailCall(argc int) {
	code.ops = append(code.ops, opTailCall, argc)
}
func (code *Code) emitPop() {
	code.ops = append(code.ops, opPop)
}
func (code *Code) emitLocal(i int, j int) {
	code.ops = append(code.ops, opLocal, i, j)
}
func (code *Code) emitSetLocal(i int, j int) {
	code.ops = append(code.ops, opSetLocal, i, j)
}
func (code *Code) emitDefGlobal(symIdx int) {
	code.ops = append(code.ops, opDefGlobal, symIdx)
}
func (code *Code) emitUndefGlobal(symIdx int) {
	code.ops = append(code.ops, opUndefGlobal, symIdx)
}
func (code *Code) emitDefMacro(symIdx int) {
	code.ops = append(code.ops, opDefMacro, symIdx)
}
func (code *Code) emitClosure(symIdx int) {
	code.ops = append(code.ops, opClosure, symIdx)
}
func (code *Code) emitJumpFalse(offset int) int {
	code.ops = append(code.ops, opJumpFalse, offset)
	return len(code.ops) - 1
}
func (code *Code) emitJump(offset int) int {
	code.ops = append(code.ops, opJump, offset)
	return len(code.ops) - 1
}
func (code *Code) setJumpLocation(loc int) {
	code.ops[loc] = len(code.ops) - loc + 1
}
func (code *Code) emitArray(alen int) {
	code.ops = append(code.ops, opArray, alen)
}
func (code *Code) emitStruct(slen int) {
	code.ops = append(code.ops, opStruct, slen)
}
func (code *Code) emitUse(symIdx int) {
	code.ops = append(code.ops, opUse, symIdx)
}
