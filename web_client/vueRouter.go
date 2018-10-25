package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"

)

//Very simple vue-router binding, support only Component templates for routes


type VueRouterConfig struct {
	*js.Object
	Routes *js.Object `js:"routes"`
}

type VueRouterOption func(*VueRouterConfig)

func VueRouterRoute(path, name, template string) VueRouterOption {
	route := struct {
		*js.Object
		Path string `js:"path"`
		Name string `js:"name"`
		Component *hvue.Config `js:"component"`
	}{Object:O()}
	route.Path = path
	if len(name) > 0 { route.Name = name }

	//use hvue.Config to generate an component object with given template
	component := &hvue.Config{Object:O()}
	hvue.Template(template)(component)
	route.Component = component


	return func(config *VueRouterConfig) {
		config.Routes.Call("push", route)
	}
}

func NewVueRouter(defaultRoute string, opts ...VueRouterOption) *js.Object {
	c := &VueRouterConfig{Object:O()}
	c.Routes = js.Global.Get("Array").New()

	for _,opt := range opts {
		opt(c)
	}

	jsrouter := js.Global.Get("VueRouter").New(c)
	if len(defaultRoute) > 0 {
		jsrouter.Call("replace", defaultRoute)
	}
	return jsrouter
}


