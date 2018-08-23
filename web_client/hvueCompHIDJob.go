// +build js

package main

import (
	"github.com/HuckRidgeSW/hvue"
)


func InitCompHIDJob() {
	hvue.NewComponent(
		"hidjob",
		hvue.Template(compHIDJobTemplate),
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

	compHIDJobTemplate = `
<div :style="{ 'display': 'flex' }">
<div class="jobstate-entry" :class="jobstate"><span>{{ job.vmId }}: {{ job.id }}</span></div>
<div v-if="job.hasSucceeded || job.hasFailed">{{ job.textResult }}</div>
<div>{{ job.textSource }}</div>
</div>
`
)

