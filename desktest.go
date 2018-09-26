package main

import (
	"fmt"
	"github.com/mame82/P4wnP1_go/common_web"
	pb "github.com/mame82/P4wnP1_go/proto"
	"github.com/mame82/P4wnP1_go/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*

#!/bin/bash
# /tmp/test.sh
echo "Bash DHCP Lease"
echo "Interface: $DHCP_LEASE_IFACE"
echo "Mac: $DHCP_LEASE_MAC"
echo "IP: $DHCP_LEASE_IP"


#!/bin/bash
# /tmp/test1.sh
echo "Bash SSH Login"
echo "Interface: $SSH_LOGIN_USER"

 */

func main() {
	pseudoService := &service.Service{
		SubSysEvent: service.NewEventManager(10),
	}
	tam := service.NewTriggerActionManager(pseudoService)

	pseudoService.SubSysEvent.Start()
	tam.Start()

	// create test trigger
	serviceUpRunScript := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_BashScript{
			BashScript: &pb.ActionStartBashScript{
				ScriptPath: "/usr/local/P4wnP1/scripts/servicestart1.sh",
			},
		},
	}
	dhcpLeaseScript := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_DhcpLeaseGranted{
			DhcpLeaseGranted: &pb.TriggerDHCPLeaseGranted{},
		},
		Action: &pb.TriggerAction_BashScript{
			BashScript: &pb.ActionStartBashScript{
				ScriptPath: "/tmp/test.sh",
			},
		},
	}
	sshLoginScript := &pb.TriggerAction{
		OneShot: true,
		Trigger: &pb.TriggerAction_SshLogin{
			SshLogin: &pb.TriggerSSHLogin{},
		},
		Action: &pb.TriggerAction_BashScript{
			BashScript: &pb.ActionStartBashScript{
				ScriptPath: "/tmp/test1.sh",
			},
		},
	}
	serviceUpLog := &pb.TriggerAction{
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(serviceUpRunScript)
	tam.AddTriggerAction(serviceUpLog)
	tam.AddTriggerAction(dhcpLeaseScript)
	tam.AddTriggerAction(sshLoginScript)

	/*
	// Pause TriggerActionManager after 5 seconds for 5 seconds
	go func() {
		time.Sleep(time.Second * 5)
		tam.Stop()
		time.Sleep(time.Second * 5)
		tam.Start()
	}()
	*/

	go func() {
		for {
			pseudoService.SubSysEvent.Emit(service.ConstructEventTriggerDHCPLease("wlan0", "aa:bb:cc:dd:ee:ff", "172.24.0.6"))
			time.Sleep(1800*time.Millisecond)
		}
	}()

	go func() {
		time.Sleep(time.Second)
		for  {
			pseudoService.SubSysEvent.Emit(service.ConstructEventTriggerSSHLogin("testuser"))
			time.Sleep(5*time.Second)
		}

	}()

	go func() {
		time.Sleep(2*time.Second)
		for  {
			pseudoService.SubSysEvent.Emit(service.ConstructEventTrigger(common_web.EVT_TRIGGER_TYPE_SERVICE_STARTED))
			time.Sleep(5*time.Second)
		}

	}()

	//use a channel to wait for SIGTERM or SIGINT
	fmt.Println("P4wnP1 service initialized, stop with SIGTERM or SIGINT")
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	si := <-sig
	log.Printf("Signal (%v) received, ending P4wnP1_service ...\n", si)
	pseudoService.SubSysEvent.Stop()
	tam.Stop()
	return

}
