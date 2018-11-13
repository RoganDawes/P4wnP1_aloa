package mgenetlink

import (
	nl "github.com/mame82/P4wnP1_go/mnetlink"
	"errors"
	"golang.org/x/sys/unix"
	"log"
)

var (
	ENlClient = errors.New("Netlink client not connected, maybe not created with NewGeNl")
)

func NewGeNl() (res *Client, err error) {
	nlcl,err := nl.NewNl(unix.NETLINK_GENERIC)
	if err != nil { return nil, err }
	res = &Client{
		nlclient: nlcl,
	}
	return
}

func (c *Client) Open() (err error) {
	if c.nlclient == nil { return ENlClient }
	return c.nlclient.Open()
}

func (c *Client) Close() (err error) {
	if c.nlclient == nil { return ENlClient }
	return c.nlclient.Close()
}

func (c *Client) AddGroupMembership(groupid uint32) (err error) {
	return c.nlclient.AddGroupMembership(int(groupid))
}

func (c *Client) DropGroupMembership(groupid uint32) (err error) {
	return c.nlclient.DropGroupMembership(int(groupid))
}

func (c *Client) Receive() (msgs []nl.Message, err error) {
	return c.nlclient.Receive()
}


// This refers to generic netlink families, not netlink families (netlink family is always NETLINK_GENERIC)
func (c *Client) GetFamilies() (families []Family, err error) {
	nlmsg := nl.Message{
		Flags: unix.NLM_F_REQUEST | unix.NLM_F_DUMP, // + dump request to get all
		Type: unix.GENL_ID_CTRL,
	}
	genlmsg := Message{
		Cmd: unix.CTRL_CMD_GETFAMILY,
		Version: 1, // control version 1 ??
	}

	genlmsg.Data = []byte{}
	if err != nil { return }

	nlmsg_data,err := genlmsg.MarshalBinary()
	if err != nil { return }
	nlmsg.SetData(nlmsg_data)



	err = c.nlclient.Send(nlmsg)
	if err != nil { return }

	// read answer (or sth else)
	msgs,err := c.nlclient.Receive()
	if err != nil { return }

	//fmt.Printf("Answer: %+v\n", msgs)

	for _,msg := range msgs {

		genl_msg := Message{} //genl message
		genl_msg.UnmarshalBinary(msg.GetData()) //parse NL message payload as genl message
		attrs,err := genl_msg.AttributesFromData() // parse genl_msg payload as attr array
		if err != nil {
			log.Println("Error parsing message data as attributes")
			continue
		}
		//fmt.Printf("Msg %d attributes: %+v\n", idx, attrs)
		curFamily,err := ParseAttrsToFamily(attrs) // parse attr array as data for a single family
		if err != nil {
			log.Println("Error parsing attributes as family")
			continue
		}
		families = append(families, curFamily)


	}
	return
}

func (c *Client) GetFamily(familyName string) (family *Family, err error) {
	nlmsg := nl.Message{
		Flags: unix.NLM_F_REQUEST,
		Type: unix.GENL_ID_CTRL,
		//Seq: seq,
		//Pid: pid,
	}
	genlmsg := Message{
		Cmd: unix.CTRL_CMD_GETFAMILY,
		Version: 1, // control version 1 ??
	}
	genlmsg_attr := nl.Attr{
		Type: unix.CTRL_ATTR_FAMILY_NAME,
	}
	genlmsg_attr.SetData(nl.Str2Bytes(familyName))

	genlmsg.Data,err = genlmsg_attr.MarshalBinary()
	if err != nil { return }

	nlmsg_data,err := genlmsg.MarshalBinary()
	if err != nil { return }
	nlmsg.SetData(nlmsg_data)



	err = c.nlclient.Send(nlmsg)
	if err != nil { return }

	// read answer (or sth else)
	msgs,err := c.nlclient.Receive()
	if err != nil { return }

	//fmt.Printf("Answer: %+v\n", msgs)

	for _,msg := range msgs {
		if family != nil {
			// multiple valid families in response
			return nil,errors.New("Multiple valid families in resposne to GETFAMILY command")
		}

		genl_msg := Message{} //genl message
		genl_msg.UnmarshalBinary(msg.GetData()) //parse NL message payload as genl message
		attrs,err := genl_msg.AttributesFromData() // parse genl_msg payload as attr array
		if err != nil {
			log.Println("Error parsing message data as attributes")
			continue
		}
		//fmt.Printf("Msg %d attributes: %+v\n", idx, attrs)
		curFamily,err := ParseAttrsToFamily(attrs) // parse attr array as data for a single family
		if err != nil {
			log.Println("Error parsing attributes as family")
			continue
		}
		family = &curFamily


	}
	return
}

type Client struct {
	nlclient *nl.Client
}
