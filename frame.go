package vesper

import (
	"fmt"
	"strings"
)

type frame struct {
	locals    *frame
	previous  *frame
	code      *Code
	ops       []int
	elements  []*Object
	firstfive [5]*Object
	pc        int
}

func (frame *frame) String() string {
	var buf strings.Builder
	buf.WriteString("#[frame ")
	if frame.code != nil {
		if frame.code.name != "" {
			buf.WriteString(" " + frame.code.name)
		} else {
			buf.WriteString(" (anonymous code)")
		}
	} else {
		buf.WriteString(" (no code)")
	}
	buf.WriteString(fmt.Sprintf(" previous: %v", frame.previous))
	buf.WriteString("]")
	return buf.String()
}

func (vm *VM) buildFrame(env *frame, pc int, ops []int, fun *Object, argc int, stack []*Object, sp int) (*frame, error) {
	f := &frame{
		previous: env,
		pc:       pc,
		ops:      ops,
		locals:   fun.frame,
		code:     fun.code,
	}
	expectedArgc := fun.code.argc
	defaults := fun.code.defaults
	if defaults == nil {
		if argc != expectedArgc {
			return nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
		}
		if argc <= 5 {
			f.elements = f.firstfive[:]
		} else {
			f.elements = make([]*Object, argc)
		}
		copy(f.elements, stack[sp:sp+argc])
		return f, nil
	}
	keys := fun.code.keys
	rest := false
	extra := len(defaults)
	if extra == 0 {
		rest = true
		extra = 1
	}
	if argc < expectedArgc {
		if extra > 0 {
			return nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected at least ", expectedArgc, ", got ", argc, ")")
		}
		return nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
	}
	totalArgc := expectedArgc + extra
	el := make([]*Object, totalArgc)
	end := sp + expectedArgc
	if rest {
		copy(el, stack[sp:end])
		restElements := stack[end : sp+argc]
		el[expectedArgc] = ListFromValues(restElements)
	} else if keys != nil {
		bindings := stack[sp+expectedArgc : sp+argc]
		if len(bindings)%2 != 0 {
			return nil, Error(ArgumentErrorKey, "Bad keyword argument(s): ", bindings)
		}
		copy(el, stack[sp:sp+expectedArgc]) //the required ones
		for i := expectedArgc; i < totalArgc; i++ {
			el[i] = defaults[i-expectedArgc]
		}
		for i := expectedArgc; i < argc; i += 2 {
			key, err := vm.ToSymbol(stack[sp+i])
			if err != nil {
				return nil, Error(ArgumentErrorKey, "Bad keyword argument: ", stack[sp+1])
			}
			gotit := false
			for j := 0; j < extra; j++ {
				if keys[j] == key {
					el[expectedArgc+j] = stack[sp+i+1]
					gotit = true
					break
				}
			}
			if !gotit {
				return nil, Error(ArgumentErrorKey, "Undefined keyword argument: ", key)
			}
		}
	} else {
		copy(el, stack[sp:sp+argc])
		for i := argc; i < totalArgc; i++ {
			el[i] = defaults[i-expectedArgc]
		}
	}
	f.elements = el
	return f, nil
}
