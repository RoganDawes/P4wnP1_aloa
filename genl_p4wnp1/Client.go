package genl_p4wnp1

import (
	"fmt"
	genl "github.com/mame82/P4wnP1_go/mgenetlink"
	nl "github.com/mame82/P4wnP1_go/mnetlink"
	"errors"
)

const (
	fam_name = "p4wnp1"
	dwc2_group_name = "p4wnp1_dwc2_mc"
)

var (
	EP4wnP1FamilyMissing = errors.New("Couldn't find generic netlink family for P4wnP1")
	EDwc2GrpMissing = errors.New("Couldn't find generic netlink mcast group for P4wnP1 dwc2")
	EDwc2GrpJoin = errors.New("Couldn't join generic netlink mcast group for P4wnP1 dwc2")
	EWrongFamily = errors.New("Message not from generic netlink family P4wnP1")
)

type Client struct {
	genl *genl.Client
	fam *genl.Family
	running bool
}

func NewClient() (c *Client, err error) {
	res := &Client{}
	res.genl,err = genl.NewGeNl()
	if err != nil { return c,err }
	return res,nil
}

func (c *Client) Open() (err error) {
	err = c.genl.Open()
	if err != nil { return }

	// try to find GENL family for P4wnP1
	c.fam,err = c.genl.GetFamily(fam_name)
	if err != nil {
		c.genl.Close()
		return EP4wnP1FamilyMissing
	}

	// try to join group for dwc2
	grpId,err := c.fam.GetGroupByName(dwc2_group_name)
	if err != nil {
		c.genl.Close()
		return EDwc2GrpMissing
	}
	err = c.genl.AddGroupMembership(grpId)
	if err != nil {
		c.genl.Close()
		return EDwc2GrpMissing
	}

	go c.rcv_loop()
	return
}

func (c *Client) rcv_loop() {
	c.running = true
	// ToDo, make loop stoppable by non-blocking/interruptable socket read a.k.a select with timeout
	for c.running {
		fmt.Println("\nWaiting for messages from P4wnP1 kernel mods...\n")
		msgs,errm := c.genl.Receive()
		if errm == nil {
			for _,msg := range msgs {
				if cmd,errp := c.parseMsg(msg); errp == nil {
					switch cmd.Cmd {
					case dwc2_cmd_connection_state:
						fmt.Println("COMMAND_CONNECTION_STATE")
						params,perr := cmd.AttributesFromData()
						if perr != nil {
							fmt.Println("Couldn't parse params for COMMAND_CONNECTION_STATE")
							continue
						}
						// find
						for _,param := range params {
							if param.Type == dwc2_attr_connection_state {
								fmt.Println("Connection State: ", param.GetDataUint8())
							}
						}
					default:
						fmt.Printf("Unknown command:\n%+v\n", cmd)
					}
				} else {
					fmt.Printf("Message ignored:\n%+v\n", msg)
					continue
				}

			}
		} else {
			fmt.Println("Receive error: ", errm)
		}
	}

	fmt.Println("GenNl rcv loop ended")
}

const (
	dwc2_cmd_connection_state = uint8(0)

	dwc2_attr_connection_dummy = uint16(0)
	dwc2_attr_connection_state = uint16(1)


)

func (c *Client) parseMsg(msg nl.Message) (cmd genl.Message, err error) {
	if msg.Type != c.fam.ID {
		// Multicast message from different familiy, ignore
		err = EWrongFamily
		return
	}

	err = cmd.UnmarshalBinary(msg.GetData())
	if err != nil { return }
	return
}

func (c *Client) Close() (err error) {
	c.running = false

	// leave dwc2 group
	if grpId,err := c.fam.GetGroupByName(dwc2_group_name); err == nil {
		c.genl.DropGroupMembership(grpId)
	}
	// close soket
	return c.genl.Close()

}
