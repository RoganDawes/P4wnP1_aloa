// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)

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


func InitCompLogger()  {

	hvue.NewComponent(
		"logger",
		hvue.Template(compLoggerTemplate),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := &struct {
				*js.Object
				Pagination *jsDataTablePagination `js:"pagination"`
			}{Object:O()}

			data.Pagination = newPagination(0, 1)

			return data
		}),
		hvue.Method("logLevelClass", LogLevelClass),
		hvue.PropObj("max-entries", hvue.Types(hvue.PNumber), hvue.Default(5)),

		hvue.Computed("classFromLevel", func(vm *hvue.VM) interface{} {
			return "info"
		}),
		hvue.Method("formatDate", func(vm *hvue.VM, timestamp *js.Object) interface{} {
			return js.Global.Get("Quasar").Get("utils").Get("date").Call("formatDate", timestamp, "YYYY-MM-DD HH:mm:ss Z")
		}),
		hvue.Computed("logArray",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("EventProcessor").Get("logArray")
			}),
	)
	//return o.NewComponent()
}

const (

	compLoggerTemplate = `
<q-page padding>
	<q-card>
		<div>
			<q-table
				:data="logArray"
				:columns="[{name:'logTime', field: 'time', label: '时间', align: 'left'}, {name:'logSource', field: 'source', label: '事件源', align: 'left'}, {name:'logLevel', field: 'level', label: '级别', align: 'left'}, {name:'logMessage', field: 'message', label: '消息', align: 'left'}]"
				row-key="name"
				:pagination="pagination"
				hide-bottom
			>
  <q-td slot="body-cell-logTime" slot-scope="props" :props="props">
    {{ formatDate(props.value) }}
  </q-td>
			</q-table>
		</div>
	</q-card>

<!--
	<div class="logger">
	<table class="log-entries">
		<tr>
			<th>时间</th>
			<th>事件源</th>
			<th>级别</th>
			<th>消息</th>
		</tr>
        <tr v-for="(logEntry,idx) in logArray" :key="idx" :class="logLevelClass(logEntry.level)">
			<td class="log-entry-time">{{ logEntry.time }}</td>
	        <td class="log-entry-source">{{ logEntry.source }}</td>
			<td class="log-entry-level">{{ logEntry.level }}</td>
			<td class="log-entry-message">{{ logEntry.message }}</td>
	    </tr>
	</table>
	</div>
-->
</q-page>
`
)

