package mgenetlink

import (
	"fmt"
	"github.com/mame82/P4wnP1_go/mnetlink"
	"golang.org/x/sys/unix"
)

type Op struct {
	Id uint32
	Flags uint32
}

func ParseAttrsToOps(attrs []mnetlink.Attr) (ops []Op, err error) {
	for _,attr := range attrs {
		//attr type == mcast group index
		// attr.data == []attr describing group
		op_attrs,err := attr.GetDataAttrs()
		if err != nil { return nil,err }
		op,err := ParseAttrsToOp(op_attrs)
		if err != nil { return nil,err }

		ops = append(ops, op)
	}

	return
}

func ParseAttrsToOp(attrs []mnetlink.Attr) (op Op, err error) {
	for _,attr := range attrs {
		switch attr.Type {
		case unix.CTRL_ATTR_OP_ID:
			op.Id = attr.GetDataUint32()
		case unix.CTRL_ATTR_OP_FLAGS:
			op.Flags = attr.GetDataUint32()
		default:
			fmt.Printf("Unknown Op attr %d: \n%v\n", attr.Type, attr.GetDataDump())

		}
	}

	return
}

