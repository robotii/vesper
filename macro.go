package vesper

import (
	"fmt"
)

type macro struct {
	name     *Object
	expander *Object //a function of one argument
}
