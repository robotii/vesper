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

// ToBlob - convert argument to a blob, if possible.
func ToBlob(obj *Object) (*Object, error) {
	switch obj.Type {
	case BlobType:
		return obj, nil
	case StringType:
		return Blob([]byte(obj.text)), nil //this copies the data
	case ArrayType:
		return arrayToBlob(obj)
	default:
		return nil, Error(ArgumentErrorKey, "to-blob expected <blob> or <string>, got a ", obj.Type)
	}
}

func arrayToBlob(obj *Object) (*Object, error) {
	el := obj.elements
	n := len(el)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		val, err := AsByteValue(el[i])
		if err != nil {
			return nil, err
		}
		b[i] = val
	}
	return Blob(b), nil
}
