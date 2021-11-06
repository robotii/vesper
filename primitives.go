package vesper

// the primitive functions for the languages
import (
	"fmt"
	"math"
	"os"
	"time"
)

// PrimitiveFunction is the native go function signature for all Vesper primitive functions
type PrimitiveFunction func(argv []*Object) (*Object, error)

// Primitive - a primitive function, written in Go, callable by VM
type primitive struct { // <function>
	name      string
	fun       PrimitiveFunction
	signature string
	argc      int       // -1 means the primitive itself checks the args (legacy mode)
	args      []*Object // if set, the length must be for total args (both required and optional). The type (or <any>) for each
	rest      *Object   // if set, then any number of this type can follow the normal args. Mutually incompatible with defaults/keys
	defaults  []*Object // if set, then that many optional args beyond argc have these default values
	keys      []*Object // if set, then it must match the size of defaults, and these are the keys
}

// Primitive creates a primitive function
func Primitive(name string, fun PrimitiveFunction, result *Object, args []*Object, rest *Object, defaults []*Object, keys []*Object) *Object {
	argc := len(args)
	if defaults != nil {
		defc := len(defaults)
		if defc > argc {
			panic("more default argument values than types: " + name)
		}
		if keys != nil {
			if len(keys) != defc {
				panic("Argument keys must have same length as argument defaults")
			}
		}
		argc = argc - defc
		for i := 0; i < defc; i++ {
			t := args[argc+i]
			if t != AnyType && defaults[i].Type != t {
				panic("argument default's type (" + defaults[i].Type.text + ") doesn't match declared type (" + t.text + ")")
			}
		}
	} else {
		if keys != nil {
			panic("Cannot have argument keys without argument defaults")
		}
	}
	signature := functionSignatureFromTypes(result, args, rest)
	prim := &primitive{name, fun, signature, argc, args, rest, defaults, keys}
	return &Object{Type: FunctionType, primitive: prim}
}

// InitPrimitives defines the global functions/variables/macros for the top level environment
func (vm *VM) InitPrimitives() {
	vm.DefineMacro("let", vm.vesperLet)
	vm.DefineMacro("letrec", vm.vesperLetrec)
	vm.DefineMacro("cond", vm.vesperCond)
	vm.DefineMacro("quasiquote", vm.vesperQuasiquote)

	vm.DefineGlobal("null", Null)
	vm.DefineGlobal("true", True)
	vm.DefineGlobal("false", False)

	vm.DefineGlobal("apply", Apply)
	vm.DefineGlobal("callcc", CallCC)
	vm.DefineGlobal("go", GoFunc)

	vm.DefineFunction("globals", vm.vesperGlobals, ArrayType)
	vm.DefineFunction("version", vm.vesperVersion, StringType)
	vm.DefineFunction("boolean?", vesperBooleanP, BooleanType, AnyType)
	vm.DefineFunction("not", vesperNot, BooleanType, AnyType)
	vm.DefineFunction("equal?", vesperEqualP, BooleanType, AnyType, AnyType)
	vm.DefineFunction("identical?", vesperIdenticalP, BooleanType, AnyType, AnyType)
	vm.DefineFunction("null?", vesperNullP, BooleanType, AnyType)
	vm.DefineFunction("def?", vesperDefinedP, BooleanType, SymbolType)

	vm.DefineFunction("type", vesperType, TypeType, AnyType)
	vm.DefineFunction("value", vesperValue, AnyType, AnyType)
	vm.DefineFunction("instance", vesperInstance, AnyType, TypeType, AnyType)

	vm.DefineFunction("type?", vesperTypeP, BooleanType, AnyType)
	vm.DefineFunction("type-name", vm.vesperTypeName, SymbolType, TypeType)
	vm.DefineFunction("keyword?", vesperKeywordP, BooleanType, AnyType)
	vm.DefineFunction("keyword-name", vm.vesperKeywordName, SymbolType, KeywordType)
	vm.DefineFunction("to-keyword", vm.vesperToKeyword, KeywordType, AnyType)
	vm.DefineFunction("symbol?", vesperSymbolP, BooleanType, AnyType)
	vm.DefineFunctionRestArgs("symbol", vm.vesperSymbol, SymbolType, AnyType, AnyType)

	vm.DefineFunctionRestArgs("string?", vesperStringP, BooleanType, AnyType)
	vm.DefineFunctionRestArgs("string", vesperString, StringType, AnyType)
	vm.DefineFunction("to-string", vesperToString, StringType, AnyType)
	vm.DefineFunction("string-length", vesperStringLength, NumberType, StringType)
	vm.DefineFunction("split", vesperSplit, ListType, StringType, StringType)
	vm.DefineFunction("join", vesperJoin, ListType, ListType, StringType)
	vm.DefineFunction("character?", vesperCharacterP, BooleanType, AnyType)
	vm.DefineFunction("to-character", vesperToCharacter, CharacterType, AnyType)
	vm.DefineFunction("substring", vesperSubstring, StringType, StringType, NumberType, NumberType)

	vm.DefineFunction("blob?", vesperBlobP, BooleanType, AnyType)
	vm.DefineFunction("to-blob", vesperToBlob, BlobType, AnyType)
	vm.DefineFunction("make-blob", vesperMakeBlob, BlobType, NumberType)
	vm.DefineFunction("blob-length", vesperBlobLength, NumberType, BlobType)
	vm.DefineFunction("blob-ref", vesperBlobRef, NumberType, BlobType, NumberType)

	vm.DefineFunction("number?", vesperNumberP, BooleanType, AnyType)
	vm.DefineFunction("int?", vesperIntP, BooleanType, AnyType)
	vm.DefineFunction("float?", vesperFloatP, BooleanType, AnyType)
	vm.DefineFunction("to-number", vesperToNumber, NumberType, AnyType)
	vm.DefineFunction("int", vesperInt, NumberType, AnyType)
	vm.DefineFunction("floor", vesperFloor, NumberType, NumberType)
	vm.DefineFunction("ceiling", vesperCeiling, NumberType, NumberType)
	vm.DefineFunction("inc", vesperInc, NumberType, NumberType)
	vm.DefineFunction("dec", vesperDec, NumberType, NumberType)
	vm.DefineFunction("+", vesperAdd, NumberType, NumberType, NumberType)
	vm.DefineFunction("-", vesperSub, NumberType, NumberType, NumberType)
	vm.DefineFunction("*", vesperMul, NumberType, NumberType, NumberType)
	vm.DefineFunction("/", vesperDiv, NumberType, NumberType, NumberType)
	vm.DefineFunction("quotient", vesperQuotient, NumberType, NumberType, NumberType)
	vm.DefineFunction("remainder", vesperRemainder, NumberType, NumberType, NumberType)
	vm.DefineFunction("modulo", vesperRemainder, NumberType, NumberType, NumberType)
	vm.DefineFunction("=", vesperNumEqual, BooleanType, NumberType, NumberType)
	vm.DefineFunction("<=", vesperNumLessEqual, BooleanType, NumberType, NumberType)
	vm.DefineFunction(">=", vesperNumGreaterEqual, BooleanType, NumberType, NumberType)
	vm.DefineFunction(">", vesperNumGreater, BooleanType, NumberType, NumberType)
	vm.DefineFunction("<", vesperNumLess, BooleanType, NumberType, NumberType)
	vm.DefineFunction("zero?", vesperZeroP, BooleanType, NumberType)
	vm.DefineFunction("abs", vesperAbs, NumberType, NumberType)
	vm.DefineFunction("exp", vesperExp, NumberType, NumberType)
	vm.DefineFunction("log", vesperLog, NumberType, NumberType)
	vm.DefineFunction("sin", vesperSin, NumberType, NumberType)
	vm.DefineFunction("cos", vesperCos, NumberType, NumberType)
	vm.DefineFunction("tan", vesperTan, NumberType, NumberType)
	vm.DefineFunction("asin", vesperAsin, NumberType, NumberType)
	vm.DefineFunction("acos", vesperAcos, NumberType, NumberType)
	vm.DefineFunction("atan", vesperAtan, NumberType, NumberType)
	vm.DefineFunction("atan2", vesperAtan2, NumberType, NumberType, NumberType)

	vm.DefineFunction("seal!", vesperSeal, AnyType, AnyType)

	vm.DefineFunction("list?", vesperListP, BooleanType, AnyType)
	vm.DefineFunction("empty?", vesperEmptyP, BooleanType, ListType)
	vm.DefineFunction("to-list", vesperToList, ListType, AnyType)
	vm.DefineFunction("cons", vesperCons, ListType, AnyType, ListType)
	vm.DefineFunction("car", vesperCar, AnyType, ListType)
	vm.DefineFunction("cdr", vesperCdr, ListType, ListType)
	vm.DefineFunction("set-car!", vesperSetCarBang, NullType, ListType, AnyType)
	vm.DefineFunction("set-cdr!", vesperSetCdrBang, NullType, ListType, ListType)
	vm.DefineFunction("list-length", vesperListLength, NumberType, ListType)
	vm.DefineFunction("reverse", vesperReverse, ListType, ListType)
	vm.DefineFunctionRestArgs("list", vesperList, ListType, AnyType)
	vm.DefineFunctionRestArgs("concat", vesperConcat, ListType, ListType)
	vm.DefineFunctionRestArgs("flatten", vesperFlatten, ListType, ListType)

	vm.DefineFunction("array?", vesperArrayP, BooleanType, AnyType)
	vm.DefineFunction("to-array", vesperToArray, ArrayType, AnyType)
	vm.DefineFunctionRestArgs("array", vesperArray, ArrayType, AnyType)
	vm.DefineFunctionOptionalArgs("make-array", vesperMakeArray, ArrayType, []*Object{NumberType, AnyType}, Null)
	vm.DefineFunction("array-length", vesperArrayLength, NumberType, ArrayType)
	vm.DefineFunction("array-ref", vesperArrayRef, AnyType, ArrayType, NumberType)
	vm.DefineFunction("array-set!", vesperArraySetBang, NullType, ArrayType, NumberType, AnyType)

	vm.DefineFunction("struct?", vesperStructP, BooleanType, AnyType)
	vm.DefineFunction("to-struct", vesperToStruct, StructType, AnyType)
	vm.DefineFunctionRestArgs("struct", vesperStruct, StructType, AnyType)
	vm.DefineFunction("make-struct", vesperMakeStruct, StructType, NumberType)
	vm.DefineFunction("struct-length", vesperStructLength, NumberType, StructType)
	vm.DefineFunction("has?", vesperHasP, BooleanType, StructType, AnyType)
	vm.DefineFunction("get", vesperGet, AnyType, StructType, AnyType)
	vm.DefineFunction("put!", vesperPutBang, NullType, StructType, AnyType, AnyType)
	vm.DefineFunction("unput!", vesperUnputBang, NullType, StructType, AnyType)
	vm.DefineFunction("keys", vesperKeys, ListType, AnyType)
	vm.DefineFunction("values", vesperValues, ListType, AnyType)

	vm.DefineFunction("function?", vesperFunctionP, BooleanType, AnyType)
	vm.DefineFunction("function-signature", vesperFunctionSignature, StringType, FunctionType)
	vm.DefineFunctionRestArgs("validate-keyword-arg-list", vm.vesperValidateKeywordArgList, ListType, KeywordType, ListType)
	vm.DefineFunction("slurp", vesperSlurp, StringType, StringType)
	vm.DefineFunctionKeyArgs("read", vm.vesperRead, AnyType, []*Object{StringType, TypeType}, []*Object{AnyType}, []*Object{vm.Intern("keys:")})
	vm.DefineFunctionKeyArgs("read-all", vm.vesperReadAll, AnyType, []*Object{StringType, TypeType}, []*Object{AnyType}, []*Object{vm.Intern("keys:")})
	vm.DefineFunction("spit", vesperSpit, NullType, StringType, StringType)
	vm.DefineFunctionKeyArgs("write", vesperWrite, NullType, []*Object{AnyType, StringType}, []*Object{EmptyString}, []*Object{vm.Intern("indent:")})
	vm.DefineFunctionKeyArgs("write-all", vesperWriteAll, NullType, []*Object{AnyType, StringType}, []*Object{EmptyString}, []*Object{vm.Intern("indent:")})
	vm.DefineFunctionRestArgs("print", vesperPrint, NullType, AnyType)
	vm.DefineFunctionRestArgs("println", vesperPrintln, NullType, AnyType)
	vm.DefineFunction("macroexpand", vm.vesperMacroexpand, AnyType, AnyType)
	vm.DefineFunction("compile", vm.vesperCompile, CodeType, AnyType)

	vm.DefineFunctionRestArgs("make-error", vesperMakeError, ErrorType, AnyType)
	vm.DefineFunction("error?", vesperErrorP, BooleanType, AnyType)
	vm.DefineFunction("error-data", vesperErrorData, AnyType, ErrorType)
	vm.DefineFunction("uncaught-error", vesperUncaughtError, NullType, ErrorType)

	vm.DefineFunctionKeyArgs("json", vesperJSON, StringType, []*Object{AnyType, StringType}, []*Object{EmptyString}, []*Object{vm.Intern("indent:")})

	vm.DefineFunctionRestArgs("getfn", vm.vesperGetFn, FunctionType, AnyType, SymbolType)
	vm.DefineFunction("method-signature", vm.vesperMethodSignature, TypeType, ListType)

	vm.DefineFunction("now", vesperNow, NumberType)
	vm.DefineFunction("since", vesperSince, NumberType, NumberType)
	vm.DefineFunction("sleep", vesperSleep, NumberType, NumberType)

	vm.DefineFunctionKeyArgs("channel", vesperChannel, ChannelType, []*Object{StringType, NumberType}, []*Object{EmptyString, Zero}, []*Object{vm.Intern("name:"), vm.Intern("bufsize:")})
	vm.DefineFunctionOptionalArgs("send", vesperSend, NullType, []*Object{ChannelType, AnyType, NumberType}, MinusOne)
	vm.DefineFunctionOptionalArgs("recv", vesperReceive, AnyType, []*Object{ChannelType, NumberType}, MinusOne)
	vm.DefineFunction("close", vesperClose, NullType, AnyType)

	vm.DefineFunction("set-random-seed!", vesperSetRandomSeedBang, NullType, NumberType)
	vm.DefineFunctionRestArgs("random", vesperRandom, NumberType, NumberType)

	vm.DefineFunction("timestamp", vesperTimestamp, StringType)

	vm.DefineFunction("getenv", vesperGetenv, StringType, StringType)
	vm.DefineFunction("load", vm.vesperLoad, StringType, AnyType)

	err := vm.Load("vesper")
	if err != nil {
		Fatal("*** ", err)
	}
}

func (vm *VM) vesperLetrec(argv []*Object) (*Object, error) {
	return vm.expandLetrec(argv[0])
}

func (vm *VM) vesperLet(argv []*Object) (*Object, error) {
	return vm.expandLet(argv[0])
}

func (vm *VM) vesperCond(argv []*Object) (*Object, error) {
	return vm.expandCond(argv[0])
}

func (vm *VM) vesperQuasiquote(argv []*Object) (*Object, error) {
	return vm.expandQuasiquote(argv[0])
}

// Actual primitive functions

func (vm *VM) vesperGlobals(_ []*Object) (*Object, error) {
	return Array(vm.Globals()...), nil
}

func (vm *VM) vesperVersion(_ []*Object) (*Object, error) {
	s := Version
	if len(vm.Extensions) > 0 {
		s += " (with "
		for i, ext := range vm.Extensions {
			if i > 0 {
				s += ", "
			}
			s += ext.Name()
		}
		s += ")"
	}
	return String(s), nil
}

func vesperDefinedP(argv []*Object) (*Object, error) {
	return toVesperBool(IsDefined(argv[0]))
}

func vesperSlurp(argv []*Object) (*Object, error) {
	return SlurpFile(argv[0].text)
}

func vesperSpit(argv []*Object) (*Object, error) {
	url := argv[0].text
	data := argv[1].text
	err := SpitFile(url, data)
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func (vm *VM) vesperRead(argv []*Object) (*Object, error) {
	return vm.Read(argv[0], argv[1])
}

func (vm *VM) vesperReadAll(argv []*Object) (*Object, error) {
	return vm.ReadAll(argv[0], argv[1])
}

func (vm *VM) vesperMacroexpand(argv []*Object) (*Object, error) {
	return vm.Macroexpand(argv[0])
}

func (vm *VM) vesperCompile(argv []*Object) (*Object, error) {
	expanded, err := vm.Macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return vm.Compile(expanded)
}

func (vm *VM) vesperLoad(argv []*Object) (*Object, error) {
	err := vm.Load(argv[0].text)
	return argv[0], err
}

func vesperType(argv []*Object) (*Object, error) {
	return argv[0].Type, nil
}

func vesperValue(argv []*Object) (*Object, error) {
	return Value(argv[0]), nil
}

func vesperInstance(argv []*Object) (*Object, error) {
	return Instance(argv[0], argv[1])
}

func (vm *VM) vesperValidateKeywordArgList(argv []*Object) (*Object, error) {
	return vm.validateKeywordArgList(argv[0], argv[1:])
}

func vesperKeys(argv []*Object) (*Object, error) {
	return structKeyList(argv[0]), nil
}

func vesperValues(argv []*Object) (*Object, error) {
	return structValueList(argv[0]), nil
}

func vesperStruct(argv []*Object) (*Object, error) {
	return Struct(argv)
}

func vesperMakeStruct(argv []*Object) (*Object, error) {
	return MakeStruct(int(argv[0].fval)), nil
}

func vesperToStruct(argv []*Object) (*Object, error) {
	return ToStruct(argv[0])
}

func vesperIdenticalP(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0] == argv[1])
}

func vesperEqualP(argv []*Object) (*Object, error) {
	return toVesperBool(Equal(argv[0], argv[1]))
}

func vesperNumEqual(argv []*Object) (*Object, error) {
	return toVesperBool(NumberEqual(argv[0].fval, argv[1].fval))
}

func vesperNumLess(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0].fval < argv[1].fval)
}

func vesperNumLessEqual(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0].fval <= argv[1].fval)
}

func vesperNumGreater(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0].fval > argv[1].fval)
}

func vesperNumGreaterEqual(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0].fval >= argv[1].fval)
}

func vesperWrite(argv []*Object) (*Object, error) {
	return String(writeIndent(argv[0], argv[1].text)), nil
}

func vesperWriteAll(argv []*Object) (*Object, error) {
	return String(writeAllIndent(argv[0], argv[1].text)), nil
}

func vesperMakeError(argv []*Object) (*Object, error) {
	return MakeError(argv...), nil
}

func vesperErrorP(argv []*Object) (*Object, error) {
	return toVesperBool(IsError(argv[0]))
}

func vesperErrorData(argv []*Object) (*Object, error) {
	return ErrorData(argv[0]), nil
}

func vesperUncaughtError(argv []*Object) (*Object, error) {
	return nil, argv[0]
}

func vesperToString(argv []*Object) (*Object, error) {
	return ToString(argv[0])
}

func vesperPrint(argv []*Object) (*Object, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func vesperPrintln(argv []*Object) (*Object, error) {
	_, _ = vesperPrint(argv)
	fmt.Println()
	return Null, nil
}

func vesperConcat(argv []*Object) (*Object, error) {
	result := EmptyList
	tail := result
	for _, lst := range argv {
		for lst != EmptyList {
			if tail == EmptyList {
				result = List(lst.car)
				tail = result
			} else {
				tail.cdr = List(lst.car)
				tail = tail.cdr
			}
			lst = lst.cdr
		}
	}
	return result, nil
}

func vesperReverse(argv []*Object) (*Object, error) {
	return Reverse(argv[0]), nil
}

func vesperFlatten(argv []*Object) (*Object, error) {
	return Flatten(argv[0]), nil
}

func vesperList(argv []*Object) (*Object, error) {
	argc := len(argv)
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func vesperListLength(argv []*Object) (*Object, error) {
	return Number(float64(ListLength(argv[0]))), nil
}

func vesperNumberP(argv []*Object) (*Object, error) {
	return toVesperBool(IsNumber(argv[0]))
}

func vesperToNumber(argv []*Object) (*Object, error) {
	return ToNumber(argv[0])
}

func vesperIntP(argv []*Object) (*Object, error) {
	return toVesperBool(IsInt(argv[0]))
}

func vesperFloatP(argv []*Object) (*Object, error) {
	return toVesperBool(IsFloat(argv[0]))
}

func vesperInt(argv []*Object) (*Object, error) {
	return ToInt(argv[0])
}

func vesperFloor(argv []*Object) (*Object, error) {
	return Number(math.Floor(argv[0].fval)), nil
}

func vesperCeiling(argv []*Object) (*Object, error) {
	return Number(math.Ceil(argv[0].fval)), nil
}

func vesperInc(argv []*Object) (*Object, error) {
	return Number(argv[0].fval + 1), nil
}

func vesperDec(argv []*Object) (*Object, error) {
	return Number(argv[0].fval - 1), nil
}

func vesperAdd(argv []*Object) (*Object, error) {
	return Number(argv[0].fval + argv[1].fval), nil
}

func vesperSub(argv []*Object) (*Object, error) {
	return Number(argv[0].fval - argv[1].fval), nil
}

func vesperMul(argv []*Object) (*Object, error) {
	return Number(argv[0].fval * argv[1].fval), nil
}

func vesperDiv(argv []*Object) (*Object, error) {
	return Number(argv[0].fval / argv[1].fval), nil
}

func vesperQuotient(argv []*Object) (*Object, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "quotient: divide by zero")
	}
	n := int64(argv[0].fval) / denom
	return Number(float64(n)), nil
}

func vesperRemainder(argv []*Object) (*Object, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "remainder: divide by zero")
	}
	n := int64(argv[0].fval) % denom
	return Number(float64(n)), nil
}

func vesperAbs(argv []*Object) (*Object, error) {
	return Number(math.Abs(argv[0].fval)), nil
}

func vesperExp(argv []*Object) (*Object, error) {
	return Number(math.Exp(argv[0].fval)), nil
}

func vesperLog(argv []*Object) (*Object, error) {
	return Number(math.Log(argv[0].fval)), nil
}

func vesperSin(argv []*Object) (*Object, error) {
	return Number(math.Sin(argv[0].fval)), nil
}

func vesperCos(argv []*Object) (*Object, error) {
	return Number(math.Cos(argv[0].fval)), nil
}

func vesperTan(argv []*Object) (*Object, error) {
	return Number(math.Tan(argv[0].fval)), nil
}

func vesperAsin(argv []*Object) (*Object, error) {
	return Number(math.Asin(argv[0].fval)), nil
}

func vesperAcos(argv []*Object) (*Object, error) {
	return Number(math.Acos(argv[0].fval)), nil
}

func vesperAtan(argv []*Object) (*Object, error) {
	return Number(math.Atan(argv[0].fval)), nil
}

func vesperAtan2(argv []*Object) (*Object, error) {
	return Number(math.Atan2(argv[0].fval, argv[1].fval)), nil
}

func vesperArray(argv []*Object) (*Object, error) {
	return Array(argv...), nil
}

func vesperToArray(argv []*Object) (*Object, error) {
	return ToArray(argv[0])
}

func vesperMakeArray(argv []*Object) (*Object, error) {
	vlen := int(argv[0].fval)
	init := argv[1]
	return MakeArray(vlen, init), nil
}

func vesperArrayP(argv []*Object) (*Object, error) {
	return toVesperBool(IsArray(argv[0]))
}

func vesperArrayLength(argv []*Object) (*Object, error) {
	return Number(float64(len(argv[0].elements))), nil
}

func vesperArrayRef(argv []*Object) (*Object, error) {
	el := argv[0].elements
	idx := int(argv[1].fval)
	if idx < 0 || idx >= len(el) {
		return nil, Error(ArgumentErrorKey, "Array index out of range")
	}
	return el[idx], nil
}

func vesperArraySetBang(argv []*Object) (*Object, error) {
	sealed := int(argv[0].fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "array-set! on sealed array")
	}
	el := argv[0].elements
	idx := int(argv[1].fval)
	if idx < 0 || idx > len(el) {
		return nil, Error(ArgumentErrorKey, "Array index out of range")
	}
	el[idx] = argv[2]
	return Null, nil
}

func vesperZeroP(argv []*Object) (*Object, error) {
	return toVesperBool(NumberEqual(argv[0].fval, 0.0))
}

func vesperNot(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0] == False)
}

func vesperNullP(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0] == Null)
}

func vesperBooleanP(argv []*Object) (*Object, error) {
	return toVesperBool(IsBoolean(argv[0]))
}

func vesperSymbolP(argv []*Object) (*Object, error) {
	return toVesperBool(IsSymbol(argv[0]))
}

func (vm *VM) vesperSymbol(argv []*Object) (*Object, error) {
	if len(argv) < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	return vm.Symbol(argv)
}

func vesperKeywordP(argv []*Object) (*Object, error) {
	return toVesperBool(IsKeyword(argv[0]))
}

func (vm *VM) vesperKeywordName(argv []*Object) (*Object, error) {
	return vm.KeywordName(argv[0])
}

func (vm *VM) vesperToKeyword(argv []*Object) (*Object, error) {
	return vm.ToKeyword(argv[0])
}

func vesperTypeP(argv []*Object) (*Object, error) {
	return toVesperBool(IsType(argv[0]))
}

func (vm *VM) vesperTypeName(argv []*Object) (*Object, error) {
	return vm.TypeName(argv[0])
}

func vesperStringP(argv []*Object) (*Object, error) {
	return toVesperBool(IsString(argv[0]))
}

func vesperCharacterP(argv []*Object) (*Object, error) {
	return toVesperBool(IsCharacter(argv[0]))
}

func vesperToCharacter(argv []*Object) (*Object, error) {
	return ToCharacter(argv[0])
}

func vesperSubstring(argv []*Object) (*Object, error) {
	s := argv[0].text
	start := int(argv[1].fval)
	end := int(argv[2].fval)
	if start < 0 {
		start = 0
	} else if start > len(s) {
		return String(""), nil
	}
	if end < start {
		return String(""), nil
	} else if end > len(s) {
		end = len(s)
	}
	return String(s[start:end]), nil
}

func vesperFunctionP(argv []*Object) (*Object, error) {
	return toVesperBool(IsFunction(argv[0]))
}

func vesperFunctionSignature(argv []*Object) (*Object, error) {
	return String(functionSignature(argv[0])), nil
}

func vesperListP(argv []*Object) (*Object, error) {
	return toVesperBool(IsList(argv[0]))
}

func vesperEmptyP(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0] == EmptyList)
}

func vesperString(argv []*Object) (*Object, error) {
	s := ""
	for _, ss := range argv {
		s += ss.String()
	}
	return String(s), nil
}

func vesperStringLength(argv []*Object) (*Object, error) {
	return Number(float64(StringLength(argv[0].text))), nil
}

func vesperCar(argv []*Object) (*Object, error) {
	lst := argv[0]
	if lst == EmptyList {
		return Null, nil
	}
	return lst.car, nil
}

func vesperCdr(argv []*Object) (*Object, error) {
	lst := argv[0]
	if lst == EmptyList {
		return lst, nil
	}
	return lst.cdr, nil
}

func vesperSetCarBang(argv []*Object) (*Object, error) {
	lst := argv[0]
	if lst == EmptyList {
		return nil, Error(ArgumentErrorKey, "set-car! expected a non-empty <list>")
	}
	sealed := int(lst.fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "set-car! on sealed list")
	}
	lst.car = argv[1]
	return Null, nil
}

func vesperSetCdrBang(argv []*Object) (*Object, error) {
	lst := argv[0]
	if lst == EmptyList {
		return nil, Error(ArgumentErrorKey, "set-cdr! expected a non-empty <list>")
	}
	sealed := int(lst.fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "set-cdr! on sealed list")
	}
	lst.cdr = argv[1]
	return Null, nil
}

func vesperCons(argv []*Object) (*Object, error) {
	return Cons(argv[0], argv[1]), nil
}

func vesperStructP(argv []*Object) (*Object, error) {
	return toVesperBool(IsStruct(argv[0]))
}

func vesperGet(argv []*Object) (*Object, error) {
	return structGet(argv[0], argv[1]), nil
}

func vesperStructLength(argv []*Object) (*Object, error) {
	return Number(float64(StructLength(argv[0]))), nil
}

func vesperHasP(argv []*Object) (*Object, error) {
	b, err := Has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	return toVesperBool(b)
}

func vesperSeal(argv []*Object) (*Object, error) {
	switch argv[0].Type {
	case StructType, ArrayType, ListType:
		argv[0].fval = 1
		return argv[0], nil
	default:
		return nil, Error(ArgumentErrorKey, "cannot seal! ", argv[0])
	}
}

func vesperPutBang(argv []*Object) (*Object, error) {
	key := argv[1]
	if !IsValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	sealed := int(argv[0].fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "put! on sealed struct")
	}
	Put(argv[0], key, argv[2])
	return Null, nil
}

func vesperUnputBang(argv []*Object) (*Object, error) {
	key := argv[1]
	if !IsValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	sealed := int(argv[0].fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "unput! on sealed struct")
	}
	Unput(argv[0], key)
	return Null, nil
}

func vesperToList(argv []*Object) (*Object, error) {
	return ToList(argv[0])
}

func vesperSplit(argv []*Object) (*Object, error) {
	return StringSplit(argv[0], argv[1])
}

func vesperJoin(argv []*Object) (*Object, error) {
	return StringJoin(argv[0], argv[1])
}

func vesperJSON(argv []*Object) (*Object, error) {
	s, err := writeToString(argv[0], true, argv[1].text)
	if err != nil {
		return nil, err
	}
	return String(s), nil
}

func (vm *VM) vesperGetFn(argv []*Object) (*Object, error) {
	if len(argv) < 1 {
		return nil, Error(ArgumentErrorKey, "getfn expected at least 1 argument, got none")
	}
	sym := argv[0]
	if sym.Type != SymbolType {
		return nil, Error(ArgumentErrorKey, "getfn expected a <symbol> for argument 1, got ", sym)
	}
	return vm.getfn(sym, argv[1:])
}

func (vm *VM) vesperMethodSignature(argv []*Object) (*Object, error) {
	return vm.methodSignature(argv[0])
}

func vesperChannel(argv []*Object) (*Object, error) {
	name := argv[0].text
	bufsize := int(argv[1].fval)
	return Channel(bufsize, name), nil
}

func vesperClose(argv []*Object) (*Object, error) {
	switch argv[0].Type {
	case ChannelType:
		CloseChannel(argv[0])
	default:
		return nil, Error(ArgumentErrorKey, "close expected a channel")
	}
	return Null, nil
}

func vesperSend(argv []*Object) (*Object, error) {
	ch := ChannelValue(argv[0])
	if ch != nil {
		val := argv[1]
		timeout := argv[2].fval
		if NumberEqual(timeout, 0.0) {
			select {
			case ch <- val:
				return True, nil
			default:
			}
		} else if timeout > 0 {
			dur := time.Millisecond * time.Duration(timeout*1000.0)
			select {
			case ch <- val:
				return True, nil
			case <-time.After(dur):
			}
		} else {
			ch <- val
			return True, nil
		}
	}
	return False, nil
}

func vesperReceive(argv []*Object) (*Object, error) {
	ch := ChannelValue(argv[0])
	if ch != nil {
		timeout := argv[1].fval
		if NumberEqual(timeout, 0.0) {
			select {
			case val, ok := <-ch:
				if ok && val != nil {
					return val, nil
				}
			default:
			}
		} else if timeout > 0 {
			dur := time.Millisecond * time.Duration(timeout*1000.0)
			select {
			case val, ok := <-ch:
				if ok && val != nil {
					return val, nil
				}
			case <-time.After(dur):
			}
		} else {
			val := <-ch
			if val != nil {
				return val, nil
			}
		}
	}
	return Null, nil
}

func vesperSetRandomSeedBang(argv []*Object) (*Object, error) {
	RandomSeed(int64(argv[0].fval))
	return Null, nil
}

func vesperRandom(argv []*Object) (*Object, error) {
	min := 0.0
	max := 1.0
	argc := len(argv)
	switch argc {
	case 0:
	case 1:
		max = argv[0].fval
	case 2:
		min = argv[0].fval
		max = argv[1].fval
	default:
		return nil, Error(ArgumentErrorKey, "random expected 0 to 2 arguments, got ", argc)
	}
	return Random(min, max), nil
}

// Timestamp returns a string timestamp
func Timestamp(t time.Time) *Object {
	format := "%d-%02d-%02dT%02d:%02d:%02d.%03dZ"
	return String(fmt.Sprintf(format, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000))
}

func vesperTimestamp(_ []*Object) (*Object, error) {
	return Timestamp(time.Now().UTC()), nil
}

func vesperBlobP(argv []*Object) (*Object, error) {
	return toVesperBool(argv[0].Type == BlobType)
}

func vesperToBlob(argv []*Object) (*Object, error) {
	return ToBlob(argv[0])
}

func vesperMakeBlob(argv []*Object) (*Object, error) {
	size := int(argv[0].fval)
	return MakeBlob(size), nil
}

func vesperBlobLength(argv []*Object) (*Object, error) {
	return Number(float64(len(BlobValue(argv[0])))), nil
}

func vesperBlobRef(argv []*Object) (*Object, error) {
	el := BlobValue(argv[0])
	idx := int(argv[1].fval)
	if idx < 0 || idx >= len(el) {
		return nil, Error(ArgumentErrorKey, "Blob index out of range")
	}
	return Number(float64(el[idx])), nil
}

func vesperNow(_ []*Object) (*Object, error) {
	return Number(Now()), nil
}

func vesperSince(argv []*Object) (*Object, error) {
	then := argv[0].fval
	dur := Now() - then
	return Number(dur), nil
}

func vesperSleep(argv []*Object) (*Object, error) {
	Sleep(argv[0].fval)
	return Number(Now()), nil
}

func vesperGetenv(argv []*Object) (*Object, error) {
	s := os.Getenv(argv[0].text)
	if s == "" {
		return Null, nil
	}
	return String(s), nil
}

func toVesperBool(b bool) (*Object, error) {
	if b {
		return True, nil
	}
	return False, nil
}
