// +build linux,arm

package service

import (
	"context"
	"fmt"
	"github.com/mame82/P4wnP1_aloa/common"
	"github.com/mame82/P4wnP1_aloa/common_web"
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"github.com/mame82/P4wnP1_aloa/service/datastore"
	"log"
	"syscall"
	"time"
)


func RegisterDefaultTriggerActions(tam *TriggerActionManager) {
	// create test trigger

	// Trigger to run startup script
	serviceUpRunScript := &pb.TriggerAction{
		IsActive: true,
		Immutable: true,
		OneShot: false,
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_BashScript{
			BashScript: &pb.ActionStartBashScript{
				ScriptName: "servicestart.sh",
			},
		},
	}
	tam.AddTriggerAction(serviceUpRunScript)

	/*
	logServiceStart := &pb.TriggerAction{
		IsActive: true,
		Trigger: &pb.TriggerAction_ServiceStarted{
			ServiceStarted: &pb.TriggerServiceStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(logServiceStart)

	logDHCPLease := &pb.TriggerAction{
		IsActive: true,
		Trigger: &pb.TriggerAction_DhcpLeaseGranted{
			DhcpLeaseGranted: &pb.TriggerDHCPLeaseGranted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(logDHCPLease)

	logUSBGadgetConnected := &pb.TriggerAction{
		IsActive: true,
		Trigger: &pb.TriggerAction_UsbGadgetConnected{
			UsbGadgetConnected: &pb.TriggerUSBGadgetConnected{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(logUSBGadgetConnected)

	logUSBGadgetDisconnected := &pb.TriggerAction{
		IsActive: true,
		Trigger: &pb.TriggerAction_UsbGadgetDisconnected{
			UsbGadgetDisconnected: &pb.TriggerUSBGadgetDisconnected{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(logUSBGadgetDisconnected)

	logWifiAp := &pb.TriggerAction{
		IsActive: true,
		Trigger: &pb.TriggerAction_WifiAPStarted{
			WifiAPStarted: &pb.TriggerWifiAPStarted{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(logWifiAp)

	logWifiSta := &pb.TriggerAction{
		IsActive: true,
		Trigger: &pb.TriggerAction_WifiConnectedAsSta{
			WifiConnectedAsSta: &pb.TriggerWifiConnectedAsSta{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(logWifiSta)

	logSSHLogin := &pb.TriggerAction{
		IsActive: true,
		Trigger: &pb.TriggerAction_SshLogin{
			SshLogin: &pb.TriggerSSHLogin{},
		},
		Action: &pb.TriggerAction_Log{
			Log: &pb.ActionLog{},
		},
	}
	tam.AddTriggerAction(logSSHLogin)
	*/
}

type Service struct {
	SubSysDataStore *datastore.Store // very first service
	//	SubSysState          interface{}
	//	SubSysLogging        interface{}
	SubSysNetwork *NetworkManager

	SubSysEvent          *EventManager
	SubSysUSB            *UsbGadgetManager
	SubSysLed            *LedService
	SubSysWifi           *WiFiService
	SubSysBluetooth      *BtService
	SubSysRPC            *server
	SubSysTriggerActions *TriggerActionManager
	SubSysGpio *GpioManager

	SubSysDwc2ConnectWatcher *Dwc2ConnectWatcher

	Ctx context.Context
	Cancel context.CancelFunc
	rebootOnStop bool
	shutdownOnStop bool
}

func NewService() (svc *Service, err error) {
	svc = &Service{}
	svc.Ctx,svc.Cancel = context.WithCancel(context.Background())

	svc.SubSysDataStore, err = datastore.Open(common.PATH_DATA_STORE, common.PATH_DATA_STORE_BACKUP + "/init.db")
	if err != nil {
		return nil, err
	}

	svc.SubSysEvent = NewEventManager(20)

	svc.SubSysLed = NewLedService()
	svc.SubSysNetwork, err = NewNetworkManager(svc) //Depends on EvenSubSys
	if err != nil {
		return nil, err
	}
	svc.SubSysUSB, err = NewUSBGadgetManager(svc) //Depends on NetworkSubSys, EvenSubSys
	//	if err == ErrUsbNotUsable { err = nil } //ToDo: delete this
	if err != nil {
		return nil, err
	}

	svc.SubSysWifi = NewWifiService(svc) //Depends on NetworkSubSys

	svc.SubSysGpio = NewGpioManager(svc) //Depends on event subsys

	svc.SubSysTriggerActions = NewTriggerActionManager(svc) //Depends on EventManager, UsbGadgetManager (to trigger HID scripts) and GpioManager

	svc.SubSysDwc2ConnectWatcher = NewDwc2ConnectWatcher(svc) // Depends on EventManager, should be started before USB gadget settings are deployed (to avoid missing initial state change)

	svc.SubSysBluetooth = NewBtService(svc, time.Second * 120) //Depends on NetworkSubSys (try to bring up bluetooth for up to 120s in background)



	svc.SubSysRPC = NewRpcServerService(svc) //Depends on all other
	return
}

func (s *Service) Start() (context.Context, context.CancelFunc) {
	log.Println("Starting service ...")

	s.SubSysEvent.Start()
	s.SubSysDwc2ConnectWatcher.Start()
	s.SubSysGpio.Start()
	s.SubSysLed.Start()
	s.SubSysRPC.StartRpcServerAndWeb("0.0.0.0", "50051", "8000", common.PATH_WEBROOT) //start gRPC service
	log.Println("Starting TriggerAction event listener ...")
	s.SubSysTriggerActions.Start()

	// Register TriggerActions
	/*
	log.Println("Register default TriggerActions ...")
	RegisterDefaultTriggerActions(s.SubSysTriggerActions)
	*/


	scriptFallback := false
	//retrieve Startup MasterTemplate name from store
	msgTemplateName := &pb.StringMessage{}
	errTemplateName := s.SubSysDataStore.Get(cSTORE_STARTUP_MASTER_TEMPLATE, msgTemplateName)
	if errTemplateName == nil {
		startupTemplate := msgTemplateName.Msg
		fmt.Printf("Loading MasterTemplate '%s' for startup ...\n", startupTemplate)

		// Deploy MasterTemplate
		_,errDeployStartupTemplate := s.SubSysRPC.DeployStoredMasterTemplate(context.Background(), &pb.StringMessage{Msg:startupTemplate})
		if errDeployStartupTemplate != nil {
			fmt.Printf("... error deploying Startup MasterTemplate '%s': %v\n", startupTemplate, errDeployStartupTemplate)
			scriptFallback = true
		}
	} else {
		fmt.Println("... error retrieving name for Startup MasterTemplate")
		scriptFallback = true
	}

	if scriptFallback {
		fmt.Println("... Fallback: Deploying TriggerAction for script based startup with 'servicestart.sh'")
		RegisterDefaultTriggerActions(s.SubSysTriggerActions)
	}


	// fire service started Event
	log.Println("Fire service started event ...")
	s.SubSysEvent.Emit(ConstructEventTrigger(common_web.TRIGGER_EVT_TYPE_SERVICE_STARTED))

	return s.Ctx, s.Cancel
}

func (s *Service) Reboot() {
	s.rebootOnStop = true
	s.Cancel()
}

func (s *Service) Shutdown() {
	s.shutdownOnStop = true
	s.Cancel()
}

func (s *Service) Stop() {
	s.SubSysTriggerActions.Stop()
	s.SubSysLed.Stop()
	s.SubSysGpio.Stop()
	s.SubSysBluetooth.Stop()
	s.SubSysDwc2ConnectWatcher.Stop()
	s.SubSysEvent.Stop()

	if s.rebootOnStop {
		fmt.Println("Rebooting...")
		syscall.Sync()
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	}

	if s.shutdownOnStop {
		fmt.Println("Shutdown...")
		syscall.Sync()
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	}
}
