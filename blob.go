package vesper

// Blob - create a new blob, using the specified byte slice as the data. The data is not copied.
func Blob(bytes []byte) *Object {
	return &Object{
		Type:  BlobType,
		Value: bytes,
	}
}

// MakeBlob - create a new blob of the given size. It will be initialized to all zeroes
func MakeBlob(size int) *Object {
	return Blob(make([]byte, size))
}

// EmptyBlob - a blob with no bytes
var EmptyBlob = MakeBlob(0)

