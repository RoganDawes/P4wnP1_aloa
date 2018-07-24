package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
	pb "../proto/gopherjs"
	"../common"
	"io"
	"context"
	"sync"
)

var preservedCOmpLoggerData *CompLoggerData = nil

type CompLoggerData struct {
	*js.Object
	LogArray *js.Object `js:"logArray"`
	cancel context.CancelFunc
	*sync.Mutex
}



func (data *CompLoggerData) AddEntry(vm *hvue.VM, ev *pb.Event ) {
	go func() {
		data.Lock()
		defer data.Unlock()
		logEv, err := DeconstructEventLog(ev)
		if err != nil {
			println("Logger: Error adding log entry, provided event couldn't be converted to log event")
			return
		}

		data.LogArray.Call("push", logEv)

		//reduce to length (note: kebab case 'max-entries' is translated to camel case 'maxEntries' by vue)
		for data.LogArray.Length() > vm.Get("maxEntries").Int() {
			data.LogArray.Call("shift") // remove first element
		}

	}()


}

func (data *CompLoggerData) StartListening(vm *hvue.VM) {

	println("Start listening called", data, vm)
	ctx,cancel := context.WithCancel(context.Background())
	data.cancel = cancel

	evStream, err := Client.Client.EventListen(ctx, &pb.EventRequest{ListenType: common.EVT_LOG})
	if err != nil {
		cancel()
		println("Error listening fo Log events", err)
		return
	}

	go func() {
		defer cancel()
		for {
			event, err := evStream.Recv()
			if err == io.EOF { break }
			if err != nil { return }

			//println("Event: ", event)
			data.AddEntry(vm, event)

		}
		return
	}()
}


func (data *CompLoggerData) StopListening(vm *hvue.VM) {

	println("Stop listening called", data, vm)
	data.cancel()
}

func LogLevelClass(vm *hvue.VM, level int) string {
	prefix := "log-entry log-entry-level-"
	switch level {
	case 1:
		return prefix + "critical"
	case 2:
		return prefix + "error"
	case 3:
		return prefix + "warning"
	case 4:
		return prefix + "information"
	case 5:
		return prefix + "verbose"
	default:
		return prefix + "undefined"
	}
}

func NewLoggerData(vm *hvue.VM) interface{} {
	loggerVmData := &CompLoggerData{
		Object: js.Global.Get("Object").New(),
	}

	loggerVmData.Mutex = &sync.Mutex{}
	loggerVmData.LogArray = js.Global.Get("Array").New()

	return loggerVmData
}



func InitCompLogger()  {

	hvue.NewComponent(
		"logger",
		hvue.Template(compLoggerTemplate),
		hvue.DataFunc(NewLoggerData),
		hvue.MethodsOf(&CompLoggerData{}),
		hvue.Method("logLevelClass", LogLevelClass),
		hvue.PropObj("max-entries", hvue.Types(hvue.PNumber), hvue.Default(5)),
		hvue.Created(func(vm *hvue.VM) {
			println("OnCreated")
			vm.Call("StartListening")
		}),
		hvue.Destroyed(func(vm *hvue.VM) {
			println("OnDestroyed")
			vm.Call("StopListening")
		}),
		hvue.Activated(func(vm *hvue.VM) {
			println("OnActivated")
		}),
		hvue.Deactivated(func(vm *hvue.VM) {
			println("OnDeactivated")
		}),
		hvue.Mounted(func(vm *hvue.VM) {
			println("OnMounted")
		}),
		/*
		hvue.Updated(func(vm *hvue.VM) {
			println("Updated")
		}),
		*/

		hvue.Computed("classFromLevel", func(vm *hvue.VM) interface{} {
			return "info"
		}),
	)
	//return o.NewComponent()
}

const (

	compLoggerTemplate = `
	<div class="logger">
	<table class="log-entries">
		<tr>
			<th>time</th>
			<th>source</th>
			<th>level</th>
			<th>message</th>
		</tr>
        <tr v-for="(logEntry,idx) in logArray" :key="idx" :class="logLevelClass(logEntry.level)">
			<td class="log-entry-time">{{ logEntry.time }}</td>
	        <td class="log-entry-source">{{ logEntry.source }}</td>
			<td class="log-entry-level">{{ logEntry.level }}</td>
			<td class="log-entry-message">{{ logEntry.message }}</td>
	    </tr>
	</table>
	</div>

`
)

