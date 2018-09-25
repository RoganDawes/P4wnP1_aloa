// +build linux

package service


var ServiceState *GlobalServiceState


type GlobalServiceState struct {
	BtSvc                 *BtService
}

func InitGlobalServiceState() (err error) {
	state := &GlobalServiceState{}
	ServiceState = state // store state in global variable


//	state.EvMgr = NewEventManager(20)

	state.BtSvc = NewBtService()
	return nil
}


func (state *GlobalServiceState) StartService() {
//	state.EvMgr.Start()
	// ToDo: Remove this, till the service is able to deploy startup settings
	state.BtSvc.StartNAP()
}

func (state *GlobalServiceState) StopService() {
//	state.EvMgr.Stop()
}
