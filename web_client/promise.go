package main

import (
	"github.com/gopherjs/gopherjs/js"
)


type Promise struct {
	*js.Object
	State string `js:"state"` //ToDo: change to dedicated type
}



type PromiseFunc func() (result interface{}, err error)


func NewPromise(pf PromiseFunc) (p *Promise) {
	f := func(resolve *js.Object, reject *js.Object) {
		p := &struct {
			*js.Object
			Resolve *js.Object `js:"resolve"`
			Reject *js.Object `js:"reject"`
		}{Object:O()}
		p.Resolve = resolve
		p.Reject = reject

		fResolve := func(args interface{}) (res interface{}) {
			return p.Call("resolve",args)
		}
		fReject := func(args interface{}) (res interface{}) {
			return p.Call("reject",args)
		}

		go func() {
			res,err := pf()
			if err != nil {
				fReject(js.Global.Get("Error").New(err.Error()))
			} else {
				fResolve(res)
			}
		}()

	}

	jsP := js.Global.Get("Promise").New(f)

	p = &Promise{
		Object: jsP,
	}
	return
}