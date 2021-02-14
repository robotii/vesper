package kana

// Object represents all objects in Kana
type Object struct {
	Type         *Object               // i.e. <string>
	code         *Code                 // non-nil for closure, code
	frame        *frame                // non-nil for closure, continuation
	primitive    *primitive            // non-nil for primitives
	continuation *continuation         // non-nil for continuation
	car          *Object               // non-nil for instances and lists
	cdr          *Object               // non-nil for slists, nil for everything else
	bindings     map[structKey]*Object // non-nil for struct
	elements     []*Object             // non-nil for vector
	fval         float64               // number
	text         string                // string, symbol, keyword, type
	Value        interface{}           // the rest of the data for more complex things
}
