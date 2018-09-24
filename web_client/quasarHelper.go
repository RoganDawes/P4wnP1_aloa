// +build js

package main

import "github.com/gopherjs/gopherjs/js"

var GlobalQuasar = QuasarGetQuasar()

type QuasarDialogPosition string

const (
	QUASAR_NOTIFICATION_TYPE_POSITIVE = "positive"
	QUASAR_NOTIFICATION_TYPE_NEGATIVE = "negative"
	QUASAR_NOTIFICATION_TYPE_WARNING  = "warning"
	QUASAR_NOTIFICATION_TYPE_INFO     = "info"

	QUASAR_NOTIFICATION_POSITION_TOP          = "top"
	QUASAR_NOTIFICATION_POSITION_TOP_LEFT     = "top-left"
	QUASAR_NOTIFICATION_POSITION_TOP_RIGHT    = "top-right"
	QUASAR_NOTIFICATION_POSITION_LEFT         = "left"
	QUASAR_NOTIFICATION_POSITION_CENTER       = "center"
	QUASAR_NOTIFICATION_POSITION_RIGHT        = "right"
	QUASAR_NOTIFICATION_POSITION_BOTTOM       = "bottom"
	QUASAR_NOTIFICATION_POSITION_BOTTOM_LEFT  = "bottom-left"
	QUASAR_NOTIFICATION_POSITION_BOTTOM_RIGHT = "bottom-right"

	QUASAR_DIALOG_POSITION_TOP    = QuasarDialogPosition("top")
	QUASAR_DIALOG_POSITION_LEFT   = QuasarDialogPosition("left")
	QUASAR_DIALOG_POSITION_RIGHT  = QuasarDialogPosition("right")
	QUASAR_DIALOG_POSITION_BOTTOM = QuasarDialogPosition("bottom")

	QUASAR_NOTIFICATION_TIMEOUT = 5000
)

type Quasar struct {
	*js.Object
	Version string                `js:"version"`
	Theme   string                `js:"theme"`
	Plugins map[string]*js.Object `js:"plugins"`
}

type QuasarNotification struct {
	*js.Object
	Message   string `js:"message"`
	Detail    string `js:"detail"`
	Type      string `js:"type"`
	Color     string `js:"color"`
	TextColor string `js:"textColor"`
	Icon      string `js:"icon"`
	Position  string `js:"position"`
	Timeout   uint   `js:"timeout"`
}

type QuasarDialogType struct {
	*js.Object
	Title             string               `js:"title"`
	Message           string               `js:"message"`
	Ok                bool                 `js:"ok"`
	Cancel            bool                 `js:"cancel"`
	PreventClose      bool                 `js:"preventClose"`
	NoBackdropDismiss bool                 `js:"noBackdropDismiss"`
	NoEscDismiss      bool                 `js:"noEscDismiss"`
	StackButtons      bool                 `js:"stackButtons"`
	Position          QuasarDialogPosition `js:"position"`

	Detail    string `js:"detail"`
	Type      string `js:"type"`
	Color     string `js:"color"`
	TextColor string `js:"textColor"`
	Icon      string `js:"icon"`
	Timeout   uint   `js:"timeout"`
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

func QuasarDialog(notification *QuasarDialogType) {
	/*
	println("Quasar Notify")
	println("Quasar:", GlobalQuasar)
	println("Quasar global get:", QuasarGetQuasar().Plugins)
	for name, plugin := range GlobalQuasar.Plugins {
		println(name,plugin)
	}
	*/
	GlobalQuasar.Plugins["Dialog"].Call("create", notification)
}

func QuasarNotifyError(errorMessage string, messageDetails string, position string) {
	notification := &QuasarNotification{Object: O()}
	notification.Message = errorMessage
	notification.Detail = messageDetails
	notification.Position = position
	notification.Type = QUASAR_NOTIFICATION_TYPE_NEGATIVE
	notification.Timeout = QUASAR_NOTIFICATION_TIMEOUT
	QuasarNotify(notification)
}

func QuasarNotifySuccess(message string, detailMessage string, position string) {
	notification := &QuasarNotification{Object: O()}
	notification.Message = message
	notification.Detail = detailMessage
	notification.Position = position
	notification.Type = QUASAR_NOTIFICATION_TYPE_POSITIVE
	notification.Timeout = QUASAR_NOTIFICATION_TIMEOUT
	QuasarNotify(notification)
}
