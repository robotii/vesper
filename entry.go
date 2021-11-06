package vesper

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Extension allows for extending Vesper
type Extension interface {
	Name() string
	Init(*VM) error
}

// Version - this version of vesper
const Version = "vesper v0.1"

var defaultVM = &VM{
	StackSize:    defaultStackSize,
	Symbols:      defaultSymtab,
	MacroMap:     copyMacros(nil),
	ConstantsMap: copyConstantMap(nil),
	Constants:    copyConstants(nil),
}

var verbose bool
var debug bool
var interactive bool

var loadPathSymbol = defaultVM.Intern("*load-path*")

// SetFlags - set various flags controlling the vm
func SetFlags(v bool, d bool, i bool) {
	verbose = v
	debug = d
	interactive = i
}

// DefineGlobal binds the value to the global name
func (vm *VM) DefineGlobal(name string, obj *Object) {
	sym := vm.Intern(name)
	if sym == nil {
		panic("Cannot define a value for this symbol: " + name)
	}
	vm.defGlobal(sym, obj)
}

func (vm *VM) definePrimitive(name string, prim *Object) {
	sym := vm.Intern(name)
	vm.defGlobal(sym, prim)
}

// DefineFunction registers a primitive function to the specified global name
func (vm *VM) DefineFunction(name string, fun PrimitiveFunction, result *Object, args ...*Object) {
	prim := Primitive(name, fun, result, args, nil, nil, nil)
	vm.definePrimitive(name, prim)
}

// DefineFunctionRestArgs registers a primitive function with Rest arguments to the specified global name
func (vm *VM) DefineFunctionRestArgs(name string, fun PrimitiveFunction, result *Object, rest *Object, args ...*Object) {
	prim := Primitive(name, fun, result, args, rest, []*Object{}, nil)
	vm.definePrimitive(name, prim)
}

// DefineFunctionOptionalArgs registers a primitive function with optional arguments to the specified global name
func (vm *VM) DefineFunctionOptionalArgs(name string, fun PrimitiveFunction, result *Object, args []*Object, defaults ...*Object) {
	prim := Primitive(name, fun, result, args, nil, defaults, nil)
	vm.definePrimitive(name, prim)
}

// DefineFunctionKeyArgs registers a primitive function with keyword arguments to the specified global name
func (vm *VM) DefineFunctionKeyArgs(name string, fun PrimitiveFunction, result *Object, args []*Object, defaults []*Object, keys []*Object) {
	prim := Primitive(name, fun, result, args, nil, defaults, keys)
	vm.definePrimitive(name, prim)
}

// DefineMacro registers a primitive macro with the specified name.
func (vm *VM) DefineMacro(name string, fun PrimitiveFunction) {
	sym := vm.Intern(name)
	prim := Primitive(name, fun, AnyType, []*Object{AnyType}, nil, nil, nil)
	vm.defMacro(sym, prim)
}

// GetKeywords - return a slice of Vesper primitive reserved words
func (vm *VM) GetKeywords() []*Object {
	//keywords reserved for the base language that Vesper compiles
	keywords := []*Object{
		vm.Intern("quote"),
		vm.Intern("fn"),
		vm.Intern("if"),
		vm.Intern("do"),
		vm.Intern("def"),
		vm.Intern("defn"),
		vm.Intern("defmacro"),
		vm.Intern("set!"),
		vm.Intern("code"),
		vm.Intern("use"),
	}
	return keywords
}

// Globals - return a slice of all defined global symbols
func (vm *VM) Globals() []*Object {
	var syms []*Object
	for _, sym := range vm.Symbols {
		if sym.car != nil {
			syms = append(syms, sym)
		}
	}
	return syms
}

// GetGlobal - return the global value for the specified symbol, or nil if the symbol is not defined.
func GetGlobal(sym *Object) *Object {
	if IsSymbol(sym) {
		return sym.car
	}
	return nil
}

func (vm *VM) defGlobal(sym *Object, val *Object) {
	sym.car = val
	delete(vm.MacroMap, sym)
}

// IsDefined - return true if the there is a global value defined for the symbol
func IsDefined(sym *Object) bool {
	return sym.car != nil
}

func undefGlobal(sym *Object) {
	sym.car = nil
}

// Macros - return a slice of all defined macros
func (vm *VM) Macros() []*Object {
	keys := make([]*Object, 0, len(vm.MacroMap))
	for k := range vm.MacroMap {
		keys = append(keys, k)
	}
	return keys
}

// GetMacro - return the macro for the symbol, or nil if not defined
func (vm *VM) GetMacro(sym *Object) *Macro {
	mac, ok := vm.MacroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func (vm *VM) defMacro(sym *Object, val *Object) {
	vm.MacroMap[sym] = NewMacro(sym, val)
}

func (vm *VM) putConstant(val *Object) int {
	idx, present := vm.ConstantsMap[val]
	if !present {
		idx = len(vm.Constants)
		vm.Constants = append(vm.Constants, val)
		vm.ConstantsMap[val] = idx
	}
	return idx
}

// Use is a synonym for load
func (vm *VM) Use(sym *Object) error {
	return vm.Load(sym.text)
}

func (vm *VM) importCode(thunk *Object) (*Object, error) {
	var args []*Object
	result, err := vm.Execute(thunk.code, args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindModuleByName returns the file filename of a vesper module
func FindModuleByName(moduleName string) (string, error) {
	loadPath := GetGlobal(loadPathSymbol)
	if loadPath == nil {
		loadPath = String(".")
	}
	path := strings.Split(StringValue(loadPath), ":")
	name := moduleName
	var lname string
	if strings.HasSuffix(name, ".vesp") {
		lname = name[:len(name)-3] + ".vem"
	} else {
		lname = name + ".vem"
		name = name + ".vsp"
	}
	for _, dirname := range path {
		filename := filepath.Join(dirname, lname)
		if IsFileReadable(filename) {
			return filename, nil
		}
		filename = filepath.Join(dirname, name)
		if IsFileReadable(filename) {
			return filename, nil
		}
	}
	return "", Error(IOErrorKey, "Module not found: ", moduleName)
}

// Load checks for a loadable module and loads it, if it exists
func (vm *VM) Load(name string) error {
	file, err := FindModuleFile(name)
	if err != nil {
		return err
	}
	return vm.LoadFile(file)
}

// LoadFile loads and executes a file returning any error
func (vm *VM) LoadFile(file string) error {
	if verbose {
		println("; loadFile: " + file)
	} else if interactive {
		println("[loading " + file + "]")
	}
	fileText, err := SlurpFile(file)
	if err != nil {
		return err
	}
	exprs, err := vm.ReadAll(fileText, nil)
	if err != nil {
		return err
	}
	for exprs != EmptyList {
		expr := Car(exprs)
		_, err = vm.Eval(expr)
		if err != nil {
			return err
		}
		exprs = Cdr(exprs)
	}
	return nil
}

// Eval evaluates an expression
func (vm *VM) Eval(expr *Object) (*Object, error) {
	if debug {
		println("; eval: ", Write(expr))
	}
	expanded, err := vm.macroexpandObject(expr)
	if err != nil {
		return nil, err
	}
	if debug {
		println("; expanded to: ", Write(expanded))
	}
	code, err := vm.Compile(expanded)
	if err != nil {
		return nil, err
	}
	if debug {
		val := strings.Replace(Write(code), "\n", "\n; ", -1)
		println("; compiled to:\n;  ", val)
	}
	return vm.importCode(code)
}

// FindModuleFile finds a readable module file or errors
func FindModuleFile(name string) (string, error) {
	i := strings.Index(name, ".")
	if i < 0 {
		file, err := FindModuleByName(name)
		if err != nil {
			return "", err
		}
		return file, nil
	}
	if !IsFileReadable(name) {
		return "", Error(IOErrorKey, "Cannot read file: ", name)
	}
	return name, nil
}

func (vm *VM) compileObject(expr *Object) (string, error) {
	if debug {
		println("; compile: ", Write(expr))
	}
	expanded, err := vm.macroexpandObject(expr)
	if err != nil {
		return "", err
	}
	if debug {
		println("; expanded to: ", Write(expanded))
	}
	thunk, err := vm.Compile(expanded)
	if err != nil {
		return "", err
	}
	if debug {
		println("; compiled to: ", Write(thunk))
	}
	return thunk.code.decompile(vm, true) + "\n", nil
}

// CompileFile compiles a file and returns a String object or an error
// caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func (vm *VM) CompileFile(name string) (*Object, error) {
	file, err := FindModuleFile(name)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; loadFile: " + file)
	}
	fileText, err := SlurpFile(file)
	if err != nil {
		return nil, err
	}

	exprs, err := vm.ReadAll(fileText, nil)
	if err != nil {
		return nil, err
	}
	result := ";\n; code generated from " + file + "\n;\n"
	var lvm string
	for exprs != EmptyList {
		expr := Car(exprs)
		lvm, err = vm.compileObject(expr)
		if err != nil {
			return nil, err
		}
		result += lvm
		exprs = Cdr(exprs)
	}
	return String(result), nil
}

// AddVesperDirectory adds a directory to the load path
func (vm *VM) AddVesperDirectory(dirname string) {
	loadPath := dirname
	tmp := GetGlobal(loadPathSymbol)
	if tmp != nil {
		loadPath = dirname + ":" + StringValue(tmp)
	}
	vm.DefineGlobal(StringValue(loadPathSymbol), String(loadPath))
}

// Init initialise the base environment and extensions
func (vm *VM) Init(extns ...Extension) *VM {
	vm.Extensions = extns
	loadPath := os.Getenv("VESPER_PATH")
	home := os.Getenv("HOME")
	if loadPath == "" {
		loadPath = ".:./lib"
		homelib := filepath.Join(home, "lib/vesper")
		_, err := os.Stat(homelib)
		if err == nil {
			loadPath += ":" + homelib
		}
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			golibdir := filepath.Join(gopath, "src/github.com/robotii/vesper/lib")
			_, err := os.Stat(golibdir)
			if err == nil {
				loadPath += ":" + golibdir
			}
		}
	}
	vm.DefineGlobal(StringValue(loadPathSymbol), String(loadPath))
	vm.InitPrimitives()
	for _, ext := range vm.Extensions {
		err := ext.Init(vm)
		if err != nil {
			Fatal("*** ", err)
		}
	}
	return vm
}

// Run the given files in the vesper vm
func (vm *VM) Run(args ...string) {
	for _, filename := range args {
		err := vm.Load(filename)
		if err != nil {
			Fatal("*** ", err.Error())
		}
	}
	// TODO: implement
}

// Main entrypoint for main Vesper interpreter
func Main(extns ...Extension) {
	var help, version, compile, verbose, debug bool
	var path string
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&version, "version", false, "shows the current version")
	flag.BoolVar(&compile, "compile", false, "compile the file and output code")
	flag.BoolVar(&verbose, "verbose", false, "verbose mode, print extra information")
	flag.BoolVar(&debug, "debug", false, "debug mode, print extra information about compilation")
	flag.StringVar(&path, "path", "", "add directories to vesper load path")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(1)
	}
	if version {
		fmt.Println(Version)
		os.Exit(0)
	}
	args := flag.Args()
	interactive := len(args) == 0
	defaultVM.Init(extns...)

	// We create and initialise a new VM here,
	// so that we have a clean environment for the interactive mode
	vm := NewVM().Init(extns...)
	if path != "" {
		for _, p := range strings.Split(path, ":") {
			expandedPath := ExpandFilePath(p)
			if IsDirectoryReadable(expandedPath) {
				vm.AddVesperDirectory(expandedPath)
				if debug {
					Println("[added directory to path: '", expandedPath, "']")
				}
			} else if debug {
				Println("[directory not readable, cannot add to path: '", expandedPath, "']")
			}
		}
	}
	if !interactive {
		if compile {
			for _, filename := range args {
				generated, err := vm.CompileFile(filename)
				if err != nil {
					Fatal("*** ", err)
				}
				Println(generated)
			}
		} else {
			SetFlags(verbose, debug, interactive)
			vm.Run(args...)
		}
	} else {
		SetFlags(verbose, debug, interactive)
		REPL()
	}
}
