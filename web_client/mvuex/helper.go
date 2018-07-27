package mvuex

//ToDo:	check for Vuex in js.Global scope and panic if missing

import (
	"github.com/gopherjs/gopherjs/js"
	"reflect"
	"errors"
)


var (
	eTooFewMutationArgs       = errors.New("Mutation function has too few arguments (min 2)")
	eTooManyMutationArgs      = errors.New("Mutation function has too many arguments (max 3)")
	eWrongActionArgCount      = errors.New("Wrong argument count! An action handler takes 3 or 4 args: actionHandler(store *Store, context *ActionContext, state *{CustomStateType} [, callArg *{CustomArgType])")
	eTooFewMutationArgsOnCall = errors.New("Mutation function called with too few arguments from JavaScrip")
	eWrongFirstMutationArg    = errors.New("Mutation function has to have *Store as first argument type")
	eWrongFirstActionArg      = errors.New("Mutation function has to have *Store as first argument type")
	eWrongSecondActionArg      = errors.New("Mutation function has to have *ActionContext as second argument type")
	eWrongSecondMutationArg   = errors.New("The second argument of the mutation function has to be a pointer to a struct of the type used for state")
	eFirstFieldIsNotPtrJsObject = errors.New("The first field of the struct has to be of type *js.Object")


	jsObjectType = reflect.TypeOf(js.Object{})
	jsStoreType    = reflect.TypeOf(Store{})
	jsActioContextType    = reflect.TypeOf(ActionContext{})
)


func o() *js.Object { return js.Global.Get("Object").New() } //Helper to create *js.Object


func castToType(targetType reflect.Type, sourceVal *js.Object) (result reflect.Value, err error) {

	switch kind := targetType.Kind(); kind {
	case reflect.Int:
		//try to convert sourceVal to int before generating reflect.Value
		result = reflect.ValueOf(sourceVal.Int())
	case reflect.Int8:
		result = reflect.ValueOf(int8(sourceVal.Int64()))
	case reflect.Int16:
		result = reflect.ValueOf(int16(sourceVal.Int64()))
	case reflect.Int32:
		result = reflect.ValueOf(int32(sourceVal.Int64()))
	case reflect.Int64:
		result = reflect.ValueOf(sourceVal.Int64())
	case reflect.Float64:
		result = reflect.ValueOf(sourceVal.Float())
	case reflect.Float32:
		result = reflect.ValueOf(float32(sourceVal.Float()))
	case reflect.Bool:
		result = reflect.ValueOf(sourceVal.Bool())
	case reflect.Uint:
		result = reflect.ValueOf(uint(sourceVal.Uint64()))
	case reflect.Uint64:
		result = reflect.ValueOf(sourceVal.Uint64())
	case reflect.Uint32:
		result = reflect.ValueOf(uint32(sourceVal.Uint64()))
	case reflect.Uint16:
		result = reflect.ValueOf(uint16(sourceVal.Uint64()))
	case reflect.Uint8:
		result = reflect.ValueOf(uint8(sourceVal.Uint64()))
	case reflect.Uintptr:
		result = reflect.ValueOf(sourceVal.Unsafe())
	case reflect.String:
		result = reflect.ValueOf(sourceVal.String())
	case reflect.Struct:
		//WE ASSUME THAT THE FIRST FIELD OF THE STRUCT IS OF TYPE *js.Object
		//check if first field is *js.Object
		if !checkIfJSStruct(targetType) {
			return result, eFirstFieldIsNotPtrJsObject
		}

		//create a pointer to a new instance of this struct
		pStructInstance := reflect.New(targetType)


		//Assign the sourceValue to the first field of the struct, which is assume to be *js.Object
//		fN := pStructInstance.Elem().Type().Name()
//		println("Assigning to '" + fN + "': ", reflect.TypeOf(sourceVal).Elem().Name(), sourceVal)
		pStructInstance.Elem().Field(0).Set(reflect.ValueOf(sourceVal))

		result = pStructInstance.Elem()
	case reflect.Ptr:
		//follow pointer one level
		derefType := targetType.Elem()
		//recursive call
		derefVal,err := castToType(derefType, sourceVal)
		if err != nil { return result, err}
		//println("dereferenced Value of type ", derefType.Kind().String(), ": ", derefVal)

		//create a pointer to the dereferenced value after it has been created itself
		result = derefVal.Addr()
	case reflect.Interface:
		result = reflect.ValueOf(sourceVal.Interface())
	default:
		// ToDo: func parsing
		println("No conversion for following type implemented", kind.String() , " from ", sourceVal)
	}

	return result, nil
}


// checks if the obj given is a struct, with *js.Object type in first field
func checkIfJSStruct(objType reflect.Type) bool {

	//check if Struct
	if objType.Kind() != reflect.Struct { return false } //not a struct
	// fetch first field
	typeField0 := objType.Field(0).Type
	//check if first field is pointer
	if typeField0.Kind() != reflect.Ptr { return false } //not a pointer
	// dereference ptr and check if equal to type js.Object
	if typeField0.Elem() != jsObjectType { return false } // not pointing to js.Object
	return true
}
