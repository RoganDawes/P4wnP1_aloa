// +build js

package main

import (
	"github.com/mame82/hvue"
	"github.com/mame82/P4wnP1_aloa/common_web"
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
				return vm.Get("$store").Get("state").Get("EventProcessor").Get("eventHidArray")
			}),
		hvue.Method("formatDate", func(vm *hvue.VM, timestamp *js.Object) interface{} {
			return js.Global.Get("Quasar").Get("utils").Get("date").Call("formatDate", timestamp, "YYYY-MM-DD HH:mm:ss Z")
		}),
		hvue.Method("evIdToString", func(vm *hvue.VM, evID int64) (res string) {
			//println("EvID", evID)
			return common_web.EventTypeHIDName[evID]
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
				:columns="[{name:'timestamp', field: 'time', label: '时间', align: 'left'}, {name:'type', field: 'evtype', label: '事件类型', align: 'left'}, {name:'vmid', field: 'vmId', label: 'VM ID', align: 'left'}, {name:'jobid', field: 'jobId', label: '任务ID', align: 'left'}, {name:'haserror', field: 'hasError', label: '执行错误', align: 'left'}, {name:'res', field: 'result', label: '结果', align: 'left'}, {name:'errormsg', field: 'error', label: '错误', align: 'left'}, {name:'msg', field: 'message', label: '消息', align: 'left'}]"
				row-key="name"
				:pagination="pagination"
				hide-bottom
			>

				<q-td slot="body-cell-timestamp" slot-scope="props" :props="props">
					{{ formatDate(props.value) }}
				</q-td>

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
