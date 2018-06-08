package hid

import (
	"github.com/robertkrimen/otto"
	"log"
	"time"
	"sync"
	"errors"
)


const (
	MAX_VM = 8
)

var halt = errors.New("Stahp")

type AsyncOttoVM struct {
	vm           *otto.Otto
	ResultErr    error
	ResultValue  otto.Value
	isWorking    bool
	runningAsync bool
	mutex        *sync.Mutex
	finCond      *sync.Cond
}

func NewAsyncOttoVM(vm *otto.Otto) *AsyncOttoVM {
	mutex := &sync.Mutex{}
	vm.Interrupt = make(chan func(), 1) //set Interrupt channel
	return &AsyncOttoVM{
		mutex:        mutex,
		isWorking:    false,
		runningAsync: false,
		ResultErr:    nil,
		ResultValue:  otto.Value{},
		vm:           vm,
		finCond:      sync.NewCond(mutex),
	}
}

func NewAsyncOttoVMClone(vm *otto.Otto) *AsyncOttoVM {
	return NewAsyncOttoVM(vm.Copy())
}

func (avm *AsyncOttoVM) Run(src interface{}) (val otto.Value, res error) {
	avm.setIsWorking(true)
	avm.runningAsync = false
	val,res= avm.vm.Run(src)
	avm.setIsWorking(false)
	return
}

func (avm *AsyncOttoVM) RunAsync(src interface{}) (error) {
	go func(avm *AsyncOttoVM) {
		avm.setIsWorking(true)
		avm.runningAsync = true

		defer func() {
			if caught := recover(); caught != nil {
				if caught == halt {
					avm.ResultErr =  errors.New("VM execution cancelled")
					avm.isWorking = false
					avm.setIsWorking(false)
					return
				}
				panic(caught) //re-raise panic, as it isn't `halt`
			}
			avm.setIsWorking(false)
		}()

		avm.ResultValue, avm.ResultErr = avm.vm.Run(src)

		//avm.setIsWorking(false)
	}(avm)
	return nil
}

func (avm *AsyncOttoVM) setIsWorking(isWorking bool) {
	avm.isWorking = isWorking
	avm.mutex.Lock()
	if !isWorking {
		avm.finCond.Broadcast()
	}
	avm.mutex.Unlock()

}

func (avm *AsyncOttoVM) IsWorking() (res bool) {
	res = avm.isWorking
	return
}

func (avm *AsyncOttoVM) WaitResult()  (val otto.Value, err error) {
	if !avm.runningAsync {
		return otto.Value{},errors.New("AsyncVM isn't running an async script from which a result could be received")
	}
	avm.mutex.Lock()
	avm.finCond.Wait()
	avm.mutex.Unlock()
	avm.runningAsync = false
	return avm.ResultValue, avm.ResultErr
}

func (avm *AsyncOttoVM) Cancel() error {
	if !avm.runningAsync {
		return errors.New("AsyncVM isn't running an async script which could be cancelled")
	}

	//interrupt vm
	avm.vm.Interrupt <- func() {
		panic(halt)
	}

	return nil
}

type HIDController struct {
	Keyboard *HIDKeyboard
	vmPool [MAX_VM]*AsyncOttoVM //ToDo: check if this could be changed to sync.Pool
	vmMaster *otto.Otto
}

func NewHIDController(keyboardDevicePath string, keyboardMapPath string, mouseDevicePath string) (ctl *HIDController, err error) {
	ctl = &HIDController{}
	//init keyboard
	ctl.Keyboard, err = NewKeyboard(keyboardDevicePath, keyboardMapPath)
	if err != nil { return nil, err	}

	//init master otto vm

	ctl.initMasterVM()

	//clone VM to pool
	for  i:=0; i< len(ctl.vmPool); i++ {
		ctl.vmPool[i] = NewAsyncOttoVMClone(ctl.vmMaster)
	}


	return
}

func (ctl *HIDController) NextUnusedVM() (idx int, vm *AsyncOttoVM, err error) {
	//iterate over pool
	for idx,avm := range ctl.vmPool {
		if !avm.IsWorking() {
			//return first non-working vm
			return idx, avm, nil //free to be used
		}
	}

	return 0, nil, errors.New("No free JavaScript VM available in pool")
}

func (ctl *HIDController) RunScript(script string) (val otto.Value, err error) {
	//fetch next free vm from pool
	_,avm,err := ctl.NextUnusedVM()
	if err != nil { return otto.Value{}, err }

	val, err = avm.Run(script)
	return
}

func (ctl *HIDController) StartScriptAsBackgroundJob(script string) (avmId int, avm *AsyncOttoVM, err error) {
	//fetch next free vm from pool
	avmId,avm,err = ctl.NextUnusedVM()
	if err != nil { return 0, nil, err }

	//try to run script async
	err = avm.RunAsync(script)
	if err != nil { return 0, nil, err }
	return
}

func (ctl *HIDController) CancelBackgroundJob(jobId int) (err error) {
	if jobId < 0 || jobId >= MAX_VM {
		return errors.New("Invalid Id for AsyncOttoVM")
	}
	return ctl.vmPool[jobId].Cancel()
}

func (ctl *HIDController) WaitBackgroundJobResult(avmId int) (otto.Value, error) {
	if avmId < 0 || avmId >= MAX_VM {
		return otto.Value{}, errors.New("Invalid Id for AsyncOttoVM")
	}
	return ctl.vmPool[avmId].WaitResult()
}

func (ctl *HIDController) GetRunningBackgroundJobs() (res []int) {
	res = make([]int,0)
	for i := 0; i< MAX_VM; i++ {
		if ctl.vmPool[i].IsWorking() {
			res = append(res, i)
		}
	}
	return
}


func (ctl *HIDController) currentlyWorkingVMs() (res []*AsyncOttoVM) {
	res = make([]*AsyncOttoVM,0)
	for i := 0; i< MAX_VM; i++ {
		if ctl.vmPool[i].IsWorking() {
			res = append(res, ctl.vmPool[i])
		}
	}
	return
}

func (ctl *HIDController) CancelAllBackgroundJobs() error {
	for i := 0; i< MAX_VM; i++ {
		if ctl.vmPool[i].IsWorking() {
			ctl.vmPool[i].Cancel()
		}
	}
	return nil
}

//Function declarations for master VM
//ToDo: Global mutex for VM callbacks (or better for atomar part of Keyboard.StringToPressKeySequence)
func (ctl *HIDController) jsType(call otto.FunctionCall) (res otto.Value) {
	arg0 := call.Argument(0)
	//fmt.Printf("JS type() called with: `%s` (%s)\n", arg0, arg0)

	if !arg0.IsString() {
		log.Printf("HIDScript type: Wrong argument, 'type' accepts a single argument of type string. Error location:  %v\n", call.CallerLocation())
		return
	}

	outStr,err := arg0.ToString()
	if err != nil {
		log.Printf("HIDScript type: couldn't convert '%s' to UTF-8 string\n", arg0)
		return
	}
	log.Printf("HIDScript type: Typing '%s ...' on HID keyboard device '%s'\n", outStr[:15], ctl.Keyboard.DevicePath)
	err = ctl.Keyboard.StringToPressKeySequence(outStr)
	if err != nil {
		log.Printf("HIDScript type: Couldn't type out `%s` on %v\n", outStr, ctl.Keyboard.DevicePath)
		return
	}
	return
}

func (ctl *HIDController) jsLayout(call otto.FunctionCall) (res otto.Value) {
	arg0 := call.Argument(0)
	//fmt.Printf("JS type() called with: `%s` (%s)\n", arg0, arg0)

	if !arg0.IsString() {
		log.Printf("HIDScript layout: Wrong argument, 'layout' accepts a single argument of type string. Error location:  %v\n", call.CallerLocation())
		return
	}

	layoutName,err := arg0.ToString()
	if err != nil {
		//shouldn't happen
		log.Printf("HIDScript layout: couldn't convert '%s' to string\n", arg0)
		return
	}

	log.Printf("HIDScript layout: Setting layout to '%s'\n", layoutName)
	err = ctl.Keyboard.SetActiveLanguageMap(layoutName)
	if err != nil {
		log.Printf("HIDScript layout: Couldn't set layout `%s`: %v\n", layoutName, err)
		return
	}
	return
}


func (ctl *HIDController) jsDelay(call otto.FunctionCall) (res otto.Value) {

	arg0 := call.Argument(0)
	//fmt.Printf("JS delay() called with: `%s` (%s)\n", arg0, arg0)

	if !arg0.IsNumber() {
		log.Printf("HIDScript delay: Wrong argument, delay accepts a single argument ot type number. Error location:  %v\n", call.CallerLocation())
		return
	}

	fDelay,err := arg0.ToFloat()
	if err != nil {
		log.Printf("HIDScript delay: Error couldn't convert `%v` to float\n", arg0)
		return
	}
	delay := int(fDelay)
	log.Printf("HIDScript delay: Sleeping `%v` milliseconds\n", delay)
	time.Sleep(time.Millisecond * time.Duration(int(delay)))

	return
}

//for pressing key combos
func (ctl *HIDController) jsPress(call otto.FunctionCall) (res otto.Value) {

	arg0 := call.Argument(0)
	//fmt.Printf("JS delay() called with: `%s` (%s)\n", arg0, arg0)

	if !arg0.IsString() {
		log.Printf("HIDScript press: Wrong argument for 'press'. 'press' accepts a single argument of type string.\n\tError location:  %v\n", call.CallerLocation())
		return
	}

	comboStr,err := arg0.ToString()
	if err != nil {
		log.Printf("HIDScript press: Error couldn't convert '%v' to string\n", arg0)
		return
	}
	log.Printf("HIDScript press: Pressing combo '%s'\n", comboStr)
	err = ctl.Keyboard.StringToPressKeyCombo(comboStr)
	if err != nil {
		log.Printf("HIDScript press: Error couldn't convert `%v` to string\n", arg0)
		oErr,vErr := otto.ToValue(err)
		if vErr == nil { return oErr}
		return
	}
	return
}

func (ctl *HIDController) jsWaitLED(call otto.FunctionCall) (res otto.Value) {
	//arg0 has to be of type number, representing an LED MASK
	//arg1 is optional and represents the timeout in seconds, in case it isn't present, we set timeout to a year (=infinite in our context ;-))
	arg0 := call.Argument(0)
	arg1 := call.Argument(1)
	log.Printf("HIDScript: Called WaitLED(%v, %v)\n", arg0, arg1)
	maskInt, err := arg0.ToInteger()
	if err != nil || !arg0.IsNumber() || !(maskInt >= 0 && maskInt <= MaskAny) {
		//We don't mention KANA and COMPOSE in the error message
		log.Printf("HIDScript WaitLED: First argument for `waitLED` has to be a bitmask representing LEDs (NUM | CAPS | SCROLL | ANY).\nError location:  %v\n", call.CallerLocation())
		return
	}

	mask := byte(maskInt)
	//fmt.Printf("Mask: %d\n", mask )

	timeout := time.Hour * 24 * 365
	switch {
	case arg1.IsUndefined():
		log.Printf("HIDScript WaitLED: No timeout given setting to a year\n")
	case arg1.IsNumber():
//		log.Printf("Timeout given: %v\n", arg1)
		timeoutInt, err := arg1.ToInteger()
		if err != nil || timeoutInt < 0 {
			log.Printf("HIDScript WaitLED: Second argument for `waitLED` is the timeout in seconds and has to be given as positive interger, but '%d' was given!\n", arg1)
			return
		}
		timeout = time.Duration(timeoutInt) * time.Second
	default:
		log.Printf("HIDScript WaitLED: Second argument for `waitLED` is the timeout in seconds and has to be given as interger or omitted for infinite timeout\n")
		return
	}

	changed,err := ctl.Keyboard.WaitLEDStateChange(mask, timeout)
	//fmt.Printf("Changed %+v\n", changed)

	res,_ = call.Otto.ToValue(struct{
		TIMEOUT bool
		NUM bool
		CAPS bool
		SCROLL bool
		COMPOSE bool
		KANA bool
	}{
		TIMEOUT: err == ErrTimeout,
		NUM: err == nil && changed.NumLock,
		CAPS: err == nil && changed.CapsLock,
		SCROLL: err == nil && changed.ScrollLock,
		COMPOSE: err == nil && changed.Compose,
		KANA: err == nil && changed.Kana,
	})
	return
}


func (ctl *HIDController) initMasterVM() (err error) {
	ctl.vmMaster = otto.New()
	err = ctl.vmMaster.Set("NUM", MaskNumLock)
	if err != nil { return err }
	err = ctl.vmMaster.Set("CAPS", MaskCapsLock)
	if err != nil { return err }
	err = ctl.vmMaster.Set("SCROLL", MaskScrollLock)
	if err != nil { return err }
	err = ctl.vmMaster.Set("COMPOSE", MaskCompose)
	if err != nil { return err }
	err = ctl.vmMaster.Set("KANA", MaskKana)
	if err != nil { return err }
	err = ctl.vmMaster.Set("ANY", MaskAny)
	if err != nil { return err }


	err = ctl.vmMaster.Set("type", ctl.jsType)
	if err != nil { return err }
	err = ctl.vmMaster.Set("delay", ctl.jsDelay)
	if err != nil { return err }
	err = ctl.vmMaster.Set("press", ctl.jsPress)
	if err != nil { return err }
	err = ctl.vmMaster.Set("waitLED", ctl.jsWaitLED)
	if err != nil { return err }
	err = ctl.vmMaster.Set("layout", ctl.jsLayout)
	if err != nil { return err }
	return nil
}
