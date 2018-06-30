package hid

import (
	"github.com/robertkrimen/otto"
	"errors"
	"fmt"
	"sync"
	"log"
	"context"
	"encoding/json"
)

var vmNum = 0
var jobNum = 0


type AsyncOttoJob struct {
	ctx            context.Context
	Cancel         context.CancelFunc
	executingVM    *AsyncOttoVM
	Id             int
	Source         interface{}
	finishedNotify chan struct{}
	isFinished bool //finishedNotify could only be used to signal finish (blocking wait for chanel close), not for non blocking polling, thus we add a state var
	ResultErr      error
	ResultValue    otto.Value
}

func (job *AsyncOttoJob) Result() interface{} {
	goRes,_ := job.ResultValue.Export() //error is always nil (otto doc)
	return goRes
}

func (job *AsyncOttoJob) SetFinished() {
	job.executingVM = nil //avoid accessing the vm from this job again
	job.isFinished = true
	// we call the cancelFunc, this would only issue an interrupt to the vm, in case the VM would still be set executingVM, but assures ctx.Done listener could react
	// to the finished Job
	job.Cancel()
	close(job.finishedNotify)
}

func (job *AsyncOttoJob) WaitFinished() {
	<-job.finishedNotify
}

func (job *AsyncOttoJob) WaitResult()  (otto.Value, error) {
	job.WaitFinished()
	return job.ResultValue, job.ResultErr
}


func (job *AsyncOttoJob) ResultJsonString() (string, error) {
	goRes := job.Result() //error is always nil (otto doc)
	json, err := json.Marshal(goRes)
	if err != nil {return "",err}
	return string(json),nil
}



type AsyncOttoVM struct {
	vm           *otto.Otto
	isWorking    bool
	sync.Mutex
	Id int
}

func NewAsyncOttoVM(vm *otto.Otto) *AsyncOttoVM {
	vm.Interrupt = make(chan func(), 1) //set Interrupt channel
	res := &AsyncOttoVM{
		isWorking:   false,
		vm:          vm,
		Id:          vmNum,
	}
	vmNum++
	return res
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

func (avm *AsyncOttoVM) SetWorking(working bool) {
	avm.Lock()
	avm.isWorking = working
	avm.Unlock()
	fmt.Printf("VM %d set to isWorking: %v\n", avm.Id, working)
	return
}

func (avm *AsyncOttoVM) RunAsync(ctx context.Context, src interface{}) (job *AsyncOttoJob, err error) {
	if avm.IsWorking() {
		return job, errors.New(fmt.Sprintf("VM %d couldn't start new job, because it is still running one"))
	}

	avm.SetWorking(true)

	//create job
	jobNum++ //jobs start from 1 not 0
	ctx,cancel := context.WithCancel(ctx)
	job = &AsyncOttoJob{
		Id:             jobNum,
		ctx:            ctx,
		Cancel:         cancel,
		executingVM:    avm,
		Source:         src,
		finishedNotify: make(chan struct{}),
	}

	fmt.Printf("RunAsync called for VM %d\n", avm.Id)

	go func(avm *AsyncOttoVM) {
		select {
		case <-job.ctx.Done():
			if job.executingVM != nil {
				fmt.Printf("Job %d received IRQ, sending to VM%d\n", job.Id, avm.Id)
				job.executingVM.vm.Interrupt  <- func() {
					log.Printf("VM %d EXECUTED INTERRUPT FUNCTION\n", avm.Id)
					panic(haltirq)
				}
			} else {
				fmt.Printf("Job %d received IRQ, NOT sending to VM as not attached to any\n", job.Id)
			}
		}
	}(avm)

	go func(avm *AsyncOttoVM) {
		defer func() { //runs after avm.vm.Run() returns (because script finished a was interrupted)
			defer avm.SetWorking(false)
			if caught := recover(); caught != nil {
				fmt.Printf("VM %d CAUGHT INTERRUPT, ENDING JOB %d\n", avm.Id, job.Id)
				if caught == haltirq {
					job.ResultErr = errors.New(fmt.Sprintf("Execution of job %d on VM %d interrupted\n", job.Id, avm.Id))

					// signal Job finished
					job.SetFinished()
					return
				}
				panic(caught) //re-raise panic, as it isn't `haltirq`
			}
			return
		}()

		fmt.Printf("START JOB %d SCRIPT ON VM %d\n", job.Id, avm.Id) //DEBUG

		//short pre-run to set JobID and VMID (ignore errors)
		avm.vm.Run(fmt.Sprintf("JID=%d;VMID=%d;", job.Id, avm.Id))

		job.ResultValue, job.ResultErr = avm.vm.Run(job.Source) //store result
		job.SetFinished()                                       // signal job finished

		//DEBUG
		if job.ResultErr == nil {
			jRes,jErr := job.ResultJsonString()
			if jErr == nil {
				fmt.Printf("JOB %d on VM %d SUCCEEDED WITH RESULT: %s\n", job.Id, avm.Id, jRes)
			} else {
				fmt.Printf("JOB %d on VM %d SUCCEEDED BUT RESULT COULDN'T BE MARSHALED TO JSON: %v\n", job.Id, avm.Id, jErr)
			}
		} else {
			fmt.Printf("JOB %d on VM %d FAILED: %v\n", job.Id, avm.Id, job.ResultErr)
		}

	}(avm)

	return job,nil
}


func (avm *AsyncOttoVM) Run(ctx context.Context,src interface{}) (val otto.Value, res error) {
	job,err := avm.RunAsync(ctx, src)
	if err != nil { return val,err }
	return job.WaitResult()
}
