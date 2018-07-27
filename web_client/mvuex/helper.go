package mvuex

//ToDo:	check for Vuex in js.Global scope and panic if missing

import (
	"github.com/gopherjs/gopherjs/js"
	"reflect"
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


		//create a pointer to a new instance of this struct
		pStructInstance := reflect.New(targetType)


		//check if first field is *js.Object
		field0 := pStructInstance.Elem().Field(0)
		if field0.Kind() != reflect.Ptr || field0.Elem().Kind() != kindJsObjectType {
			return result, eFirstFieldIsntPtrJsObject
		}

		//Assign the sourceValue to the first field of the struct, which is assume to be *js.Object
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
