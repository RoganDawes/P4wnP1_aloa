package mgenetlink

import (
	"github.com/mame82/P4wnP1_go/mnetlink"
	"golang.org/x/sys/unix"
	"errors"
)

var (
	EInvalidMessage = errors.New("Invalid generic netlink message")
)


type Message struct {
	Cmd      uint8
	Version  uint8
	Reserved uint16
	Data []byte
}

func (m *Message) MarshalBinary() (data []byte, err error) {
	data = make([]byte, unix.GENL_HDRLEN + len(m.Data))

	data[0] = m.Cmd
	data[1] = m.Version
	copy(data[unix.GENL_HDRLEN:], m.Data)

	return
}

func (m *Message) UnmarshalBinary(src []byte) (err error) {
	if len(src) < unix.GENL_HDRLEN {
		return EInvalidMessage
	}

	m.Cmd = src[0]
	m.Version = src[1]

	m.Data = src[unix.GENL_HDRLEN:]
	return
}

func (m Message) AttributesFromData() (attrs []mnetlink.Attr, err error) {
	for offs := 0; len(m.Data[offs:]) >= unix.NLA_HDRLEN;  {
		attr := mnetlink.Attr{}
		err = attr.UnmarshalBinary(m.Data[offs:])
		if err != nil { return nil, err }
		attrs = append(attrs, attr)
		offs += mnetlink.AlignAttr(int(attr.Len))
	}

	return attrs, nil
}
