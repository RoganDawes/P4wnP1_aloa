// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"github.com/mame82/P4wnP1_go/common_web"
)

type CompHIDJobsData struct {
	*js.Object
}

func newCompHIDJobsData(vm *hvue.VM) interface{} {

	cc := &CompHIDJobsData{
		Object: js.Global.Get("Object").New(),
	}

	return cc
}

func InitCompHIDJobs() {
	hvue.NewComponent(
		"hidjobs",
		hvue.Template(compHIDJobsTemplate),
		hvue.DataFunc(newCompHIDJobsData),
		hvue.Computed("events",
			func(vm *hvue.VM) interface{} {
				return vm.Store.Get("state").Get("eventLog").Get("eventHidArray")
			}),
		hvue.Computed("jobs",
			func(vm *hvue.VM) interface{} {
				jobList := vm.Store.Get("state").Get("hidJobList").Get("jobs")
				return js.Global.Get("Object").Call("values",jobList)
			}),
		hvue.Method("evIdToString", func (vm *hvue.VM, evID int64) (res string) {
			println("EvID",evID)
			return common_web.EventType_name[evID]
		}),
	)
}

const (
	//{ "evtype": 0, "vmId": 2, "jobId": 3, "hasError": false, "result": "null", "error": "", "message": "Script started", "time": "2018-07-30 04:56:42.297533 +0000 UTC m=+7625.097825001" }
	compHIDJobsTemplate = `
<div>
<div>
	<hidjob  v-for="job in jobs" :job="job" :key="job.id"></hidjob>
</div>
<table border="1">
<tr>
	<th>Event Type</th>
	<th>VM ID</th>
	<th>Job ID</th>
	<th>Has error</th>
	<th>Result</th>
	<th>Error</th>
	<th>Message</th>
	<th>Time</th>
</tr>
<tr v-for="e in events">
	<td>{{ evIdToString(e.evtype) }}</td>
	<td>{{ e.vmId }}</td>
	<td>{{ e.jobId }}</td>
	<td>{{ e.hasError }}</td>
	<td>{{ e.result }}</td>
	<td>{{ e.error }}</td>
	<td>{{ e.message }}</td>
	<td>{{ e.time }}</td>
</tr>

</table>
</div>
`
)

