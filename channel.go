package vesper

import (
	"fmt"
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
