package vesper


// ErrorKey - used for generic errors
var ErrorKey = Intern("error:")

// ArgumentErrorKey used for argument errors
var ArgumentErrorKey = Intern("argument-error:")

// SyntaxErrorKey used for syntax errors
var SyntaxErrorKey = Intern("syntax-error:")

// MacroErrorKey used for macro errors
var MacroErrorKey = Intern("macro-error:")

// IOErrorKey used for IO errors
var IOErrorKey = Intern("io-error:")

// InterruptKey used for interrupts that were captured
var InterruptKey = Intern("interrupt:")

// Error creates a new Error from the arguments. The first is an actual Vesper keyword object,
