package hid

import (
	"github.com/robertkrimen/otto"
	"errors"
	"fmt"
	"sync"
)

type AsyncOttoVM struct {
	vm           *otto.Otto
	ResultErr    *error
	ResultValue  otto.Value
	Finished chan bool
	isWorking    bool
	sync.Mutex
}

func NewAsyncOttoVM(vm *otto.Otto) *AsyncOttoVM {
	vm.Interrupt = make(chan func(), 1) //set Interrupt channel
	return &AsyncOttoVM{
		isWorking:    false,
		ResultValue:  otto.Value{},
		Finished: make(chan bool,1), //buffer of 1 as we don't want to block on setting the finished signal
		vm:           vm,
	}
}

func NewAsyncOttoVMClone(vm *otto.Otto) *AsyncOttoVM {
	return NewAsyncOttoVM(vm.Copy())
}

func (avm AsyncOttoVM) IsWorking() bool {
	avm.Lock()
	res := avm.isWorking
	avm.Unlock()
	return res

}

func (avm *AsyncOttoVM) Run(src interface{}) (val otto.Value, res error) {
	fmt.Printf("BLOCKING RUN start state: %+v Finished: %d\n",avm, len(avm.Finished))

	res = avm.RunAsync(src)
	if res != nil { return }
	fmt.Printf("BLOCKING RUN before wait state: %+v Finished: %d\n",avm, len(avm.Finished))
	defer fmt.Printf("BLOCKING RUN after wait state: %+v Finished: %d\n",avm, len(avm.Finished))
	return avm.WaitResult()
}

func (avm *AsyncOttoVM) RunAsync(src interface{}) (error) {
	avm.Lock()
	avm.isWorking = true
	// ToDo: This has to replaced by real job control, to preserve results in case waitResult() is called late (results have to be stored per job, not per VM)
	for len(avm.Finished) > 0 {
		fmt.Println("CONSUMING FINISH EVENT BEFORE VM REUSE")
		<-avm.Finished
	} // We consume old finish events (there was no call to waitResult() up to that point)
	avm.ResultErr = nil
	avm.ResultValue = otto.Value{}
	avm.Unlock()

	go func(avm *AsyncOttoVM) {


		defer func() {
fmt.Println("STARTING DEFER FUNC")

			if caught := recover(); caught != nil {
				if caught == halt {
fmt.Println("VM CANCELED")
					nErr := errors.New("VM execution cancelled")
					avm.ResultErr = &nErr

					avm.Lock()
					avm.isWorking = false
					avm.Unlock()

					//destroy reference to Otto
					//avm.vm = nil

					//signal finished
					avm.Finished <- true
					return
				}
				panic(caught) //re-raise panic, as it isn't `halt`
			}
			avm.Lock()
			avm.isWorking = false
			avm.Unlock()
		}()

		//		fmt.Println("Running vm")
		tmpValue, tmpErr := avm.vm.Run(src)
		avm.ResultValue = tmpValue //store value first, to have it accessible when error is retrieved from channel
		avm.ResultErr = &tmpErr
		avm.Finished <- true
		//		fmt.Println("Stored vm result")



		//avm.setIsWorking(false)
	}(avm)
	return nil
}

func (avm *AsyncOttoVM) WaitResult()  (val otto.Value, err error) {

	//Always run async
	/*
	if !avm.runningAsync {
		return otto.Value{},errors.New("AsyncVM isn't running an async script from which a result could be received")
	}
	*/

	if avm.vm == nil {
		return val,errors.New("Async vm isn't valid anymore")
	}

	//wait for finished signal
	<- avm.Finished

	//destroy reference to vm
	//avm.vm = nil
	return avm.ResultValue, *avm.ResultErr
}

func (avm *AsyncOttoVM) Cancel() error {
	if avm.vm == nil {
		return errors.New("Async vm isn't valid anymore")
	}

	// Note: in between here, there's a small race condition, if avm.vm is set to nil
	// after the check above (f.e. the vm returns an result and thus the pointer is set to nil,
	// right after the check)

	//interrupt vm

	if avm.IsWorking() {
		if len(avm.vm.Interrupt) == 0 {
			fmt.Println("SENDING IRQ TO VM")
			avm.vm.Interrupt <- func() {
				panic(halt)
			}
			fmt.Printf("WAITING FOR RESULT TO BE SURE CANCEL WORKED: %+v\n", avm)

		} else {
			fmt.Println("VM ALREADY INTERRUPTED")
		}

		//consume result
		avm.WaitResult()
	} else {
		fmt.Println("VM NOT WORKING, NO NEED TO CANCEL")
	}





	return nil
}

