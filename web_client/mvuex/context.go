package mvuex

import "github.com/gopherjs/gopherjs/js"

// Actions use a context instead of the store itself: https://vuex.vuejs.org/guide/actions.html
type ActionContext struct { //Don't use Context as name to avoid conflicts
	*js.Object

	Getters		*js.Object	`js:"getters"`
	Commit		func(...interface{}) *js.Object	`js:"commit"`
	Dispatch	func(...interface{}) *js.Object	`js:"dispatch"`
	State		*js.Object `js:"state"`
	RootGetters *js.Object `js:"rootGetters"`
	RootState	*js.Object `js:"rootState"`

}

