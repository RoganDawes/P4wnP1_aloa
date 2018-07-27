package mvuex

import (
	"reflect"
	"github.com/gopherjs/gopherjs/js"
)

func wrapGoMutationFunc(reflectedGoFunc reflect.Value ) (jsFunc *js.Object, err error) {
	// A mutationfunction is assumed to have this prototype
	//  func(
	// 		store *Store,					//first argument is of type store
	// 		ptrToStateStruct struct{}, 		//ptr to a struct with same type as the struct provided as state
	// 		additionalArgs ...interface{}	//optional arguments for the mutation (Go types)
	// )

	numGoArgs := reflectedGoFunc.Type().NumIn() //Number of arguments of the Go target method
	if numGoArgs < 2 {
		return nil, eTooFewMutationArgs
	}
	if numGoArgs > 3 {
		return nil, eTooManyMutationArgs
	}
	if goArg0 := reflectedGoFunc.Type().In(0); goArg0.Kind() != reflect.Ptr || goArg0.Elem() != jsStoreType {
		return nil, eWrongFirstMutationArg
	}
	if goArg1 := reflectedGoFunc.Type().In(1); goArg1.Kind() != reflect.Ptr || goArg1.Elem().Kind() != reflect.Struct {
		return nil, eWrongSecondMutationArg
	}


	// Here we know, the goFunc has at least two args, with first arg being *Store type
	// and second arg being a custom data struct.
	// The JavaScript call received, should provide two args at minimum:
	//	- arg 0: state data instance
	//	- arg 1..n: arguments handed in when the respective mutation function is called via commit, in case the
	//    mutation function is called without arguments, arg 1 is of type "undefined"
	// Additionally, the "this" argument provides the store instance, which we hand to the Go function as first argument

	// following two lines moved out of the inner function to avoid rerunning
	goCallArgTargetTypes := make([]reflect.Type, numGoArgs)
	goCallArgsTargetValues := make([]reflect.Value,numGoArgs) //create call args slice, containing the store arg
	for i := 0; i < reflectedGoFunc.Type().NumIn(); i++ {
		goCallArgTargetTypes[i] = reflectedGoFunc.Type().In(i)
	}


	jsFunc = js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		//check if js provides enough args
		if len(arguments) < numGoArgs - 1 {
			panic(eTooFewMutationArgsOnCall.Error())
		}

		//Note: All the logic in MakeFunc ends up in the final JS function and reruns every tim the function is triggered

		storeVal,err := castToType(goCallArgTargetTypes[0], this) //cast 'this' to type of first function arg (type = *Store)
		if err != nil { panic("Error converting JavaScript provided argument for mutation function to *Store: " + err.Error()) }
		goCallArgsTargetValues[0] = storeVal



		//add the remaining args
		for idx,jsArg := range arguments {
			//If the target function in Go has less arguments than we got provided from JavaScript, we gonna ignore the rest
			targetIdx := idx+1 //offset by one, as we started with *Store as first arg for the Go function
			if targetIdx >= numGoArgs { break }

			//get method argument type at this poistion
			goTargetArgT := goCallArgTargetTypes[targetIdx]
			castedArg, err := castToType(goTargetArgT, jsArg)

			if err != nil { panic("Error converting JS object to "  + goTargetArgT.Kind().String()) }

			goCallArgsTargetValues[targetIdx] = castedArg
		}

		results := reflectedGoFunc.Call(goCallArgsTargetValues)

		return results
	})

	return jsFunc, nil
}

func Mutation(name string, goFunc interface{}) StoreOption {
	return func(c *StoreConfig) {
		//println("Creating MUTATION FUNC")
		if c.Mutations == js.Undefined { c.Mutations = o() }

		reflectedGoFunc := reflect.ValueOf(goFunc)
		if reflectedGoFunc.Kind() != reflect.Func { //check if the provided interface is a go function
			panic("Mutation " + name + " is not a func")
		}

		//try to convert the provided function to a JavaScript function usable as Mutation
		jsFunc, err := wrapGoMutationFunc(reflectedGoFunc)
		if err != nil {panic("Error exposing the mutation function '"+ name + "' to JavaScript: " + err.Error())}

		c.Mutations.Set(name, jsFunc)
		//c.Mutations.Set(name, makeMethod(name, false, reflectGoFunc.Type(), reflectGoFunc))
	}
}

