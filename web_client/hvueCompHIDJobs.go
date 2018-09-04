// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
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
		"hid-job-overview",
		hvue.Template(compHIDJobOverviewTemplate),
		hvue.DataFunc(newCompHIDJobsData),
		hvue.Computed("jobs",
			func(vm *hvue.VM) interface{} {
				jobList := vm.Get("$store").Get("state").Get("hidJobList").Get("jobs")
				return js.Global.Get("Object").Call("values",jobList)
			}),
	)

	hvue.NewComponent(
		"hid-job-event-overview",
		hvue.Template(compHIDJobEventOverviewTemplate),
		hvue.DataFunc(newCompHIDJobsData),
		hvue.Computed("events",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("eventReceiver").Get("eventHidArray")
			}),
		hvue.Method("evIdToString", func (vm *hvue.VM, evID int64) (res string) {
			println("EvID",evID)
			return common_web.EventType_name[evID]
		}),
	)

	hvue.NewComponent(
		"hid-job-overview-item",
		hvue.Template(compHIDJobOverViewItemTemplate),
		hvue.Computed("jobstate",
			func(vm *hvue.VM) interface{} {
				//fetch job and cast back to jobstate
				job := &jsHidJobState{Object:vm.Get("job")}
				switch {
				case job.HasFailed && !job.HasSucceeded:
					return "FAILED"
				case job.HasSucceeded && !job.HasFailed:
					return "SUCCEEDED"
				case !(job.HasFailed || job.HasSucceeded):
					return "RUNNING"
				default:
					return "UNKNOWN_STATE"

				}
			}),
		hvue.Computed("jobcolor",
			func(vm *hvue.VM) interface{} {
				//fetch job and cast back to jobstate
				job := &jsHidJobState{Object:vm.Get("job")}
				switch {
				case job.HasFailed && !job.HasSucceeded:
					return "negative"
				case job.HasSucceeded && !job.HasFailed:
					return "positive"
				case !(job.HasFailed || job.HasSucceeded):
					return "warning"
				default:
					return "info"

				}
			}),
		hvue.Computed("jobicon",
			func(vm *hvue.VM) interface{} {
				//fetch job and cast back to jobstate
				job := &jsHidJobState{Object:vm.Get("job")}
				switch {
				case job.HasFailed && !job.HasSucceeded:
					return "error"
				case job.HasSucceeded && !job.HasFailed:
					return "check_circle"
				case !(job.HasFailed || job.HasSucceeded):
					return "sync"
				default:
					return "help"

				}
			}),
		hvue.PropObj("job", hvue.Required),
	)

}

const (
	/*
// HIDJobList
type jsHidJobState struct {
*js.Object
Id             int64  `js:"id"`
VmId           int64  `js:"vmId"`
HasFailed      bool   `js:"hasFailed"`
HasSucceeded   bool   `js:"hasSucceeded"`
LastMessage    string `js:"lastMessage"`
TextResult     string `js:"textResult"`
LastUpdateTime string `js:"lastUpdateTime"` //JSON timestamp from server
ScriptSource   string `js:"textSource"`
}
 */


	compHIDJobOverViewItemTemplate = `
<q-item highlight>
	<q-item-side :icon="jobicon" :color="jobcolor" />
	<q-item-main>
		<q-item-tile label>Job {{ job.id }}</q-item-tile>
		<q-item-tile sublabel>State {{ jobstate }} </q-item-tile>
	</q-item-main>

   	<q-item-side right v-if="job.hasSucceeded || job.hasFailed">
		<q-btn flat round dense icon="more_horiz">
			<q-popover>
				{{ job.textResult }}
			</q-popover>
		</q-btn>
	</q-item-side>
</q-item>
`



	//{ "evtype": 0, "vmId": 2, "jobId": 3, "hasError": false, "result": "null", "error": "", "message": "Script started", "time": "2018-07-30 04:56:42.297533 +0000 UTC m=+7625.097825001" }
	compHIDJobOverviewTemplate = `
	<q-card class="q-ma-sm">
		<q-list>
			<q-list-header>HID Script jobs</q-list-header>
			<hid-job-overview-item v-for="job in jobs" :job="job" :key="job.id"></hid-job-overview-item>
		</q-list>
	</q-card>

`
	//{ "evtype": 0, "vmId": 2, "jobId": 3, "hasError": false, "result": "null", "error": "", "message": "Script started", "time": "2018-07-30 04:56:42.297533 +0000 UTC m=+7625.097825001" }
	compHIDJobEventOverviewTemplate = `
<q-page>
<div>
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

</q-page>
`
)

