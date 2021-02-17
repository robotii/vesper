package vesper

const (
	opNone = iota
	opLiteral
	opLocal
	opJumpFalse
	opJump
	opTailCall
	opCall
	opReturn
	opClosure
	opPop
	opGlobal
	opDefGlobal
	opSetLocal
	opUse
	opDefMacro
	opArray
	opStruct
	opUndefGlobal
	opCount
)

// NoneSymbol represents a nil operation
var NoneSymbol = Intern("none")

// LiteralSymbol represents a bytecode for a literal
var LiteralSymbol = Intern("literal")

// LocalSymbol represents bytecode for a local variable
var LocalSymbol = Intern("local")

// JumpfalseSymbol represents a jump on false
var JumpfalseSymbol = Intern("jumpfalse")

// JumpSymbol represents an unconditional jump
var JumpSymbol = Intern("jump")

// TailcallSymbol represents a tailcall operation
var TailcallSymbol = Intern("tailcall")

// CallSymbol represents a function call
var CallSymbol = Intern("call")

// ReturnSymbol represents a return from the current function
var ReturnSymbol = Intern("return")

// ClosureSymbol represents the creation of a closure
var ClosureSymbol = Intern("closure")

// PopSymbol represents the pop operation
var PopSymbol = Intern("pop")

// GlobalSymbol represents a reference to a global
var GlobalSymbol = Intern("global")

// DefglobalSymbol represents the definition of a global symbol
var DefglobalSymbol = Intern("defglobal")

// SetlocalSymbol represents setting a local varioble
var SetlocalSymbol = Intern("setlocal")

// UseSymbol represents the "use" of a module
var UseSymbol = Intern("use")

// DefmacroSymbol represents a macro definition
var DefmacroSymbol = Intern("defmacro")

// ArraySymbol represents the creation of a literal array
var ArraySymbol = Intern("array")

// StructSymbol represents creation of a literal struct
var StructSymbol = Intern("struct")

// UndefineSymbol represents an undefine operation
var UndefineSymbol = Intern("undefine")

// FuncSymbol represents a function definition
var FuncSymbol = Intern("func")

var opsyms = initOpsyms()

func initOpsyms() []*Object {
	syms := []*Object{
		opNone:        NoneSymbol,
		opLiteral:     LiteralSymbol,
		opLocal:       LocalSymbol,
		opJumpFalse:   JumpfalseSymbol,
		opJump:        JumpSymbol,
		opTailCall:    TailcallSymbol,
		opCall:        CallSymbol,
		opReturn:      ReturnSymbol,
		opClosure:     ClosureSymbol,
		opPop:         PopSymbol,
		opGlobal:      GlobalSymbol,
		opDefGlobal:   DefglobalSymbol,
		opSetLocal:    SetlocalSymbol,
		opUse:         UseSymbol,
		opDefMacro:    DefmacroSymbol,
		opArray:       ArraySymbol,
		opStruct:      StructSymbol,
		opUndefGlobal: UndefineSymbol,
	}
	return syms
}
