package kana

// IsSymbol returns true if the object is a symbol
func IsSymbol(obj *Object) bool {
	return obj.Type == SymbolType
}

// IsKeyword returns true if the object is a keyword
func IsKeyword(obj *Object) bool {
	return obj.Type == KeywordType
}

// Intern puts a string into the global environment
func Intern(name string) *Object {
	// TODO: Implement
	return nil
}
