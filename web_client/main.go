package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "../proto/gopherjs"
	dom "honnef.co/go/js/dom"
	"github.com/oskca/gopherjs-vue"
	"github.com/gopherjs/gopherjs/js"
)

var (
	document   = dom.GetWindow().Document().(dom.HTMLDocument)
	serverAddr = GetBaseURL()
	Client     = pb.NewP4WNP1Client(
		serverAddr + ":80",
	)
	GS *pb.GadgetSettings
)

func GetBaseURL() string {
	document := js.Global.Get("window").Get("document")
	location := document.Get("location")
	port := location.Get("port").String()
	url := location.Get("protocol").String() + "//" + location.Get("hostname").String()
	if len(port) > 0 {
		url = url + ":" + port
	}
	return url
}

func main() {
	println(GetBaseURL())


	fmt.Printf("Address %v\n", strings.TrimSuffix(document.BaseURI(), "/"))
	fmt.Printf("Client %v\n", Client)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	gs, err := Client.GetDeployedGadgetSetting(ctx, &pb.Empty{})
	if err == nil {
		//export Gadget setting
		js.Global.Set("gs", gs)
		GS = gs
	} else {
		fmt.Printf("Error rpc call: %v\n", err)
	}


	vue.NewComponent(New, template).Register("usb-settings")
	vm := vue.New("#app", new(controller))
	js.Global.Set("vm", vm)
	println("vm:", vm)
}
