// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
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
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				ShowDetails bool `js:"ShowDetails"`
			}{Object: O()}
			data.ShowDetails = false
			return data
		}),
		hvue.Method("cancel", func(vm *hvue.VM) {
			job := &jsHidJobState{Object:vm.Get("job")}
			println("Aborting job :", job.Id)
			vm.Get("$store").Call("dispatch", VUEX_ACTION_CANCEL_HID_JOB, job.Id)
		}),
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

	hvue.NewComponent(
		"job-detail-modal",
		hvue.Template(templateHIDJobDetails),
		hvue.ComputedWithGetSet(
			"visible",
			func(vm *hvue.VM) interface{} {
				return vm.Get("value")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				vm.Call("$emit", "input", newValue)
			},
		),
		hvue.PropObj(
			"value",
			hvue.Required,
			hvue.Types(hvue.PBoolean),
		),
		hvue.PropObj(
			"job",
			hvue.Required,
		),
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

	templateHIDJobDetails = `
<q-modal v-model="visible">
	<q-modal-layout>
		<q-toolbar slot="header">
			<q-toolbar-title>
				HIDScript job details
			</q-toolbar-title>
		</q-toolbar>

		<div class="row gutter-sm">
			<div class="col-3">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Job ID</q-item-tile>
						<q-item-tile>
							<q-input readonly v-model="job.id" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</div>
			<div class="col-3">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>VM ID</q-item-tile>
						<q-item-tile>
							<q-input readonly v-model="job.vmId" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</div>
			<div class="col-6">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>Sate</q-item-tile>
						<q-item-tile>
							<q-input readonly :color="jobcolor" v-model="jobstate" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</div>

			<div class="col-12">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>HIDScript result</q-item-tile>
						<q-item-tile>
							<q-input readonly :color="jobcolor" v-model="job.textResult" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</div>

			<div class="col-12">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile label>HIDScript source</q-item-tile>
						<q-item-tile>
							<q-input readonly type="textarea" v-model="job.textSource" inverted></q-input>
						</q-item-tile>
					</q-item-main>
				</q-item>
			</div>


			<div class="col-12">
				<q-item tag="label">
					<q-item-main>
						<q-item-tile>
							<q-btn color="secondary" v-close-overlay label="close" />
						</q-item-tile>
					</q-item-main>
				</q-item>
			</div>
		</div>
	</q-modal-layout>
</q-modal>
`


	compHIDJobOverViewItemTemplate = `
<q-item highlight>
	<job-detail-modal v-model="ShowDetails" :job="job"></job-detail-modal>

	<q-item-side :icon="jobicon" :color="jobcolor" />
	<q-item-main>
		<q-item-tile label>Job {{ job.id }}</q-item-tile>
		<q-item-tile sublabel>State {{ jobstate }} </q-item-tile>
	</q-item-main>


   	<q-item-side right>
		<div class="row no-wrap">
			<div v-if="!job.hasSucceeded && !job.hasFailed">
				<q-btn flat round dense color="negative" icon="cancel" @click="cancel">
					<q-tooltip>
						cancel HIDScript job {{ job.id }}
					</q-tooltip>
				</q-btn>
			</div>
			<div>
				<q-btn flat round dense icon="info" @click="ShowDetails=true">
					<q-tooltip>
						show HIDScript job details
					</q-tooltip>
				</q-btn>
			</div>
		</div>
	</q-item-side>

</q-item>
`



//{ "evtype": 0, "vmId": 2, "jobId": 3, "hasError": false, "result": "null", "error": "", "message": "Script started", "time": "2018-07-30 04:56:42.297533 +0000 UTC m=+7625.097825001" }
compHIDJobOverviewTemplate = `
	<q-card class="full-height">
		<q-list>
			<q-collapsible opened icon-toggle>
				<template slot="header">
					<q-item-main label="Running jobs" :sublabel="'(' + $store.getters.hidjobsRunning.length + ' running jobs)'"/>
					<q-item-side v-if="$store.getters.hidjobsRunning.length > 0" right>
						<q-btn icon="cancel" color="red" @click="$store.dispatch('cancelAllHIDJobs')" round inverted flat>
							<q-tooltip>
								cancel all running HIDScript jobs
							</q-tooltip>
						</q-btn>
					</q-item-side>
				</template>
				<hid-job-overview-item v-for="job in $store.getters.hidjobsRunning" :job="job" :key="job.id"></hid-job-overview-item>
			</q-collapsible>
		</q-list>

		<q-list>
			<q-collapsible opened icon-toggle>
				<template slot="header">
					<q-item-main label="Succeeded" :sublabel="'(' + $store.getters.hidjobsSucceeded.length + ' successful jobs)'"/>
					<q-item-side v-if="$store.getters.hidjobsSucceeded.length > 0" right>
						<q-btn icon="delete" color="red" @click="$store.dispatch('removeSucceededHidJobs')" round inverted flat>
							<q-tooltip>
								delete succeeded HID jobs from list
							</q-tooltip>
						</q-btn>
					</q-item-side>
				</template>
				<hid-job-overview-item v-for="job in $store.getters.hidjobsSucceeded" :job="job" :key="job.id"></hid-job-overview-item>
			</q-collapsible>
		</q-list>
		<q-list>
			<q-collapsible  opened icon-toggle>
				<template slot="header">
					<q-item-main label="Failed" :sublabel="'(' + $store.getters.hidjobsFailed.length + ' failed jobs)'"/>
					<q-item-side v-if="$store.getters.hidjobsFailed.length > 0" right>
						<q-btn icon="delete" color="red" @click="$store.dispatch('removeFailedHidJobs')" round inverted flat>
							<q-tooltip>
								delete failed HID jobs from list
							</q-tooltip>
						</q-btn>

					</q-item-side>
				</template>
				<hid-job-overview-item v-for="job in $store.getters.hidjobsFailed" :job="job" :key="job.id"></hid-job-overview-item>
			</q-collapsible>
		</q-list>
	</q-card>

`
)

