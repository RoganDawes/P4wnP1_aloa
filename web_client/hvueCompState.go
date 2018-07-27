package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
)

type CompStateData struct {
	*js.Object

}

func newCompStateData(vm *hvue.VM) interface{} {
	newVM := &CompToggleSwitchData{
		Object: O(),
	}
	return newVM
}

/*
type StoreState struct {
	*js.Object
	Counter int `js:"count"`
	Text string `js:"text"`
}

func createState() interface{} {
	state := StoreState{Object:O()}
	state.Counter = 1337
	state.Text = "Hi there"
	return state
}
*/
func InitCompState() {
	/*
	state := createState()
	store := mvuex.NewStore(
		mvuex.State(state),
		mvuex.Mutation("increment", func (store *mvuex.Store, state *StoreState, add int) {
			state.Counter += add
			return
		}),
		mvuex.Mutation("decrement", func (store *mvuex.Store, state *StoreState) {
			state.Counter--
			return
		}),
		mvuex.Mutation("setText", func (store *mvuex.Store, state *StoreState, newText string) {
			state.Text = newText
			return
		}),
	)

	js.Global.Set("store", store)
	*/

	hvue.NewComponent(
		"state",
		hvue.Template(compStateTemplate),
		hvue.DataFunc(newCompStateData),
		hvue.Computed("count", func(vm *hvue.VM) interface{} {
			return js.Global.Get("store").Get("state").Get("count")
		}),
		hvue.ComputedWithGetSet("text",
			func(vm *hvue.VM) interface{} {
				return js.Global.Get("store").Get("state").Get("text")
			},
			func(vm *hvue.VM, newValue *js.Object) {
				js.Global.Get("store").Call("commit", "setText", newValue)
			}),
		hvue.Method("increment", func(vm *hvue.VM, count *js.Object) {
			// normal way to access the store.commit() function
			js.Global.Get("store").Call("commit", "increment", count)

			//Quick way to access the commit function, possible as we have the Go instance of "store" in scope
			//store.Commit("increment", count)
		}),
		hvue.Method("decrement", func(vm *hvue.VM) {
			// normal way to access the store.commit() function
			js.Global.Get("store").Call("commit", "decrement")

			//Quick way to access the commit function, possible as we have the Go instance of "store" in scope
			//store.Commit("decrement")
		}),
	)
}

const (

	compStateTemplate = `
<div>
  <p>{{ count }}</p>
  <p>{{ text }}</p>
  <input v-model="text"></input>
  <p>
    <button @click="increment(1,2,3)">+</button>
	<button @click="increment(2)">+2</button>
    <button @click="decrement">-</button>
  </p>
</div>
`
)
