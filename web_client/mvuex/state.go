package mvuex

import (
	"github.com/gopherjs/gopherjs/js"
	"reflect"
)

func State(value interface{}) StoreOption {

	// Check if value is struct with *js.Object in first field
	if !checkIfJSStruct(reflect.TypeOf(value)) {
		panic(eFirstFieldIsNotPtrJsObject)
	}

	return func(c *StoreConfig) {
		if c.State != js.Undefined {
			//if state has been defined before
			panic("Cannot use mvuex.Sate more than once")
		}
		c.Object.Set("state", value)
	}
}
