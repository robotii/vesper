package kana

// Code represents compiled bytecode that can be called
type Code struct {
	name     string
	ops      []int
	argc     int
	defaults []*Object
	keys     []*Object
}
