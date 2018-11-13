package mgenetlink

import (
	"fmt"
	"github.com/mame82/P4wnP1_go/mnetlink"
	"errors"
	"golang.org/x/sys/unix"
)

var (
	EParseGroup = errors.New("Error parsing multicast group data")
)

type McastGroup struct {
	Id   uint32
	Name string
}



func ParseAttrsToMcastGroups(attrs []mnetlink.Attr) (groups []McastGroup, err error) {
	for _,attr := range attrs {
		//attr type == mcast group index
		// attr.data == []attr describing group
		grp_attrs,err := attr.GetDataAttrs()
		if err != nil { return nil,err }
		mcats_grp,err := ParseAttrsToMcastGroup(grp_attrs)
		if err != nil { return nil,err }

		groups = append(groups, mcats_grp)
	}

	return
}

func ParseAttrsToMcastGroup(attrs []mnetlink.Attr) (group McastGroup, err error) {
	for _,attr := range attrs {
		switch attr.Type {
		case unix.CTRL_ATTR_MCAST_GRP_ID:
			//fmt.Printf("Multicast Group ID: %d\n", attr.GetDataUint32())
			group.Id = attr.GetDataUint32()
		case unix.CTRL_ATTR_MCAST_GRP_NAME:
			//fmt.Printf("Multicast Group Name: %s\n", attr.GetDataString())
			group.Name = attr.GetDataString()
		case unix.CTRL_ATTR_MCAST_GRP_UNSPEC:
			//fmt.Printf("Multicast Group Unspec: \n%v\n", attr.GetDataDump())
			return group,EParseGroup
		default:
			fmt.Printf("Unknown McastGrp attr %d: \n%v\n", attr.Type, attr.GetDataDump())
			return group,EParseGroup
		}
	}

	return
}
