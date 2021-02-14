package kana

import "math"

const epsilon = 0.000000001

// Number - create a Number object for the given value
func Number(f float64) *Object {
	return &Object{
		Type: NumberType,
		fval: f,
	}
}

// NumberEqual returns true if the object is equal to the argument, within epsilon
func NumberEqual(f1 float64, f2 float64) bool {
	if f1 == f2 {
		return true
	}
	if math.Abs(f1-f2) < epsilon {
		return true
	}
	return false
}
