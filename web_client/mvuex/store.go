package mvuex

import (
	"github.com/gopherjs/gopherjs/js"
	"reflect"
)


type StoreOption func(*StoreConfig)

type Store struct {
	*js.Object

	 Getters	*js.Object	`js:"getters"`
	 Commit		func(...interface{}) *js.Object	`js:"commit"`
	 Dispatch	func(...interface{}) *js.Object	`js:"dispatch"`
	 Strict 	bool		`js:"strict"`
}

// StoreConfig is the config object for NewStore.
type StoreConfig struct {
	*js.Object
	State		*js.Object	`js:"state"`
	Mutations	*js.Object	`js:"mutations"`
	Actions		*js.Object	`js:"actions"`

	stateValue reflect.Value
}


// Option sets the options specified.
func (c *StoreConfig) Option(opts ...StoreOption) {
	for _, opt := range opts {
		opt(c)
	}
}

func NewStore(opts ...StoreOption) *Store {
	c := &StoreConfig{Object: o()}
	c.Option(opts...)
	store := &Store{Object: js.Global.Get("Vuex").Get("Store").New(c)}


	return store
}