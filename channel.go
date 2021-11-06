package vesper

import (
	"fmt"
	"time"
)

// this type is similar to what an extension type (outside the vesper package) would look like:
// the Value field of Object stores a pointer to the types specific data

// ChannelType - the type of Vesper's channel object
var ChannelType = defaultVM.Intern("<channel>")

type channel struct {
	name    string
	bufsize int
	channel chan *Object
}

func (ch *channel) String() string {
	s := "#[channel"
	if ch.name != "" {
		s += " " + ch.name
	}
	if ch.bufsize > 0 {
		s += fmt.Sprintf(" [%d]", ch.bufsize)
	}
	if ch.channel == nil {
		s += " CLOSED"
	}
	return s + "]"
}

// Channel - create a new channel with the given buffer size and name
func Channel(bufsize int, name string) *Object {
	return NewObject(ChannelType, &channel{name: name, bufsize: bufsize, channel: make(chan *Object, bufsize)})
}

// ChannelValue - return the Go channel object for the Vesper channel
func ChannelValue(obj *Object) chan *Object {
	if obj.Value == nil {
		return nil
	}
	v, _ := obj.Value.(*channel)
	return v.channel
}

// CloseChannel - close the channel object
func CloseChannel(obj *Object) {
	v, _ := obj.Value.(*channel)
	if v != nil {
		c := v.channel
		if c != nil {
			v.channel = nil
			close(c)
		}
	}
}

// VM Primitives

func vesperChannel(argv []*Object) (*Object, error) {
	name := argv[0].text
	bufsize := int(argv[1].fval)
	return Channel(bufsize, name), nil
}

func vesperClose(argv []*Object) (*Object, error) {
	switch argv[0].Type {
	case ChannelType:
		CloseChannel(argv[0])
	default:
		return nil, Error(ArgumentErrorKey, "close expected a channel")
	}
	return Null, nil
}

func vesperSend(argv []*Object) (*Object, error) {
	ch := ChannelValue(argv[0])
	if ch != nil {
		val := argv[1]
		timeout := argv[2].fval
		if NumberEqual(timeout, 0.0) {
			select {
			case ch <- val:
				return True, nil
			default:
			}
		} else if timeout > 0 {
			dur := time.Millisecond * time.Duration(timeout*1000.0)
			select {
			case ch <- val:
				return True, nil
			case <-time.After(dur):
			}
		} else {
			ch <- val
			return True, nil
		}
	}
	return False, nil
}

func vesperReceive(argv []*Object) (*Object, error) {
	ch := ChannelValue(argv[0])
	if ch != nil {
		timeout := argv[1].fval
		if NumberEqual(timeout, 0.0) {
			select {
			case val, ok := <-ch:
				if ok && val != nil {
					return val, nil
				}
			default:
			}
		} else if timeout > 0 {
			dur := time.Millisecond * time.Duration(timeout*1000.0)
			select {
			case val, ok := <-ch:
				if ok && val != nil {
					return val, nil
				}
			case <-time.After(dur):
			}
		} else {
			val := <-ch
			if val != nil {
				return val, nil
			}
		}
	}
	return Null, nil
}

func initChannelFunctions(vm *VM) {
	vm.DefineFunctionKeyArgs("channel", vesperChannel, ChannelType, []*Object{StringType, NumberType}, []*Object{EmptyString, Zero}, []*Object{vm.Intern("name:"), vm.Intern("bufsize:")})
	vm.DefineFunctionOptionalArgs("send", vesperSend, NullType, []*Object{ChannelType, AnyType, NumberType}, MinusOne)
	vm.DefineFunctionOptionalArgs("recv", vesperReceive, AnyType, []*Object{ChannelType, NumberType}, MinusOne)
	vm.DefineFunction("close", vesperClose, NullType, AnyType)
}
