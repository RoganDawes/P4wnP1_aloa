package mnetlink

import (
	"encoding/hex"
	"errors"
	"golang.org/x/sys/unix"
)



var (
	EInvalidAttribute = errors.New("Invalid netlink attribute")
)

func AlignAttr(len int) int {
	return ((len) + unix.NLA_ALIGNTO - 1) & ^( unix.NLA_ALIGNTO - 1)
}

type Attr struct {
	Type uint16
	Len uint16
	data []byte
}

func (a *Attr) UnmarshalBinary(data []byte) (err error) {
	if len(data) < unix.NLA_HDRLEN {
		// less data than attribute header length
		err = EInvalidAttribute
		return
	}

	a.Len = hbo.Uint16(data[0:2])
	a.Type = hbo.Uint16(data[2:4])

	if AlignAttr(int(a.Len)) > len(data) {
		return EInvalidAttribute
	}

	a.data = make([]byte, a.Len - unix.NLA_HDRLEN)
	copy(a.data, data[unix.NLA_HDRLEN:])
	return
}

func (a *Attr) MarshalBinary() (data []byte, err error) {
	if a.Len < unix.NLA_HDRLEN {
		// less data than attribute header length
		err = EInvalidAttribute
		return
	}

	data = make([]byte, unix.NLA_HDRLEN + AlignAttr(len(a.data)))
	hbo.PutUint16(data[0:2], a.Len)
	hbo.PutUint16(data[2:4], a.Type)
	copy(data[unix.NLA_HDRLEN:], a.data)

	return
}



func (a *Attr) SetData(data []byte) (err error) {
	a.data = data
	a.Len = uint16(len(data) + unix.NLA_HDRLEN)
	return
}

func (a Attr) GetData() []byte {
	return a.data
}

func (a Attr) GetDataString() string {
	return Bytes2Str(a.data)
}

func (a Attr) GetDataUint32() uint32 {
	return hbo.Uint32(a.data[0:4])
}

func (a Attr) GetDataUint16() uint16 {
	return hbo.Uint16(a.data[0:2])
}

func (a Attr) GetDataUint8() uint8 {
	return uint8(a.data[0])
}

func (a Attr) GetDataDump() string {
	return hex.Dump(a.data)
}

func (a Attr) GetDataAttrs() (attrs []Attr, err error) {
	for offs := 0; len(a.data[offs:]) >= unix.NLA_HDRLEN;  {
		attr := Attr{}
		err = attr.UnmarshalBinary(a.data[offs:])
		if err != nil { return nil, err }
		attrs = append(attrs, attr)
		offs += AlignAttr(int(attr.Len))
	}

	return attrs, nil
}

