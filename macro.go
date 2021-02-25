package vesper

import (
	"fmt"
)

type macro struct {
	name     *Object
	expander *Object //a function of one argument
}

// Macro - create a new Macro
func Macro(name *Object, expander *Object) *macro {
	return &macro{name, expander}
}

func (mac *macro) String() string {
	return fmt.Sprintf("(macro %v %v)", mac.name, mac.expander)
}
