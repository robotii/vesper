package vesper

import (
	"fmt"
	"os"
	"time"
)

// Print prints the expressions
func Print(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(str(args[i]))
	}
	fmt.Print(str(args[max]))
}

// Println prints the values followed by a newline
func Println(args ...interface{}) {
	Print(args...)
	fmt.Println()
}

// Fatal prints the values, followed by immediate exit
func Fatal(args ...interface{}) {
	Println(args...)
	os.Exit(1)
}

// Sleep for the given number of seconds
func Sleep(delayInSeconds float64) {
	dur := time.Duration(delayInSeconds * float64(time.Second))
	time.Sleep(dur)
}

// Now returns the time in seconds  since the epoch
func Now() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / float64(time.Second)
}

func str(o interface{}) string {
	if lob, ok := o.(*Object); ok {
		return lob.String()
	}
	return fmt.Sprintf("%v", o)
}

func copyEnv(src map[string]*Object) map[string]*Object {
	m := make(map[string]*Object, len(src))
	for k, v := range src {
		m[k] = v
	}
	return m
}

func copyMacros(src map[*Object]*Macro) map[*Object]*Macro {
	m := make(map[*Object]*Macro, len(src))
	for k, v := range src {
		m[k] = v
	}
	return m
}

func copyConstantMap(src map[*Object]int) map[*Object]int {
	m := make(map[*Object]int, len(src))
	for k, v := range src {
		m[k] = v
	}
	return m
}

func copyConstants(src []*Object) []*Object {
	m := make([]*Object, len(src))
	copy(m, src)
	return m
}
