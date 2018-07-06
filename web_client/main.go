package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "../proto/gopherjs"
	dom "honnef.co/go/js/dom"
)

var (
	document   = dom.GetWindow().Document().(dom.HTMLDocument)
	serverAddr = "http://raspberrypi.local"
)

func main() {
	fmt.Println("Hello")

	client := pb.NewP4WNP1Client(
		"http://raspberrypi.local:80",
	)
	fmt.Printf("Address %v\n", strings.TrimSuffix(document.BaseURI(), "/"))
	fmt.Printf("Client %v\n", client)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	gs, err := client.GetDeployedGadgetSetting(ctx, &pb.Empty{})
	if err == nil {
		str:=fmt.Sprintf("Gs: %+v\n", gs)
		fmt.Println(str)
		div_cont:= dom.GetWindow().Document().GetElementByID("content").(*dom.HTMLDivElement)
		new_div := dom.GetWindow().Document().CreateElement("div").(*dom.HTMLDivElement)
		new_div.SetTextContent(fmt.Sprintf("Result of GetDeployedGadgetSetting gRPC-web call:\n%s ",str))
		div_cont.AppendChild(new_div)
	} else {
		fmt.Printf("Error rpc call: %v\n", err)
	}

}
