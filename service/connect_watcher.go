package service

import (
	"fmt"
	"github.com/mame82/P4wnP1_go/common_web"
	"github.com/mame82/P4wnP1_go/service/dwc2"
)

/*
Needs modified dwc2 kernel module, sending multicast netlink messages to group 24

The initial connection state is undefined, the event is fired based on an IRQ of the gadget core (not in host mode)
which has the USBRST flag set if a host connection occurs or unset if a disconnection occurs.
As these IRQs are triggered, as soon as a new gadget is deployed, the state wouldn't stay undefined, once the USB
stack has initialized the gadget settings.

The mentioned IRQ fires multiple times:
On disconnect: two times 0x00000000
On connect: 0x00000000 followed by two times 0x00001000

To cope with that behavior, we only fire an event if the new state differs from the last one
*/

const (
	mcast_group = 24
	host_connected = byte(0x01)
	host_disconnected = byte(0x00)
)

type Dwc2ConnectWatcher struct {
	rootSvc *Service

	isRunning bool
	connected bool
	nl *dwc2.Dwc2Netlink
	firstUpdateDone bool
}

func (d * Dwc2ConnectWatcher) udateNeeded(newStateConnected bool) (needed bool) {
	if !d.firstUpdateDone {
		d.firstUpdateDone = true
		return true
	}
	return newStateConnected != d.connected
}

func (d * Dwc2ConnectWatcher) update(newStateConnected bool) {
	d.connected = newStateConnected

	// --> here a event could be triggered (in case the event manager is registered)
	if d.connected {
		fmt.Println("Connected to USB host")
		d.rootSvc.SubSysEvent.Emit(ConstructEventTrigger(common_web.EVT_TRIGGER_TYPE_USB_GADGET_CONNECTED))
		d.rootSvc.SubSysEvent.Emit(ConstructEventLog("USB watcher", 1, "Connected to USB host"))

	} else {
		fmt.Println("Disconnected from USB host")
		d.rootSvc.SubSysEvent.Emit(ConstructEventTrigger(common_web.EVT_TRIGGER_TYPE_USB_GADGET_DISCONNECTED))
		d.rootSvc.SubSysEvent.Emit(ConstructEventLog("USB watcher", 1, "Disconnected from USB host"))
	}
}


func (d * Dwc2ConnectWatcher) evt_loop() {
	for d.isRunning {
		indata := d.nl.Read()
		if len(indata) != 1 {
			continue // ignore, we want tor receive an uint32
		}
		val := indata[0]
		switch val {
		case host_connected:
				d.update(true)
		case host_disconnected:
				d.update(false)
		default:
			fmt.Println("Unknown value from DWC2: ", val)
		}
	}
}

func (d * Dwc2ConnectWatcher) IsConnected() bool {
	return d.connected
}


func (d * Dwc2ConnectWatcher) Start() {
	d.nl.OpenNlKernelSock()
	d.isRunning = true
	go d.evt_loop()
}

func (d * Dwc2ConnectWatcher) Stop() {
	d.isRunning = false
	d.nl.Close()
}

func NewDwc2ConnectWatcher(rootSvc *Service) (d *Dwc2ConnectWatcher) {
	d = &Dwc2ConnectWatcher{
		nl: dwc2.NewDwc2Nl(24),
		rootSvc: rootSvc,
	}
	return d
}

