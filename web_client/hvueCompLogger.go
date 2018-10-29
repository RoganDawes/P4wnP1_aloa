// +build js

package main

import (
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
//		hvue.DataFunc(NewLoggerData),
//		hvue.MethodsOf(&CompLoggerData{}),
		hvue.Method("logLevelClass", LogLevelClass),
		hvue.PropObj("max-entries", hvue.Types(hvue.PNumber), hvue.Default(5)),
		hvue.Created(func(vm *hvue.VM) {
			println("OnCreated")
//			vm.Call("StartListening")
		}),
		hvue.Destroyed(func(vm *hvue.VM) {
			println("OnDestroyed")
//			vm.Call("StopListening")
		}),

		hvue.Computed("classFromLevel", func(vm *hvue.VM) interface{} {
			return "info"
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
<q-page>
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
</q-page>
`
)

