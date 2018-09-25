package service

import "github.com/mame82/P4wnP1_go/service/datastore"

const (
	// ToDo: change to non-temporary folder to persist over reboot
	pPATH_DATA_STORE = "/tmp/store"
)

type Service struct {
	SubSysDataStore      *datastore.Store // very first service
	SubSysState          interface{}
	SubSysLogging        interface{}
	SubSysNetwork *NetworkManager

	SubSysEvent          *EventManager
	SubSysUSB            *UsbGadgetManager
	SubSysLed            *LedService
	SubSysWifi           *WiFiService
	SubSysBluetooth      *BtService
	SubSysRPC            *server
	SubSysTriggerActions interface{}
}

func NewService() (svc *Service, err error) {
	svc = &Service{}

	svc.SubSysDataStore,err = datastore.Open(pPATH_DATA_STORE)
	if err != nil { return nil,err}

	svc.SubSysEvent = NewEventManager(20)

	svc.SubSysLed = NewLedService()
	svc.SubSysNetwork, err = NewNetworkManager()
	if err != nil { return nil,err}
	svc.SubSysUSB,err = NewUSBGadgetManager(svc) //Depends on NetworkSubSys, EvenSubSys
	if err == ErrUsbNotUsable { err = nil } //ToDo: delete this

	if err != nil { return nil,err}
	svc.SubSysWifi = NewWifiService(svc) //Depends on NetworkSubSys



	svc.SubSysRPC = NewRpcServerService(svc)  //Depends on all other
	return
}

func (s *Service) Start() {
	s.SubSysEvent.Start()
	s.SubSysLed.Start()
	s.SubSysRPC.StartRpcServerAndWeb("0.0.0.0", "50051", "8000", "/usr/local/P4wnP1/www") //start gRPC service
}

func (s *Service) Stop() {
	s.SubSysLed.Stop()

	s.SubSysEvent.Stop()
}
