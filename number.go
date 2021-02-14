package kana

// Number - create a Number object for the given value
func Number(f float64) *Object {
	return &Object{
		Type: NumberType,
		fval: f,
	}
}
