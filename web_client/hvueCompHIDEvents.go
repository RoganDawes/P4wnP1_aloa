// +build js

package main

import (
	"github.com/HuckRidgeSW/hvue"
	"github.com/mame82/P4wnP1_go/common_web"
	"github.com/gopherjs/gopherjs/js"
)

type CompHIDEventsData struct {
	*js.Object
	Pagination *jsDataTablePagination `js:"pagination"`
}

func newCompHIDEventsData(vm *hvue.VM) interface{} {
	data := &CompHIDEventsData{ Object: O()	}
	data.Pagination = newPagination(0, 1)
	return data
}

func InitCompHIDEvents() {

	hvue.NewComponent(
		"hid-job-event-overview",
		hvue.Template(compHIDJobEventOverviewTemplate),
		hvue.DataFunc(newCompHIDEventsData),
		hvue.Computed("events",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("eventReceiver").Get("eventHidArray")
			}),
		hvue.Method("evIdToString", func(vm *hvue.VM, evID int64) (res string) {
			println("EvID", evID)
			return common_web.EventType_name[evID]
		}),
	)

}

const (
	//{ "evtype": 0, "vmId": 2, "jobId": 3, "hasError": false, "result": "null", "error": "", "message": "Script started", "time": "2018-07-30 04:56:42.297533 +0000 UTC m=+7625.097825001" }
	compHIDJobEventOverviewTemplate = `
	<q-card>
<!--		<div class="scroll" style="overflow: auto;max-height: 20vh;">	-->
		<div>
			<q-table
				:data="events"
				:columns="[{name:'type', field: 'evtype', label: 'Event Type', align: 'left'}, {name:'vmid', field: 'vmId', label: 'VM ID', align: 'left'}, {name:'jobid', field: 'jobId', label: 'Job ID', align: 'left'}, {name:'haserror', field: 'hasError', label: 'Has error', align: 'left'}, {name:'res', field: 'result', label: 'Result', align: 'left'}, {name:'errormsg', field: 'error', label: 'Error', align: 'left'}, {name:'msg', field: 'message', label: 'Message', align: 'left'}, {name:'timestamp', field: 'time', label: 'Time', align: 'left'}]"
				row-key="name"
				:pagination="pagination"
				hide-bottom
			>

				<q-td slot="body-cell-type" slot-scope="props" :props="props">
					{{ evIdToString(props.value) }}
				</q-td>
				<q-td slot="body-cell-msg" slot-scope="props" :props="props">
					{{ props.value.slice(0,30) }}
					<q-btn v-if="props.value.length > 30" dense icon="more_horiz">
						
						<q-popover>
							<div class="q-ma-md" style="max-width: 400px; max-height: 400px;">
								<pre>{{ props.value }}</pre>
							</div>
						</q-popover>
					</q-btn>
				</q-td>
			</q-table>
		</div>
	</q-card>
`
)
