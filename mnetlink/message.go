package mnetlink

import (
	"errors"
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

var (
	EInvalidMessage = errors.New("Invalid netlink message")
)


func AlignMsg(len int) int {
	return ((len) + unix.NLMSG_ALIGNTO - 1) & ^( unix.NLMSG_ALIGNTO - 1)
}

type Message struct {
	Len   uint32
	Type  uint16
	Flags uint16
	Seq   uint32
	Pid   uint32
	data []byte
}

func (m *Message) MarshalBinary() (data []byte, err error) {
	msgLen := AlignMsg(int(m.Len))
	if msgLen < unix.NLMSG_HDRLEN {
		return nil, EInvalidMessage
	}

	data = make([]byte, msgLen)

	hbo.PutUint32(data[0:4], m.Len)
	hbo.PutUint16(data[4:6], uint16(m.Type))
	hbo.PutUint16(data[6:8], uint16(m.Flags))
	hbo.PutUint32(data[8:12], m.Seq)
	hbo.PutUint32(data[12:16], m.Pid)
	copy(data[unix.NLMSG_HDRLEN:], m.data)

	return
}

func (m *Message) UnmarshalBinary(src []byte) (err error) {
	m.Len = hbo.Uint32(src[0:4])
	// truncate message to length
	src = src[:m.Len]

	m.Type = hbo.Uint16(src[4:6])
	m.Flags = hbo.Uint16(src[6:8])
	m.Seq = hbo.Uint32(src[8:12])
	m.Pid = hbo.Uint32(src[12:16])
	m.data = src[16:]
	return
}

func (m *Message) SetData(data []byte) (err error) {
	m.data = data
	m.Len = uint32(len(data) + unix.NLMSG_HDRLEN)
	return
}

func (m Message) GetData() []byte {
	return m.data
}

func (m Message) HasFlagMulti() bool {
	return (m.Flags & unix.NLM_F_MULTI) > 0
}

func (m Message) HasFlagDump() bool {
	return (m.Flags & unix.NLM_F_DUMP) > 0
}

func (m Message) HasFlagAck() bool {
	return (m.Flags & unix.NLM_F_ACK) > 0
}

func (m Message) IsTypeDone() bool {
	return m.Type == unix.NLMSG_DONE
}

func (m Message) IsTypeError() bool {
	return m.Type == unix.NLMSG_ERROR
}

func (m Message) IsTypeNoop() bool {
	return m.Type == unix.NLMSG_NOOP
}

func (m Message) GetErrNo() error {
	//neg_err := int(hbo.Uint32(m.data[0:4]))
	neg_err := *(*int32)(unsafe.Pointer(&m.data[0]))

	return syscall.Errno(neg_err * -1)
}