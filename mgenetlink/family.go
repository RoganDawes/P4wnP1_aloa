package mgenetlink

import (
	"errors"
	"fmt"
	"github.com/mame82/P4wnP1_go/mnetlink"
	"golang.org/x/sys/unix"
)

var (
	EGroupNotFound = errors.New("Group not found")
)

type Family struct {
	ID      uint16
	Version uint8
	Name    string
	Groups  []McastGroup
	Ops []Op //ToDo: struct and parsing
}

func ParseAttrsToFamily(attrs []mnetlink.Attr) (family Family, err error) {
	for _,attr := range attrs {

		switch attr.Type {
		case unix.CTRL_ATTR_FAMILY_ID:
			//fmt.Printf("Family ID: %d\n", attr.GetDataUint16())
			family.ID = attr.GetDataUint16()
		case unix.CTRL_ATTR_FAMILY_NAME:
			//fmt.Printf("Family Name: %s\n", attr.GetDataString())
			family.Name = attr.GetDataString()
		case unix.CTRL_ATTR_VERSION:
			//fmt.Printf("Family Version: %d\n", attr.GetDataUint8())
			family.Version = attr.GetDataUint8()
		case unix.CTRL_ATTR_MCAST_GROUPS:
			mcast_attr,err := attr.GetDataAttrs()
			if err != nil { return family,err }
			//fmt.Printf("Family Mcast Groups: \n%+v\n", mcast_attr)
			family.Groups,err = ParseAttrsToMcastGroups(mcast_attr)
			if err != nil { return family,err }
		case unix.CTRL_ATTR_HDRSIZE:
			//fmt.Printf("ATTR HDRSIZE: \n%v\n", attr.GetDataDump())
		case unix.CTRL_ATTR_MAXATTR:
			//fmt.Printf("ATTR MAXATTR: \n%v\n", attr.GetDataDump())
		case unix.CTRL_ATTR_OPS:
			//fmt.Printf("Family Ops: \n%v\n", attr.GetDataDump())
			ops_attr,err := attr.GetDataAttrs()
			if err != nil { return family,err }
			//fmt.Printf("Family Mcast Groups: \n%+v\n", mcast_attr)
			family.Ops,err = ParseAttrsToOps(ops_attr)
			if err != nil { return family,err }


		default:
			fmt.Printf("Unknown attr %d: \n%v\n", attr.Type, attr.GetDataDump())
		}
	}

	return
}

func (fam *Family) GetGroupByName(name string) (id uint32, err error) {
	for _,grp := range fam.Groups {
		if grp.Name == name {
			return grp.Id,nil
		}
	}
	return id,EGroupNotFound
}