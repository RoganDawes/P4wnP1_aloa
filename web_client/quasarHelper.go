// +build js

package main

import "github.com/gopherjs/gopherjs/js"

var GlobalQuasar = QuasarGetQuasar()

const (
	QUASAR_NOTIFICATION_TYPE_POSITIVE = "positive"
	QUASAR_NOTIFICATION_TYPE_NEGATIVE = "negative"
	QUASAR_NOTIFICATION_TYPE_WARNING = "warning"
	QUASAR_NOTIFICATION_TYPE_INFO = "info"

	QUASAR_NOTIFICATION_POSITION_TOP = "top"
	QUASAR_NOTIFICATION_POSITION_TOP_LEFT = "top-left"
	QUASAR_NOTIFICATION_POSITION_TOP_RIGHT = "top-right"
	QUASAR_NOTIFICATION_POSITION_LEFT = "left"
	QUASAR_NOTIFICATION_POSITION_CENTER = "center"
	QUASAR_NOTIFICATION_POSITION_RIGHT = "right"
	QUASAR_NOTIFICATION_POSITION_BOTTOM = "bottom"
	QUASAR_NOTIFICATION_POSITION_BOTTOM_LEFT = "bottom-left"
	QUASAR_NOTIFICATION_POSITION_BOTTOM_RIGHT = "bottom-right"
)

type Quasar struct {
	*js.Object
	Version string `js:"version"`
	Theme string `js:"theme"`
	Plugins map[string]*js.Object `js:"plugins"`
}

type QuasarNotification struct {
	*js.Object
	Message string `js:"message"`
	Detail string `js:"detail"`
	Type string `js:"type"`
	Color string `js:"color"`
	TextColor string `js:"textColor"`
	Icon string `js:"icon"`
	Position string `js:"position"`
	Timeout uint `js:"timeout"`
}

func QuasarGetQuasar() *Quasar {
	q := js.Global.Get("Quasar")
	return &Quasar{Object: q}
}

func QuasarNotify(notification *QuasarNotification) {
	/*
	println("Quasar Notify")
	println("Quasar:", GlobalQuasar)
	println("Quasar global get:", QuasarGetQuasar().Plugins)
	for name, plugin := range GlobalQuasar.Plugins {
		println(name,plugin)
	}
	*/
	GlobalQuasar.Plugins["Notify"].Call("create", notification)
}

