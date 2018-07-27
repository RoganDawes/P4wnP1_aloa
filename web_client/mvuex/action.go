package mvuex

import (
	"github.com/gopherjs/gopherjs/js"
	"reflect"

	"errors"
	"strconv"
)

func wrapGoActionFunc(reflectedGoFunc reflect.Value ) (jsFunc *js.Object, err error) {

	numGoArgs := reflectedGoFunc.Type().NumIn() //Number of arguments of the Go target method
	if numGoArgs < 3 || numGoArgs > 4{
		return nil, eWrongActionArgCount
	}
	// Check if first arg 0 is of type *Store
	if arg := reflectedGoFunc.Type().In(0); arg.Kind() != reflect.Ptr || arg.Elem() != jsStoreType {
		return nil, eWrongFirstActionArg
	}
	// Check if first arg 1 is of type *ActionContext
	if arg := reflectedGoFunc.Type().In(1); arg.Kind() != reflect.Ptr || arg.Elem() != jsActioContextType {
		return nil, eWrongSecondActionArg
	}
	//Check if the remaining args are pointers to structs with *js.Object as first field
	for i := 2; i < numGoArgs; i++ {
		if arg:=reflectedGoFunc.Type().In(i); arg.Kind() != reflect.Ptr || !checkIfJSStruct(arg.Elem()) {
			return nil, errors.New("Arg at position " + strconv.Itoa(i) +" isn't a pointer to a struct with *js.Object in first field")
		}
	}

	goCallArgTargetTypes := make([]reflect.Type, numGoArgs) //store handler parameter types in slice
	for i := 0; i < reflectedGoFunc.Type().NumIn(); i++ { goCallArgTargetTypes[i] = reflectedGoFunc.Type().In(i) }

	goCallArgsTargetValues := make([]reflect.Value,numGoArgs) //create a slice for the parameter values, used in the call to the Go function


	jsFunc = js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		// this: points to the store
		// arg0: points to a context representation of the store
		// arg1: point to an optional argument for the action (undefined otherwise)

		// argument passing to the handler
		// goHandler(store *Store, context *ActionContext, state *{CustomStateType} [, callArg *{CustomArgType])
		//	--> the store is the root store of Vuex (the handler should use the context instead)
		//  --> the context is a representation of the store, dedicated for this action (async access)
		//  --> == context.State, but casted to the Go type presented by the handler function
		//  --> (optional) in case the handler function takes an additional argument, arg1 from the JS call will be casted to this

		/*
		println("Action this:", this)
		println("Action args:", arguments)
		//Globalize args (context) for investigation
		js.Global.Set("actionargs", arguments)
		*/

		storeVal,err := castToType(goCallArgTargetTypes[0], this) //cast 'this' to type of first function arg (type = *Store)
		if err != nil { panic("Error converting JavaScript provided 'this' for action function to *Store: " + err.Error()) }
		goCallArgsTargetValues[0] = storeVal

		contextVal,err := castToType(goCallArgTargetTypes[1], arguments[0]) //cast arg0 to type of second function arg (type = *ActionContext)
		if err != nil { panic("Error converting JavaScript provided first argument for action function to *ActionContext: " + err.Error()) }
		goCallArgsTargetValues[1] = contextVal

		//extract state from context, in order to cast it to the provided type
		jsStateObj := arguments[0].Get("state")
		stateVal,err := castToType(goCallArgTargetTypes[2], jsStateObj) //cast 'context.state' to type of third function arg
		if err != nil { panic("Error converting JavaScript provided context.state for action function to *" + goCallArgTargetTypes[2].Elem().Name() + ": " + err.Error()) }
		goCallArgsTargetValues[2] = stateVal

		//Check if handler receives 4th arg
		if numGoArgs == 4 {
			// check if argument 1 is provided by JS
			if arguments[1] == js.Undefined {
				panic("The action handler awaits an argument of type " + goCallArgTargetTypes[3].Name() + " but the dispatched action doesn't provide this parameter")
			}

			// try to convert to target type
			actionParamVal,err := castToType(goCallArgTargetTypes[3], arguments[1])
			if err != nil { panic("Error converting JavaScript provided optional parameter for action function to *" + goCallArgTargetTypes[3].Elem().Name() + ": " + err.Error()) }
			goCallArgsTargetValues[3] = actionParamVal
		}

		// Call the go function and return the result
		return reflectedGoFunc.Call(goCallArgsTargetValues)
	})

	return jsFunc, nil
}


func Action(name string, goFunc interface{}) StoreOption {
	return func(c *StoreConfig) {
		//println("Creating ACTION FUNC")
		if c.Actions == js.Undefined { c.Actions = o() }

		reflectedGoFunc := reflect.ValueOf(goFunc)
		if reflectedGoFunc.Kind() != reflect.Func { //check if the provided interface is a go function
			panic("Action " + name + " is not a func")
		}

		//try to convert the provided function to a JavaScript function usable as Mutation
		jsFunc, err := wrapGoActionFunc(reflectedGoFunc)
		if err != nil {panic("Error exposing the action function '"+ name + "' to JavaScript: " + err.Error())}

		c.Actions.Set(name, jsFunc)
		//c.Mutations.Set(name, makeMethod(name, false, reflectGoFunc.Type(), reflectGoFunc))
	}
}
