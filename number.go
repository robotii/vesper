package vesper

import (
	"math"
	"math/rand"
	"strconv"
)

const epsilon = 0.000000001

// Zero is the Vesper 0 value
var Zero = Number(0)

// One is the Vesper 1 value
var One = Number(1)

// MinusOne is the Vesper -1 value
var MinusOne = Number(-1)

// Number - create a Number object for the given value
func Number(f float64) *Object {
	return &Object{
		Type: NumberType,
		fval: f,
	}
}

// Int converts an integer to a vector Number
func Int(n int64) *Object {
	return Number(float64(n))
}

// Round - return the closest integer value to the float value
func Round(f float64) float64 {
	if f > 0 {
		return math.Floor(f + 0.5)
	}
	return math.Ceil(f - 0.5)
}

// ToNumber - convert object to a number, if possible
func ToNumber(o *Object) (*Object, error) {
	switch o.Type {
	case NumberType:
		return o, nil
	case CharacterType:
		return Number(o.fval), nil
	case BooleanType:
		return Number(o.fval), nil
	case StringType:
		f, err := strconv.ParseFloat(o.text, 64)
		if err == nil {
			return Number(f), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "cannot convert to an number: ", o)
}

// ToInt - convert the object to an integer number, if possible
func ToInt(o *Object) (*Object, error) {
	switch o.Type {
	case NumberType:
		return Number(Round(o.fval)), nil
	case CharacterType:
		return Number(o.fval), nil
	case BooleanType:
		return Number(o.fval), nil
	case StringType:
		n, err := strconv.ParseInt(o.text, 10, 64)
		if err == nil {
			return Number(float64(n)), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "cannot convert to an integer: ", o)
}

// IsInt returns true if the object is an integer
func IsInt(obj *Object) bool {
	if obj.Type == NumberType {
		f := obj.fval
		return math.Trunc(f) == f
	}
	return false
}

// IsFloat returns true if the object is a float
func IsFloat(obj *Object) bool {
	if obj.Type == NumberType {
		return !IsInt(obj)
	}
	return false
}

// AsFloat64Value returns the floating point value of the object
func AsFloat64Value(obj *Object) (float64, error) {
	if obj.Type == NumberType {
		return obj.fval, nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

// AsInt64Value returns the int64 value of the object
func AsInt64Value(obj *Object) (int64, error) {
	if obj.Type == NumberType {
		return int64(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

// AsIntValue returns the int value of the object
func AsIntValue(obj *Object) (int, error) {
	if obj.Type == NumberType {
		return int(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

// AsByteValue returns teh value of the object as a byte
func AsByteValue(obj *Object) (byte, error) {
	if obj.Type == NumberType {
		return byte(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

// NumberEqual returns true if the object is equal to the argument, within epsilon
func NumberEqual(f1 float64, f2 float64) bool {
	return f1 == f2 || math.Abs(f1-f2) < epsilon
}

var randomGenerator = rand.New(rand.NewSource(1))

// RandomSeed seeds the random number generator with the given seed value
func RandomSeed(n int64) {
	randomGenerator = rand.New(rand.NewSource(n))
}

// Random returns a random value within the given range
func Random(min float64, max float64) *Object {
	return Number(min + (randomGenerator.Float64() * (max - min)))
}

// RandomList returns a list of random numbers with the given size and range
func RandomList(size int, min float64, max float64) *Object {
	result := EmptyList
	tail := EmptyList
	for i := 0; i < size; i++ {
		tmp := List(Random(min, max))
		if result == EmptyList {
			result = tmp
			tail = tmp
		} else {
			tail.cdr = tmp
			tail = tmp
		}
	}
	return result
}
