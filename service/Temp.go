package service

type SubSys interface {
	GetState()
	DeploySettings()
	LoadSettings()
	StoreSettings()
	SetLogger()
	Start()
	Stop()
}
