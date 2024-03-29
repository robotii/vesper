package vesper

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultIndentSize = "    "

type dataReader struct {
	vm  *VM
	in  *bufio.Reader
	pos int
}

// IsDirectoryReadable - return true of the directory is readable
func IsDirectoryReadable(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.Mode().IsDir()
	}
	return false
}

// IsFileReadable - return true of the file is readable
func IsFileReadable(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.Mode().IsRegular()
	}
	return false
}

// ExpandFilePath returns the absolute path of a file or directory
func ExpandFilePath(path string) string {
	expanded, err := ExpandPath(path)
	if err != nil {
		return path
	}
	return expanded
}

// ExpandPath returns the absolute path including
func ExpandPath(filename string) (string, error) {
	if filepath.IsAbs(filename) {
		return filepath.Clean(filename), nil
	}

	if filename == "" {
		return "", fmt.Errorf("empty path specified")
	}

	if filename[0] == '~' {
		home := os.Getenv("HOME")
		if home != "" {
			return filepath.Join(home, filename[1:]), nil
		}
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(u.HomeDir, filename[1:]), nil
	}

	p, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	return p, nil
}

// SlurpFile - return the file contents as a string
func SlurpFile(path string) (*Object, error) {
	path = ExpandFilePath(path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return EmptyString, err
	}
	return String(string(b)), nil
}

// SpitFile - write the string to the file.
func SpitFile(path string, data string) error {
	path = ExpandFilePath(path)
	return ioutil.WriteFile(path, []byte(data), 0644)
}

// Read - only reads the first item in the input, along with how many characters it read
// for subsequence calls, you can slice the string to continue
func (vm *VM) Read(input *Object, keys *Object) (*Object, error) {
	if !IsString(input) {
		return nil, Error(ArgumentErrorKey, "read invalid input: ", input)
	}
	r := strings.NewReader(input.text)
	reader := vm.newDataReader(r)
	obj, err := reader.readData(keys)
	if err != nil {
		if err == io.EOF {
			return Null, nil
		}
		return nil, err
	}
	return obj, nil
}

// ReadAll - read all items in the input, returning a list of them.
func (vm *VM) ReadAll(input *Object, keys *Object) (*Object, error) {
	if !IsString(input) {
		return nil, Error(ArgumentErrorKey, "read-all invalid input: ", input)
	}
	reader := vm.newDataReader(strings.NewReader(input.text))
	lst := EmptyList
	tail := EmptyList
	val, err := reader.readData(keys)
	for err == nil {
		if lst == EmptyList {
			lst = List(val)
			tail = lst
		} else {
			tail.cdr = List(val)
			tail = tail.cdr
		}
		val, err = reader.readData(keys)
	}
	if err != io.EOF {
		return nil, err
	}
	return lst, nil
}

func (vm *VM) newDataReader(in io.Reader) *dataReader {
	br := bufio.NewReader(in)
	return &dataReader{vm, br, 0}
}

func (dr *dataReader) getChar() (rune, error) {
	r, _, e := dr.in.ReadRune()
	if e != nil {
		return 0, e
	}
	dr.pos++
	return r, nil
}

func (dr *dataReader) ungetChar() error {
	e := dr.in.UnreadRune()
	if e != nil {
		dr.pos--
	}
	return e
}

func (dr *dataReader) readData(keys *Object) (*Object, error) {
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
			c, e = dr.getChar()
			continue
		}
		switch c {
		case ';':
			e = dr.decodeComment()
			if e != nil {
				break
			} else {
				c, e = dr.getChar()
			}
		case '\'':
			o, err := dr.readData(keys)
			if err != nil {
				return nil, err
			}
			if o == nil {
				return o, nil
			}
			return List(QuoteSymbol, o), nil
		case '`':
			o, err := dr.readData(keys)
			if err != nil {
				return nil, err
			}
			return List(QuasiquoteSymbol, o), nil
		case '~', '^':
			c, e := dr.getChar()
			if e != nil {
				return nil, e
			}
			sym := UnquoteSymbol
			if c != '@' {
				_ = dr.ungetChar()
			} else {
				sym = UnquoteSplicingSymbol
			}
			o, err := dr.readData(keys)
			if err != nil {
				return nil, err
			}
			return List(sym, o), nil
		case '#':
			return dr.decodeReaderMacro(keys)
		case '(':
			return dr.decodeList(keys)
		case '[':
			return dr.decodeArray(keys)
		case '{':
			return dr.decodeStruct(keys)
		case '"':
			return dr.decodeString()
		case ')', ']', '}':
			return nil, Error(SyntaxErrorKey, "Unexpected '", string(c), "'")
		default:
			atom, err := dr.decodeAtom(c)
			return atom, err
		}
	}
	return nil, e
}

func (dr *dataReader) decodeComment() error {
	c, e := dr.getChar()
	for e == nil {
		if c == '\n' {
			return nil
		}
		c, e = dr.getChar()
	}
	return e
}

func (dr *dataReader) decodeString() (*Object, error) {
	var buf []rune
	c, e := dr.getChar()
	escape := false
	for e == nil {
		if escape {
			escape = false
			switch c {
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'f':
				buf = append(buf, '\f')
			case 'b':
				buf = append(buf, '\b')
			case 'r':
				buf = append(buf, '\r')
			case 'x':
				r, e := dr.decodeUnicode(2)
				if e != nil {
					return nil, e
				}
				buf = append(buf, r)
			case 'u':
				r, e := dr.decodeUnicode(4)
				if e != nil {
					return nil, e
				}
				buf = append(buf, r)
			case 'U':
				r, e := dr.decodeUnicode(8)
				if e != nil {
					return nil, e
				}
				buf = append(buf, r)
			default:
				buf = append(buf, c)
			}
		} else if c == '"' {
			break
		} else if c == '\\' {
			escape = true
		} else {
			escape = false
			buf = append(buf, c)
		}
		c, e = dr.getChar()
	}
	return String(string(buf)), e
}

func (dr *dataReader) decodeUnicode(size int) (rune, error) {
	var buf []rune
	for i := 0; i < size; i++ {
		c, e := dr.getChar()
		if e != nil {
			return 0, e
		}
		buf = append(buf, c)
	}
	r, e := strconv.ParseInt(string(buf), 16, 32)
	if e != nil {
		return 0, e
	}
	return rune(r), nil
}

func (dr *dataReader) decodeList(keys *Object) (*Object, error) {
	items, err := dr.decodeSequence(')', keys)
	if err != nil {
		return nil, err
	}
	return ListFromValues(items), nil
}

func (dr *dataReader) decodeArray(keys *Object) (*Object, error) {
	items, err := dr.decodeSequence(']', keys)
	if err != nil {
		return nil, err
	}
	return Array(items...), nil
}

func (dr *dataReader) skipToData(skipColon bool) (rune, error) {
	c, err := dr.getChar()
	for err == nil {
		if isWhitespace(c) || (skipColon && c == ':') {
			c, err = dr.getChar()
			continue
		}
		if c == ';' {
			err = dr.decodeComment()
			if err == nil {
				c, err = dr.getChar()
			}
			continue
		}
		return c, nil
	}
	return 0, err
}

func (dr *dataReader) decodeStruct(keys *Object) (*Object, error) {
	var items []*Object
	var err error
	var c rune
	for {
		c, err = dr.skipToData(false)
		if err != nil {
			return nil, err
		}
		if c == ':' {
			return nil, Error(SyntaxErrorKey, "Unexpected ':' in struct")
		}
		if c == '}' {
			return Struct(items)
		}
		err = dr.ungetChar()
		if err != nil {
			return nil, err
		}
		element, err := dr.readData(nil)
		if err != nil {
			return nil, err
		}
		if keys != nil && keys != AnyType {
			switch keys {
			case KeywordType:
				element, err = dr.vm.ToKeyword(element)
				if err != nil {
					return nil, err
				}
			case SymbolType:
				element, err = dr.vm.ToSymbol(element)
				if err != nil {
					return nil, err
				}
			case StringType:
				element, err = ToString(element)
				if err != nil {
					return nil, err
				}
			}
		}
		items = append(items, element)
		c, err = dr.skipToData(true)
		if err != nil {
			return nil, err
		}
		if c == '}' {
			return nil, Error(SyntaxErrorKey, "mismatched key/value in struct")
		}
		err = dr.ungetChar()
		if err != nil {
			return nil, err
		}
		element, err = dr.readData(keys)
		if err != nil {
			return nil, err
		}
		items = append(items, element)
	}
}

func (dr *dataReader) decodeSequence(endChar rune, keys *Object) ([]*Object, error) {
	c, err := dr.getChar()
	var items []*Object
	for err == nil {
		if isWhitespace(c) {
			c, err = dr.getChar()
			continue
		}
		if c == ';' {
			err = dr.decodeComment()
			if err == nil {
				c, err = dr.getChar()
			}
			continue
		}
		if c == endChar {
			return items, nil
		}
		_ = dr.ungetChar()
		element, err := dr.readData(keys)
		if err != nil {
			return nil, err
		}
		items = append(items, element)
		c, err = dr.getChar()
		if err != nil {
			return nil, err
		}
	}
	return nil, err
}

func (dr *dataReader) decodeAtom(firstChar rune) (*Object, error) {
	s, err := dr.decodeAtomString(firstChar)
	if err != nil {
		return nil, err
	}
	slen := len(s)
	keyword := false
	if s[slen-1] == ':' {
		keyword = true
		s = s[:slen-1]
	} else {
		if s == "null" {
			return Null, nil
		} else if s == "true" {
			return True, nil
		} else if s == "false" {
			return False, nil
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		if keyword {
			return nil, Error(SyntaxErrorKey, "Keyword cannot have a name that looks like a number: ", s, ":")
		}
		return Number(f), nil
	}
	if keyword {
		s += ":"
	}
	sym := dr.vm.Intern(s)
	return sym, nil
}

func (dr *dataReader) decodeAtomString(firstChar rune) (string, error) {
	var buf []rune
	if firstChar != 0 {
		if firstChar == ':' {
			return "", Error(SyntaxErrorKey, "Invalid keyword: colons only valid at the end of symbols")
		}
		buf = append(buf, firstChar)
	}
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
			break
		}
		if c == ':' {
			buf = append(buf, c)
			break
		}
		if isDelimiter(c) {
			_ = dr.ungetChar()
			break
		}
		buf = append(buf, c)
		c, e = dr.getChar()
	}
	if e != nil && e != io.EOF {
		return "", e
	}
	s := string(buf)
	return s, nil
}

func (dr *dataReader) decodeType(firstChar rune) (string, error) {
	var buf []rune
	if firstChar != '<' {
		return "", Error(SyntaxErrorKey, "Invalid type name")
	}
	buf = append(buf, firstChar)
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
			break
		}
		if c == '>' {
			buf = append(buf, c)
			break
		}
		if isDelimiter(c) {
			_ = dr.ungetChar()
			break
		}
		buf = append(buf, c)
		c, e = dr.getChar()
	}
	if e != nil && e != io.EOF {
		return "", e
	}
	s := string(buf)
	return s, nil
}

func namedChar(name string) (rune, error) {
	switch name {
	case "null":
		return 0, nil
	case "alarm":
		return 7, nil
	case "backspace":
		return 8, nil
	case "tab":
		return 9, nil
	case "newline":
		return 10, nil
	case "return":
		return 13, nil
	case "escape":
		return 27, nil
	case "space":
		return 32, nil
	case "delete":
		return 127, nil
	default:
		if strings.HasPrefix(name, "x") {
			hex := name[1:]
			i, err := strconv.ParseInt(hex, 16, 64)
			if err != nil {
				return 0, err
			}
			return rune(i), nil
		}
		return 0, Error(SyntaxErrorKey, "Bad named character: #\\", name)
	}
}

func (dr *dataReader) decodeReaderMacro(keys *Object) (*Object, error) {
	c, e := dr.getChar()
	if e != nil {
		return nil, e
	}
	switch c {
	case '\\': // to handle character literals.
		c, e = dr.getChar()
		if e != nil {
			return nil, e
		}
		if isWhitespace(c) || isDelimiter(c) {
			return Character(rune(c)), nil
		}
		c2, e := dr.getChar()
		if e != nil {
			if e != io.EOF {
				return nil, e
			}
			c2 = 32
		}
		if !isWhitespace(c2) && !isDelimiter(c2) {
			var name []rune
			name = append(name, c)
			name = append(name, c2)
			c, e = dr.getChar()
			for (e == nil || e != io.EOF) && !isWhitespace(c) && !isDelimiter(c) {
				name = append(name, c)
				c, e = dr.getChar()
			}
			if e != io.EOF && e != nil {
				return nil, e
			}
			_ = dr.ungetChar()
			r, e := namedChar(string(name))
			if e != nil {
				return nil, e
			}
			return Character(r), nil
		} else if e == nil {
			_ = dr.ungetChar()
		}
		return Character(rune(c)), nil
	case '!':
		err := dr.decodeComment()
		return Null, err
	case '[':
		s, err := dr.decodeAtomString(0)
		if err != nil {
			return nil, err
		}
		return nil, Error(SyntaxErrorKey, "Unreadable object: #[", s, "]")
	default:
		atom, err := dr.decodeType(c)
		if err != nil {
			return nil, err
		}
		if IsValidTypeName(atom) {
			val, err := dr.readData(keys)
			if err != nil {
				return nil, Error(SyntaxErrorKey, "Bad reader macro: #", atom, " ...")
			}
			return Instance(dr.vm.Intern(atom), val)
		}
		return nil, Error(SyntaxErrorKey, "Bad reader macro: #", atom, " ...")
	}
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r' || r == ','
}

func isDelimiter(r rune) bool {
	return r == '(' || r == ')' || r == '[' || r == ']' || r == '{' || r == '}' || r == '"' || r == '\'' || r == ';' || r == ':'
}

// Write returns a string representation of the object
func Write(obj *Object) string {
	return writeIndent(obj, "")
}

// Pretty returns a pretty-printed version of the object
func Pretty(obj *Object) string {
	return writeIndent(obj, defaultIndentSize)
}

func writeIndent(obj *Object, indentSize string) string {
	s, _ := writeToString(obj, false, indentSize)
	return s
}

// WriteAll writes all the objects in the list to a string
func WriteAll(obj *Object) string {
	return writeAllIndent(obj, "")
}

// PrettyAll - pretty-prints all the objects in the list to a string
func PrettyAll(obj *Object) string {
	return writeAllIndent(obj, "    ")
}

func writeAllIndent(obj *Object, indent string) string {
	if IsList(obj) {
		var buf strings.Builder
		for obj != EmptyList {
			o := Car(obj)
			s, _ := writeToString(o, false, indent)
			buf.WriteString(s)
			buf.WriteString("\n")
			obj = Cdr(obj)
		}
		return buf.String()
	}
	s, _ := writeToString(obj, false, indent)
	if indent == "" {
		return s + "\n"
	}
	return s
}

func writeToString(obj *Object, json bool, indentSize string) (string, error) {
	strdn, err := writeData(obj, json, "", indentSize)
	if err != nil {
		return "", err
	}
	if indentSize != "" {
		return strdn + "\n", nil
	}
	return strdn, nil
}

func writeData(obj *Object, json bool, indent string, indentSize string) (string, error) {
	switch obj.Type {
	case BooleanType, NullType, NumberType:
		return obj.String(), nil
	case ListType:
		if json {
			return writeArray(listToArray(obj), json, indent, indentSize)
		}
		return writeList(obj, indent, indentSize), nil
	case KeywordType:
		if json {
			return EncodeString(unkeywordedString(obj)), nil
		}
		return obj.String(), nil
	case SymbolType, TypeType:
		return obj.String(), nil
	case StringType:
		return EncodeString(obj.text), nil
	case ArrayType:
		return writeArray(obj, json, indent, indentSize)
	case StructType:
		return writeStruct(obj, json, indent, indentSize)
	case CharacterType:
		c := rune(obj.fval)
		switch c {
		case 0:
			return "#\\null", nil
		case 7:
			return "#\\alarm", nil
		case 8:
			return "#\\backspace", nil
		case 9:
			return "#\\tab", nil
		case 10:
			return "#\\newline", nil
		case 13:
			return "#\\return", nil
		case 27:
			return "#\\escape", nil
		case 32:
			return "#\\space", nil
		case 127:
			return "#\\delete", nil
		default:
			if c < 127 && c > 32 {
				return "#\\" + string(c), nil
			}
			return fmt.Sprintf("#\\x%04X", c), nil
		}
	default:
		if json {
			return "", Error(ArgumentErrorKey, "Data cannot be described in JSON: ", obj)
		}
		if obj == nil {
			return "", Error(ArgumentErrorKey, "Data cannot be nil")
		}
		return obj.String(), nil
	}
}

func writeList(lst *Object, indent string, indentSize string) string {
	if lst == EmptyList {
		return "()"
	}
	if lst.cdr != EmptyList {
		if lst.car == QuoteSymbol {
			return "'" + Cadr(lst).String()
		} else if lst.car == QuasiquoteSymbol {
			return "`" + Cadr(lst).String()
		} else if lst.car == UnquoteSymbol {
			return "~" + Cadr(lst).String()
		} else if lst.car == UnquoteSplicingSymbol {
			return "~@" + Cadr(lst).String()
		}
	}
	var buf strings.Builder
	buf.WriteString("(")
	delim := " "
	nextIndent := ""
	if indentSize != "" {
		nextIndent = indent + indentSize
		delim = "\n" + nextIndent
		buf.WriteString("\n" + nextIndent)
	}
	s, _ := writeData(lst.car, false, nextIndent, indentSize)
	buf.WriteString(s)
	lst = lst.cdr
	for lst != EmptyList {
		buf.WriteString(delim)
		s, _ := writeData(lst.car, false, nextIndent, indentSize)
		buf.WriteString(s)
		lst = lst.cdr
	}
	if indentSize != "" {
		buf.WriteString("\n" + indent)
	}
	buf.WriteString(")")
	return buf.String()
}

func writeArray(a *Object, json bool, indent string, indentSize string) (string, error) {
	var buf strings.Builder
	buf.WriteString("[")
	vlen := len(a.elements)
	if vlen > 0 {
		delim := ""
		if json {
			delim = ","
		}
		nextIndent := ""
		if indentSize != "" {
			nextIndent = indent + indentSize
			delim = delim + "\n" + nextIndent
			buf.WriteString("\n" + nextIndent)
		} else {
			delim = delim + " "
		}
		s, err := writeData(a.elements[0], json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		for i := 1; i < vlen; i++ {
			s, err := writeData(a.elements[i], json, nextIndent, indentSize)
			if err != nil {
				return "", err
			}
			buf.WriteString(delim)
			buf.WriteString(s)
		}
	}
	if indentSize != "" {
		buf.WriteString("\n" + indent)
	}
	buf.WriteString("]")
	return buf.String(), nil
}

func writeStruct(strct *Object, json bool, indent string, indentSize string) (string, error) {
	var buf strings.Builder
	buf.WriteString("{")
	size := len(strct.bindings)
	delim := ""
	sep := " "
	if json {
		delim = ","
		sep = ": "
	}
	nextIndent := ""
	if size > 0 {
		if indentSize != "" {
			nextIndent = indent + indentSize
			delim = delim + "\n" + nextIndent
			buf.WriteString("\n" + nextIndent)
		} else {
			delim = delim + " "
		}
	}
	first := true
	for k, v := range strct.bindings {
		if first {
			first = false
		} else {
			buf.WriteString(delim)
		}
		s, err := writeData(k, json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		buf.WriteString(sep)
		s, err = writeData(v, json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
	}
	if indentSize != "" {
		buf.WriteString("\n" + indent)
	}
	buf.WriteString("}")
	return buf.String(), nil
}
