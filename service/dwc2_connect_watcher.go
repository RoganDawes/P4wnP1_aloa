// +build arm

package service

import (
	"fmt"
	"errors"
	"github.com/mame82/P4wnP1_aloa/common_web"

	genl "github.com/mame82/P4wnP1_aloa/mgenetlink"
	nl "github.com/mame82/P4wnP1_aloa/mnetlink"
)

/*
Needs modified dwc2 kernel module, sending multicast generic netlink messages for genl family 'p4wnp1' on multicast group 'dwc2'

*/

const (
	fam_name = "p4wnp1"
	dwc2_group_name = "dwc2"

	// commands
	dwc2_cmd_connection_state = uint8(0)

	// attributes
	dwc2_attr_connection_dummy = uint16(0)
	dwc2_attr_connection_state = uint16(1)

)

var (
	EP4wnP1FamilyMissing = errors.New("Couldn't find generic netlink family for P4wnP1")
	EDwc2GrpMissing = errors.New("Couldn't find generic netlink mcast group for P4wnP1 dwc2")
	EDwc2GrpJoin = errors.New("Couldn't join generic netlink mcast group for P4wnP1 dwc2")
	EWrongFamily = errors.New("Message not from generic netlink family P4wnP1")
)


type Dwc2ConnectWatcher struct {
	rootSvc *Service

	genl *genl.Client
	fam *genl.Family


	isRunning bool
	connected bool
	firstUpdateDone bool
}


func (d * Dwc2ConnectWatcher) update(newStateConnected bool) {
	d.connected = newStateConnected

	// --> here a event could be triggered (in case the event manager is registered)
	if d.connected {
		fmt.Println("Connected to USB host")
		d.rootSvc.SubSysEvent.Emit(ConstructEventTrigger(common_web.TRIGGER_EVT_TYPE_USB_GADGET_CONNECTED))
		//d.rootSvc.SubSysEvent.Emit(ConstructEventLog("USB watcher", 1, "Connected to USB host"))

	} else {
		fmt.Println("Disconnected from USB host")
		d.rootSvc.SubSysEvent.Emit(ConstructEventTrigger(common_web.TRIGGER_EVT_TYPE_USB_GADGET_DISCONNECTED))
		//d.rootSvc.SubSysEvent.Emit(ConstructEventLog("USB watcher", 1, "Disconnected from USB host"))
	}
}

func (d * Dwc2ConnectWatcher) parseMsg(msg nl.Message) (cmd genl.Message, err error) {
	if msg.Type != d.fam.ID {
		// Multicast message from different familiy, ignore
		err = EWrongFamily
		return
	}

	err = cmd.UnmarshalBinary(msg.GetData())
	if err != nil { return }
	return
}


func (d * Dwc2ConnectWatcher) evt_loop() {
	d.isRunning = true
	// ToDo, make loop stoppable by non-blocking/interruptable socket read a.k.a select with timeout
	for d.isRunning {
		fmt.Println("\nWaiting for messages from P4wnP1 kernel mods...\n")
		msgs,errm := d.genl.Receive()
		if errm == nil {
			for _,msg := range msgs {
				if cmd,errp := d.parseMsg(msg); errp == nil {
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
								switch param.GetDataUint8() {
								case 0: //disconnected
									d.update(false)
								case 1: //connected
									d.update(true)
								}
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

func (d * Dwc2ConnectWatcher) IsConnected() bool {
	return d.connected
}


func (d * Dwc2ConnectWatcher) Start() (err error){
	d.genl,err = genl.NewGeNl() //genl client
	if err != nil { return err }

	err = d.genl.Open() //Connect to generic netlink
	if err != nil { return }

	// try to find GENL family for P4wnP1
	d.fam,err = d.genl.GetFamily(fam_name)
	if err != nil {
		d.genl.Close()
		return EP4wnP1FamilyMissing
	}

	// try to join group for dwc2
	grpId,err := d.fam.GetGroupByName(dwc2_group_name)
	if err != nil {
		d.genl.Close()
		return EDwc2GrpMissing
	}
	err = d.genl.AddGroupMembership(grpId)
	if err != nil {
		d.genl.Close()
		return EDwc2GrpMissing
	}




	d.isRunning = true
	go d.evt_loop()

	return nil
}

func (d * Dwc2ConnectWatcher) Stop() error {
	d.isRunning = false

	// leave dwc2 group
	if grpId,err := d.fam.GetGroupByName(dwc2_group_name); err == nil {
		d.genl.DropGroupMembership(grpId)
	}
	// close soket
	return d.genl.Close()

}

func NewDwc2ConnectWatcher(rootSvc *Service) (d *Dwc2ConnectWatcher) {

	d = &Dwc2ConnectWatcher{
		rootSvc: rootSvc,
	}
	return d
}

