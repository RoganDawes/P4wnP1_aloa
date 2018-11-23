// +build linux

package hid

import (
	"errors"
	"github.com/robertkrimen/otto"
	"fmt"
	"context"
	"sync"
	"log"
	"time"
)


const (
	MAX_VM = 8

)

var (
	globalJobList = make(map[*AsyncOttoJob]bool)
	globalJobListMutex = sync.Mutex{}

	hidControllerReuse *HIDController

	haltirq       = errors.New("Stahp")
	ErrAbort      = errors.New("Event listening aborted")
	ErrIrq        = errors.New("Aborted due to interrupt request")
	ErrNotAllowed = errors.New("Calling not allowed (currently disabled)")
	ErrNoKeyboard = errors.New("Using HID keyboard not allowed (currently disabled)")
	ErrNoMouse = errors.New("Using HID mouse not allowed (currently disabled)")
)

// Note: The HID Controller uses multiple otto.Otto VMs
// The HIDController object can't be used, once the USB stack re-initializes, as the HID keyboard and mouse
// depend on the device file currently in use. Thus I started multiple attempts to cleanly destroy the otto VMs
// in order to allow creating a new HIDController object. This didn't succeed, as no matter what I tried, the
// (no more used and referenced) otto VMs leak memory. To cope with that, the max amount of concurrent existing
// otto VMs is limited to MAX_VM. The VMs (clone from a master VM, providing all the functionality) have to be
// reused. This again means, the HIDController object exists during the whole runtime and needs a method to be
// reconfigured. The cause of a reconfiguration is a change in the HID functions, leading to the need of using
// different device files for keyboard / mouse communication (f.e. a setup with keyboard + mouse has /dev/hidg0
// and /dev/hidg1, while a changed setup with mouse or keyboard only use /dev/hidg0 for the respective device -
// there wouldn't exist a /dev/hidg1).
// In order to reuse the HIDController, there has to be some kind of Reset() method, fulfilling the following tasks:
// 	- init a new HIDKeyboard if necessary, based on the new devicepath
// 		- the HIDkeyboard uses a KeyboardLEDStateWatcher, which constantly reads the device file (receives USB HID
// 		output reports) and thus needs to be stopped and reinitialized
//	- init a new Mouse if necessary, there should be no problem overwriting the former Mouse object and let the GC
// 	do its work
//	- interrupt all script currently running on the ottoVMs



type HIDController struct {
	Keyboard *HIDKeyboard
	Mouse *Mouse
	vmPool [MAX_VM]*AsyncOttoVM
	vmMaster *otto.Otto
	ctx context.Context
	cancel context.CancelFunc
	eventHandler EventHandler

	abortLock *sync.Mutex
	isAborted bool
}

func NewHIDController(ctx context.Context, keyboardDevicePath string, keyboardMapPath string, mouseDevicePath string) (ctl *HIDController, err error) {
	//ToDo: check if hidcontroller could work with a context with cancellation, which then could be cancelled on reuse of the object
	if hidControllerReuse == nil {
		hidControllerReuse = &HIDController{}
		hidControllerReuse.abortLock = &sync.Mutex{}
		hidControllerReuse.isAborted = true

		hidControllerReuse.initMasterVM()
		//clone VM to pool
		for  i:=0; i< len(hidControllerReuse.vmPool); i++ {
			hidControllerReuse.vmPool[i] = NewAsyncOttoVMClone(hidControllerReuse.vmMaster)
		}
	} else {
		// an old HIDController object is reused and has to be cleaned

		hidControllerReuse.Abort()
		hidControllerReuse.Keyboard = nil
		hidControllerReuse.Mouse = nil
	}


	hidControllerReuse.abortLock.Lock()
	defer hidControllerReuse.abortLock.Unlock()

	hidControllerReuse.isAborted = false

	ctx,cancel := context.WithCancel(ctx)
	hidControllerReuse.ctx = ctx
	hidControllerReuse.cancel = cancel

	//Note: to disable mouse/keyboard support, the respective device path has to have zero length

	//init keyboard
	if len(keyboardDevicePath) > 0 {
		hidControllerReuse.Keyboard, err = NewKeyboard(ctx, keyboardDevicePath, keyboardMapPath)
		if err != nil { return nil, err	}
	}

	//init Mouse
	if len(mouseDevicePath) > 0 {
		hidControllerReuse.Mouse,err = NewMouse(mouseDevicePath)
		if err != nil { return nil, err	}
	}

	hidControllerReuse.SetDefaultHandler()
	return hidControllerReuse, nil
}

func (ctl *HIDController) SetEventHandler(handler EventHandler) {
	ctl.eventHandler = handler
	// set same handler for all child VMs
	for _,vm := range ctl.vmPool { vm.SetEventHandler(handler) }
}

func  (ctl *HIDController) HandleEvent(event Event) {
	fmt.Printf("!!! HID Controller Event: %+v\n", event)
}

func (ctl *HIDController) SetDefaultHandler() {
	ctl.SetEventHandler(ctl)
}

func (ctl *HIDController) emitEvent(event Event) {
	if ctl.eventHandler == nil { return }
	ctl.eventHandler.HandleEvent(event)

}

func (ctl *HIDController) Abort()  {
	ctl.abortLock.Lock()
	defer ctl.abortLock.Unlock()

	if ctl.isAborted {
		return
	}
	ctl.isAborted = true

	// stop the old LED reader if present
	// !! Keyboard.Close() has to be called before sending IRQ to VMs (there're JS function which register
	// blocking callbacks to the Keyboard's LED reader, which would hinder the VM interrupt in triggering)
	if hidControllerReuse.Keyboard != nil {
		hidControllerReuse.Keyboard.Close() //interrupts go routines reading from device file and lets LEDStateListeners die
	}

	if hidControllerReuse.Mouse != nil {
		hidControllerReuse.Mouse.Close()
	}

	// Interrupt all VMs already running
	//hidControllerReuse.CancelAllBackgroundJobs()
	hidControllerReuse.CancelAllBackgroundJobs()

	ctl.emitEvent(Event{
		Type:    EventType_CONTROLLER_ABORTED,
		Message: "Called abort on HID Controller",
		Job:     nil,
		Vm:      nil,
	})
}

func (ctl *HIDController) NextUnusedVM() (vm *AsyncOttoVM, err error) {
	//iterate over pool
	for _,avm := range ctl.vmPool {
		if !avm.IsWorking() {
			//return first non-working vm
			return avm, nil //free to be used
		}
	}

	return nil, errors.New("No free JavaScript VM available in pool")
}

func (ctl *HIDController) RunScript(ctx context.Context, script string, anonymousSelfInvoked bool) (val otto.Value, err error) {
	/*
	//fetch next free vm from pool
	avm,err := ctl.NextUnusedVM()
	if err != nil { return otto.Value{}, err }

	val, err = avm.Run(ctx, script)
	*/

	// use backround job and wait, to force keeping track of job in joblist, in case the result is never fetched due to
	// remote CLI abort after running endless script (wouldn't cancel the script)
	job,err := ctl.StartScriptAsBackgroundJob(ctx, script, anonymousSelfInvoked)
	if err != nil { return val,err }
	val,err = ctl.WaitBackgroundJobResult(ctx, job)
	return
}

func (ctl *HIDController) GetBackgroundJobByID(id int) (job *AsyncOttoJob, err error) {
	globalJobListMutex.Lock()
	for j,_ := range globalJobList {
		if j.Id == id {
			job = j
			break
		}
	}
	globalJobListMutex.Unlock()

	if job == nil {
		return nil, errors.New(fmt.Sprintf("Job with ID %d not found in list of background jobs\n", id))
	} else {
		return job,nil
	}
}


func (ctl *HIDController) retrieveJobFromOtto(vm *otto.Otto) (job *AsyncOttoJob, runVM *AsyncOttoVM, err error) {
	found := false
	globalJobListMutex.Lock()
	for cJob,_ := range globalJobList {
		vme := cJob.executingVM
		if vme == nil { continue }
		if vme.vm == vm {
			found = true
			job = cJob
			runVM = vme
		}

	}
	globalJobListMutex.Unlock()
	if found {
		return
	} else {
		return nil,nil,errors.New("Couldn't retrieve job of this Otto VM")
	}
}

func (ctl *HIDController) StartScriptAsBackgroundJob(ctx context.Context,script string, anonymousSelfInvoked bool) (job *AsyncOttoJob, err error) {
	//fetch next free vm from pool
	avm,err := ctl.NextUnusedVM()
	if err != nil {
		ctl.emitEvent(Event{
			Job:job,
			Message:"No free Java VM available to run the script",
			Type:EventType_JOB_NO_FREE_VM,
		})
		return nil, err
	}

	//try to run script async
	job,err = avm.RunAsync(ctx,script,anonymousSelfInvoked)
	if err != nil { return nil, err }

	ctl.emitEvent(Event{
		Job:job,
		Vm:avm,
		Type:EventType_JOB_STARTED,
		Message:script, //pack script source into message

		//ScriptSource:script,
	})

	//add job to global list
	globalJobListMutex.Lock()
	globalJobList[job] = true
	globalJobListMutex.Unlock()

	//ad go routine which monitors context of the job, to remove it from global list, once done or interrupted
	go func() {
		//fmt.Printf("StartScriptAsBackgroundJob: started finish watcher for job %d\n",job.Id)
		select {
		case <- job.ctx.Done():
			fmt.Printf("StartScriptAsBackgroundJob: Job %d finished or interrupted, removing from global list\n",job.Id)

			ctl.emitEvent(Event{
				Job:job,
				Vm:avm,
				Type:EventType_JOB_STOPPED,
				Message:"Script stopped",
				//ScriptSource:script,
			})

			//remove job from global list after result has been retrieved
			globalJobListMutex.Lock()
			delete(globalJobList,job)
			globalJobListMutex.Unlock()

		}
	}()

	return
}

func (ctl *HIDController) WaitBackgroundJobResult(ctx context.Context, job *AsyncOttoJob) (val otto.Value, err error) {
	globalJobListMutex.Lock()
	if !globalJobList[job] {
		err = errors.New(fmt.Sprintf("Tried to retrieve results of job with id %d failed, because it is not in the list of background jobs", job.Id))
	}
	globalJobListMutex.Unlock()
	if err != nil {return}

	//cancel job when provided context is done
	// ToDo: do profiling to check if this go routine stays open, in case ctx is never done
	go func(ctx context.Context, job *AsyncOttoJob) {
		select {
		case <- ctx.Done():
			job.Cancel()
		}
	}(ctx,job)

	val,err = job.WaitResult() //Blocking result wait

	return
}


func (ctl *HIDController) CancelAllBackgroundJobs() {
	globalJobListMutex.Lock()
	oldList := globalJobList
	// cancel known jobs
	for job,_ := range oldList {
		fmt.Printf("Cancelling Job %d\n", job.Id)
		job.Cancel()
		/*
		ctl.emitEvent(Event{
			Job:job,
			Vm:job.executingVM,
			Type:EventType_JOB_CANCELLED,
			Message:"Script execution cancelled",
		})
		*/
	}
	globalJobList = make(map[*AsyncOttoJob]bool) //Create new empty list
	globalJobListMutex.Unlock()

}


func (ctl *HIDController) GetAllBackgroundJobs() (jobs []*AsyncOttoJob, err error) {
	globalJobListMutex.Lock()

	for job,_ := range globalJobList {
		jobs = append(jobs, job)
	}

	globalJobListMutex.Unlock()
	return jobs, nil
}


//Function declarations for master VM
func (ctl *HIDController) jsType(call otto.FunctionCall) (res otto.Value) {
	if ctl.Keyboard == nil { log.Println(ErrNoKeyboard);  oErr,_ := otto.ToValue(ErrNoKeyboard); return oErr }
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
	var partial string
	if len(outStr) > 15 {
		partial = outStr[:15]
	} else {
		partial = outStr
	}
	log.Printf("HIDScript type: Typing '%s ...' on HID keyboard device '%s'\n", partial, ctl.Keyboard.DevicePath)
	err = ctl.Keyboard.StringToPressKeySequence(outStr)
	if err != nil {
		log.Printf("HIDScript type: Couldn't type out `%s` on %v\n", outStr, ctl.Keyboard.DevicePath)
		return
	}
	return
}

func (ctl *HIDController) jsLayout(call otto.FunctionCall) (res otto.Value) {
	if ctl.Keyboard == nil { log.Println(ErrNoKeyboard);  oErr,_ := otto.ToValue(ErrNoKeyboard); return oErr }
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

func (ctl *HIDController) jsTypingSpeed(call otto.FunctionCall) (res otto.Value) {
	if ctl.Keyboard == nil { log.Println(ErrNoKeyboard);  oErr,_ := otto.ToValue(ErrNoKeyboard); return oErr }
	typeDelay := call.Argument(0) //delay between keypresses in milliseconds
	typeJitter := call.Argument(1) //additional random jitter between keypresses, maximum in milliseconds

	if delay,err:= typeDelay.ToInteger();err != nil || delay < 0 {
		log.Printf("HIDScript typingSpeed: First argument has to be positive integer, representing the delay between key presses in milliseconds\n")
		return
	} else {
		//ToDo: this isn't thread safe at all, additionally it influences type speed of every other running Script
		ctl.Keyboard.KeyDelay = int(delay)
	}
	if jitter,err:= typeJitter.ToInteger();err != nil || jitter < 0 {
		log.Printf("HIDScript typingSpeed: Second argument has to be positive integer, representing the maximum of an additional random jitter in milliseconds\n")
		return
	} else {
		//ToDo: this isn't thread safe at all, additionally it influences type speed of every other running Script
		ctl.Keyboard.KeyDelayJitter = int(jitter)
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
//	log.Printf("HIDScript delay: Sleeping `%v` milliseconds\n", delay)

	select {
		case <- time.After(time.Millisecond * time.Duration(int(delay))):
			return
		case irq := <-call.Otto.Interrupt:
			irq()
	}
	return
}

//for pressing key combos
func (ctl *HIDController) jsPress(call otto.FunctionCall) (res otto.Value) {
	if ctl.Keyboard == nil { log.Println(ErrNoKeyboard);  oErr,_ := otto.ToValue(ErrNoKeyboard); return oErr }
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
//	log.Printf("HIDScript: Called WaitLED(%v, %v)\n", arg0, arg1)
	maskInt, err := arg0.ToInteger()
	if err != nil || !arg0.IsNumber() || !(maskInt >= 0 && maskInt <= MaskAnyOrNone) {
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
			log.Printf("HIDScript WaitLED: Second argument for `waitLED` is the timeout in milliseconds and has to be given as positive interger, but '%d' was given!\n", arg1)
			return
		}
		timeout = time.Duration(timeoutInt) * time.Millisecond
	default:
		log.Printf("HIDScript WaitLED: Second argument for `waitLED` is the timeout in milliseconds and has to be given as interger or omitted for infinite timeout\n")
		return
	}

	var changed *HIDLEDState
	if ctl.Keyboard == nil { err = ErrNoKeyboard } else {
		changed,err = ctl.Keyboard.WaitLEDStateChange(call.Otto.Interrupt, mask, timeout)
		if err == ErrIrq {
			panic(haltirq)
		}
	}


	errStr := ""
	if err != nil {errStr = fmt.Sprintf("%v",err)}
	resStruct := struct{
		ERROR bool
		ERRORTEXT string
		TIMEOUT bool
		NUM bool
		CAPS bool
		SCROLL bool
		COMPOSE bool
		KANA bool
	}{
		ERROR: err != nil,
		ERRORTEXT: errStr,
		TIMEOUT: err == ErrTimeout,
		NUM: err == nil && changed.NumLock,
		CAPS: err == nil && changed.CapsLock,
		SCROLL: err == nil && changed.ScrollLock,
		COMPOSE: err == nil && changed.Compose,
		KANA: err == nil && changed.Kana,
	}
	res,_ = call.Otto.ToValue(resStruct)

	//Event generation
	resText := "WaitLED ended:"
	if resStruct.ERROR { resText = " ERROR " + errStr} else {
		if resStruct.NUM { resText += " NUM" }
		if resStruct.CAPS { resText += " CAPS" }
		if resStruct.SCROLL { resText += " SCROLL" }
		if resStruct.COMPOSE { resText += " COMPOSE" }
		if resStruct.KANA { resText += " KANA" }
	}
	sJob, sVM, err := ctl.retrieveJobFromOtto(call.Otto)
	if err == nil {
		ctl.emitEvent(Event{
			Type: EventType_JOB_WAIT_LED_FINISHED,
			Job: sJob,
			Vm: sVM,
			Message: resText,
		})
	} else {
		ctl.emitEvent(Event{
			Type: EventType_JOB_WAIT_LED_FINISHED,
			Message: resText + " (unknown Job)",
		})
	}
	return
}

func (ctl *HIDController) jsWaitLEDRepeat(call otto.FunctionCall) (res otto.Value) {
	//arg0 has to be of type number, representing an LED MASK
	//arg1 repeat delay (number float)
	//arg2 repeat count (number integer)
	//arg3 is optional and represents the timeout in seconds, in case it isn't present, we set timeout to a year (=infinite in our context ;-))
	arg0 := call.Argument(0) //trigger mask
	arg1 := call.Argument(1) //minimum repeat count till trigger
	arg2 := call.Argument(2) //maximum interval between LED changes of same LED, to be considered as repeat
	arg3 := call.Argument(3) //timeout
	log.Printf("HIDScript: Called WaitLEDRepeat(%v, %v, %v, %v)\n", arg0, arg1, arg2, arg3)

	//arg0: Typecheck trigger mask
	maskInt, err := arg0.ToInteger()
	if err != nil || !arg0.IsNumber() || !(maskInt >= 0 && maskInt <= MaskAnyOrNone) {
		//We don't mention KANA and COMPOSE in the error message
		log.Printf("HIDScript WaitLEDRepeat: First argument for `waitLED` has to be a bitmask representing LEDs (NUM | CAPS | SCROLL | ANY).\nError location:  %v\n", call.CallerLocation())
		return
	}
	mask := byte(maskInt)

	//arg1: repeat count (positive int > 0)
	repeatCount := 3 //default (first LED change is usually too slow to count, thus we need 4 changes and ultimately end up initial LED state)
	switch {
	case arg1.IsUndefined():
		log.Printf("HIDScript WaitLEDRepeat: No repeat count given, defaulting to '%v' led changes\n", repeatCount)
	case arg1.IsNumber():
		repeatInt, err := arg1.ToInteger()
		if err != nil || repeatInt < 1 {
			log.Printf("HIDScript WaitLEDRepeat: Second argument for `waitLEDRepeat` is the repeat count and has to be provided as positive interger, but '%d' was given!\n", arg1)
			return
		}
		repeatCount = int(repeatInt)
	default:
		log.Printf("HIDScript WaitLEDRepeat: Second argument for `waitLEDRepeat` is the repeat count and has to be provided as positive interger or omitted for default of '%v'\n", repeatCount)
		return
	}


	//arg2: //maximum interval between LED changes of same LED in milliseconds, to be considered as repeat
	maxInterval := 800 * time.Millisecond //default 800 ms
	switch {
	case arg2.IsUndefined():
		log.Printf("HIDScript WaitLEDRepeat: No maximum interval given (time allowed between LED changes, to be considered as repeat). Using default of %v\n", maxInterval)
	case arg2.IsNumber():
		//		log.Printf("Timeout given: %v\n", arg1)
		maxIntervalInt, err := arg2.ToInteger()
		if err != nil || maxInterval < 0 {
			log.Printf("HIDScript WaitLEDRepeat: Third argument for `waitLEDRepeat` is the maximum interval between LED changes in milliseconds and has to be provided as positive interger, but '%d' was given!\n", arg1)
			return
		}
		maxInterval = time.Duration(maxIntervalInt) * time.Millisecond
	default:
		log.Printf("HIDScript WaitLEDRepeat: Third argument for `waitLEDRepeat` is the maximum interval between LED changes in milliseconds and has to be provided as positive interger or omitted to default to '%d'\n", maxInterval)
		return
	}


	//arg3: Typecheck timeout (positive integer or undefined)
	timeout := time.Hour * 24 * 365
	switch {
	case arg3.IsUndefined():
		log.Printf("HIDScript WaitLEDRepeat: No timeout given setting to a year\n")
	case arg3.IsNumber():
		//		log.Printf("Timeout given: %v\n", arg1)
		timeoutInt, err := arg3.ToInteger()
		if err != nil || timeoutInt < 0 {
			log.Printf("HIDScript WaitLEDRepeat: Second argument for `waitLED` is the timeout in milliseconds and has to be given as positive interger, but '%d' was given!\n", arg1)
			return
		}
		timeout = time.Duration(timeoutInt) * time.Millisecond
	default:
		log.Printf("HIDScript WaitLEDRepeat: Second argument for `waitLED` is the timeout in milliseconds and has to be given as interger or omitted for infinite timeout\n")
		return
	}


	var changed *HIDLEDState
	if ctl.Keyboard == nil { err = ErrNoKeyboard } else {
		log.Printf("HIDScript: Waiting for repeated LED change. Mask for considered LEDs: %v, Minimum repeat count: %v, Maximum repeat delay: %v, Timeout: %v\n", mask, repeatCount, maxInterval, timeout)
		changed,err = ctl.Keyboard.WaitLEDStateChangeRepeated(call.Otto.Interrupt, mask, repeatCount, maxInterval, timeout)
		if err == ErrIrq {
			panic(haltirq)
		}
	}



	errStr := ""
	if err != nil {errStr = fmt.Sprintf("%v",err)}
	resStruct := struct{
		ERROR bool
		ERRORTEXT string
		TIMEOUT bool
		NUM bool
		CAPS bool
		SCROLL bool
		COMPOSE bool
		KANA bool
	}{
		ERROR: err != nil,
		ERRORTEXT: errStr,
		TIMEOUT: err == ErrTimeout,
		NUM: err == nil && changed.NumLock,
		CAPS: err == nil && changed.CapsLock,
		SCROLL: err == nil && changed.ScrollLock,
		COMPOSE: err == nil && changed.Compose,
		KANA: err == nil && changed.Kana,
	}
	res,_ = call.Otto.ToValue(resStruct)

	//Event generation
	resText := "WaitLED ended:"
	if resStruct.ERROR { resText = " ERROR " + errStr} else {
		if resStruct.NUM { resText += " NUM" }
		if resStruct.CAPS { resText += " CAPS" }
		if resStruct.SCROLL { resText += " SCROLL" }
		if resStruct.COMPOSE { resText += " COMPOSE" }
		if resStruct.KANA { resText += " KANA" }
	}
	sJob, sVM, err := ctl.retrieveJobFromOtto(call.Otto)
	if err == nil {
		ctl.emitEvent(Event{
			Type: EventType_JOB_WAIT_LED_REPEATED_FINISHED,
			Job: sJob,
			Vm: sVM,
			Message: resText,
		})
	} else {
		ctl.emitEvent(Event{
			Type: EventType_JOB_WAIT_LED_REPEATED_FINISHED,
			Message: resText + " (unknown Job)",
		})
	}
	return
}

// Move mouse relative in given mouse units (-127 to +127 per axis)
func (ctl *HIDController) jsMove(call otto.FunctionCall) (res otto.Value) {
	if ctl.Mouse == nil { log.Println(ErrNoMouse); return }

	argx := call.Argument(0)
	argy := call.Argument(1)
	log.Printf("HIDScript: Called move(%v, %v)\n", argx, argy)

	var x,y int
	if lx,err:= argx.ToInteger();err != nil || lx < -127 || lx > 127 {
		log.Printf("HIDScript move: First argument has to be integer between -127 and +127 describing relative mouse movement on x-axis\n")
		return
	} else { x = int(lx) }
	if ly,err:= argy.ToInteger();err != nil || ly < -127 || ly > 127 {
		log.Printf("HIDScript move: Second argument has to be integer between -127 and +127 describing relative mouse movement on y-axis\n")
		return
	} else { y = int(ly) }
	x8 := int8(x)
	y8 := int8(y)
	ctl.Mouse.Move(x8,y8)
	return
}

// Move mouse relative in across given distance in mouse units, devide into substeps of 1 DPI per step (parameters uint6 -32768 to +32767 per axis)
func (ctl *HIDController) jsMoveStepped(call otto.FunctionCall) (res otto.Value) {
	if ctl.Mouse == nil { log.Println(ErrNoMouse); return }
	argx := call.Argument(0)
	argy := call.Argument(1)
//	log.Printf("HIDScript: Called moveStepped(%v, %v)\n", argx, argy)

	var x,y int
	if lx,err:= argx.ToInteger();err != nil || lx < -32768 || lx > 32767 {
		log.Printf("HIDScript moveStepped: First argument has to be integer between -32768 and +32767 describing relative mouse movement on x-axis\n")
		return
	} else { x = int(lx) }
	if ly,err:= argy.ToInteger();err != nil || ly < -32768 || ly > 32767 {
		log.Printf("HIDScript moveStepped: Second argument has to be integer between -32768 and +32767 describing relative mouse movement on y-axis\n")
		return
	} else { y = int(ly) }
	x16 := int16(x)
	y16 := int16(y)
	ctl.Mouse.MoveStepped(x16,y16)
	return
}


// Move mouse to absolute position (-1.0 to +1.0 per axis)
func (ctl *HIDController) jsMoveTo(call otto.FunctionCall) (res otto.Value) {
	if ctl.Mouse == nil { log.Println(ErrNoMouse); return }
	argx := call.Argument(0)
	argy := call.Argument(1)
	log.Printf("HIDScript: Called moveTo(%v, %v)\n", argx, argy)

	var x,y float64
	if lx,err:= argx.ToFloat();err != nil || lx < -1.0 || lx > 1.0 {
		log.Printf("HIDScript move: First argument has to be a float between -1.0 and +1.0 describing relative mouse movement on x-axis\n")
		return
	} else { x = float64(lx) }
	if ly,err:= argy.ToFloat();err != nil || ly < -1.0 || ly > 1.0 {
		log.Printf("HIDScript move: Second argument has to be a float between -1.0 and +1.0 describing relative mouse movement on y-axis\n")
		return
	} else { y = float64(ly) }
	ctl.Mouse.MoveTo(x,y)
	return
}

func (ctl *HIDController) jsButton(call otto.FunctionCall) (res otto.Value) {
	if ctl.Mouse == nil { log.Println(ErrNoMouse); return }
	//arg0 has to be of type number, representing a bitmask for BUTTON1..3
	arg0 := call.Argument(0)
	log.Printf("HIDScript: Called button(%v)\n", arg0)
	maskInt, err := arg0.ToInteger()
	maskByte := byte(maskInt)
	if err != nil || !arg0.IsNumber() || maskInt != int64(maskByte) || !(maskByte >= 0 && maskByte <= BUTTON3) {
		log.Printf("HIDScript button: Argument has to be a bitmask representing Buttons (BT1 || BT2 || BT3).\nError location:  %v\n", call.CallerLocation())
		return
	}


	var bt [3]bool
	if maskByte & BUTTON1 > 0 { bt[0] = true}
	if maskByte & BUTTON2 > 0 { bt[1] = true}
	if maskByte & BUTTON3 > 0 { bt[2] = true}
	err = ctl.Mouse.SetButtons(bt[0], bt[1], bt[2])

	return
}

func (ctl *HIDController) jsClick(call otto.FunctionCall) (res otto.Value) {
	if ctl.Mouse == nil { log.Println(ErrNoMouse); return }
	//arg0 has to be of type number, representing a bitmask for BUTTON1..3
	arg0 := call.Argument(0)
	log.Printf("HIDScript: Called click(%v)\n", arg0)
	maskInt, err := arg0.ToInteger()
	maskByte := byte(maskInt)
	if err != nil || !arg0.IsNumber() || maskInt != int64(maskByte) || !(maskByte >= 0 && maskByte <= BUTTON3) {
		log.Printf("HIDScript click: Argument has to be a bitmask representing Buttons (BT1 || BT2 || BT3).\nError location:  %v\n", call.CallerLocation())
		return
	}


	var bt [3]bool
	if maskByte & BUTTON1 > 0 { bt[0] = true}
	if maskByte & BUTTON2 > 0 { bt[1] = true}
	if maskByte & BUTTON3 > 0 { bt[2] = true}
	err = ctl.Mouse.Click(bt[0], bt[1], bt[2])

	return
}

func (ctl *HIDController) jsDoubleClick(call otto.FunctionCall) (res otto.Value) {
	if ctl.Mouse == nil { log.Println(ErrNoMouse); return }

	//arg0 has to be of type number, representing a bitmask for BUTTON1..3
	arg0 := call.Argument(0)
	log.Printf("HIDScript: Called doubleClick(%v)\n", arg0)
	maskInt, err := arg0.ToInteger()
	maskByte := byte(maskInt)
	if err != nil || !arg0.IsNumber() || maskInt != int64(maskByte) || !(maskByte >= 0 && maskByte <= BUTTON3) {
		log.Printf("HIDScript doubleClick: Argument has to be a bitmask representing Buttons (BT1 || BT2 || BT3).\nError location:  %v\n", call.CallerLocation())
		return
	}


	var bt [3]bool
	if maskByte & BUTTON1 > 0 { bt[0] = true}
	if maskByte & BUTTON2 > 0 { bt[1] = true}
	if maskByte & BUTTON3 > 0 { bt[2] = true}
	err = ctl.Mouse.DoubleClick(bt[0], bt[1], bt[2])

	return
}


func (ctl *HIDController) initMasterVM() (err error) {
	ctl.vmMaster = otto.New()
	ctl.initVM(ctl.vmMaster)
	return nil
}


func (ctl *HIDController) initVM(vm *otto.Otto) (err error) {
	err = vm.Set("NUM", MaskNumLock)
	if err != nil { return err }
	err = vm.Set("CAPS", MaskCapsLock)
	if err != nil { return err }
	err = vm.Set("SCROLL", MaskScrollLock)
	if err != nil { return err }
	err = vm.Set("COMPOSE", MaskCompose)
	if err != nil { return err }
	err = vm.Set("KANA", MaskKana)
	if err != nil { return err }
	err = vm.Set("ANY", MaskAny)
	if err != nil { return err }
	err = vm.Set("ANY_OR_NONE", MaskAnyOrNone) //special masking, report back LED state updates with no change (re-attachment of Keyboard on Win/Linux)
	if err != nil { return err }

	err = vm.Set("BT1", BUTTON1)
	if err != nil { return err }
	err = vm.Set("BT2", BUTTON2)
	if err != nil { return err }
	err = vm.Set("BT3", BUTTON3)
	if err != nil { return err }
	err = vm.Set("BTNONE", 0)
	if err != nil { return err }


	err = vm.Set("typingSpeed", ctl.jsTypingSpeed) //This function influences all scripts
	if err != nil { return err }

	err = vm.Set("type", ctl.jsType)
	if err != nil { return err }
	err = vm.Set("delay", ctl.jsDelay)
	if err != nil { return err }
	err = vm.Set("press", ctl.jsPress)
	if err != nil { return err }
	err = vm.Set("waitLED", ctl.jsWaitLED)
	if err != nil { return err }
	err = vm.Set("waitLEDRepeat", ctl.jsWaitLEDRepeat)
	if err != nil { return err }
	err = vm.Set("layout", ctl.jsLayout) // this influences all scripts
	if err != nil { return err }

	err = vm.Set("move", ctl.jsMove)
	if err != nil { return err }
	err = vm.Set("moveStepped", ctl.jsMoveStepped)
	if err != nil { return err }
	err = vm.Set("moveTo", ctl.jsMoveTo)
	if err != nil { return err }

	err = vm.Set("button", ctl.jsButton)
	if err != nil { return err }
	err = vm.Set("click", ctl.jsClick)
	if err != nil { return err }
	err = vm.Set("doubleClick", ctl.jsDoubleClick)
	if err != nil { return err }
	return nil
}
