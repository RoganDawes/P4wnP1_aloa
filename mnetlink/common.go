package mnetlink

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

var hbo = HostByteOrder()

func Hbo() binary.ByteOrder {
	return hbo
}

// Detect host byteorder (used by netlink)
func HostByteOrder() (res binary.ByteOrder) {
	i := int(0x0100)
	ptr := unsafe.Pointer(&i)
	if 0x01 == *(*byte)(ptr) {
		return binary.BigEndian
	}
	return binary.LittleEndian
}


func Str2Bytes(s string) []byte {
	return append([]byte(s), 0x00)
}

func Bytes2Str(b []byte) string {
	return string(bytes.TrimSuffix(b, []byte{0x00}))
}
