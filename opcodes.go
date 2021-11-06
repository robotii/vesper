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

var (
	// NoneSymbol represents a nil operation
	NoneSymbol = defaultVM.Intern("none")
	// LiteralSymbol represents a bytecode for a literal
	LiteralSymbol = defaultVM.Intern("literal")
	// LocalSymbol represents bytecode for a local variable
	LocalSymbol = defaultVM.Intern("local")
	// JumpfalseSymbol represents a jump on false
	JumpfalseSymbol = defaultVM.Intern("jumpfalse")
	// JumpSymbol represents an unconditional jump
	JumpSymbol = defaultVM.Intern("jump")
	// TailcallSymbol represents a tailcall operation
	TailcallSymbol = defaultVM.Intern("tailcall")
	// CallSymbol represents a function call
	CallSymbol = defaultVM.Intern("call")
	// ReturnSymbol represents a return from the current function
	ReturnSymbol = defaultVM.Intern("return")
	// ClosureSymbol represents the creation of a closure
	ClosureSymbol = defaultVM.Intern("closure")
	// PopSymbol represents the pop operation
	PopSymbol = defaultVM.Intern("pop")
	// GlobalSymbol represents a reference to a global
	GlobalSymbol = defaultVM.Intern("global")
	// DefglobalSymbol represents the definition of a global symbol
	DefglobalSymbol = defaultVM.Intern("defglobal")
	// SetlocalSymbol represents setting a local varioble
	SetlocalSymbol = defaultVM.Intern("setlocal")
	// UseSymbol represents the "use" of a module
	UseSymbol = defaultVM.Intern("use")
	// DefmacroSymbol represents a macro definition
	DefmacroSymbol = defaultVM.Intern("defmacro")
	// ArraySymbol represents the creation of a literal array
	ArraySymbol = defaultVM.Intern("array")
	// StructSymbol represents creation of a literal struct
	StructSymbol = defaultVM.Intern("struct")
	// UndefineSymbol represents an undefine operation
	UndefineSymbol = defaultVM.Intern("undefine")
	// FuncSymbol represents a function definition
	FuncSymbol = defaultVM.Intern("func")
)

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
