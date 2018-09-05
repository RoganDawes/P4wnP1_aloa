// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/HuckRidgeSW/hvue"
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
				//return vm.Get("$store").Get("state").Get("hidJobList").Get("jobs")
				return vm.Get("$store").Get("getters").Get("hidjobs")
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
	<q-card class="full-height">
  		<q-card-title>
    		HIDScript jobs
  		</q-card-title>

		<q-list>
			<q-list-header>Running</q-list-header>
			<hid-job-overview-item v-for="job in $store.getters.hidjobsRunning" :job="job" :key="job.id"></hid-job-overview-item>
			<q-list-header>Succeeded</q-list-header>
			<hid-job-overview-item v-for="job in $store.getters.hidjobsSucceeded" :job="job" :key="job.id"></hid-job-overview-item>
			<q-list-header>Failed</q-list-header>
			<hid-job-overview-item v-for="job in $store.getters.hidjobsFailed" :job="job" :key="job.id"></hid-job-overview-item>
		</q-list>
	</q-card>

`
)

