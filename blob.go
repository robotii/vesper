package vesper

// Blob - create a new blob, using the specified byte slice as the data. The data is not copied.
func Blob(bytes []byte) *Object {
	return &Object{
		Type:  BlobType,
		Value: bytes,
	}
}

