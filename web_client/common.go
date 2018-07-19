package main

import (
	"github.com/gopherjs/gopherjs/js"
	"crypto/md5"
	"encoding/hex"
	"time"
	pb "../proto/gopherjs"
	"context"
)

func O() *js.Object {
	return js.Global.Get("Object").New()
}

func Alert(in interface{}) {
	js.Global.Call("alert", in)
}

//Converts string to MD5 hex representation, avoid using fmt package
func StringToMD5(input string) string {
	sum := md5.Sum([]byte(input))
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum[:])
	return string(dst)
}

func UploadHIDScript(filename string, content string) (err error) {
	ctx,cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//ToDO: filename could be generated here and returned as result

	//ToDo: split upload data into chunks

	client := pb.NewP4WNP1Client(serverAddr)
	_,err = client.FSWriteFile(
		ctx,
		&pb.WriteFileRequest{
			Data: []byte(content),
			Append: false,
			Path: "/tmp/" + filename,
			MustNotExist: false,
		},
	)

	return err
}

func RunHIDScript(filename string, timeoutSeconds uint32) (job *pb.HIDScriptJob,err error) {
	ctx,cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//ToDo: split upload data into chunks

	client := pb.NewP4WNP1Client(serverAddr)
	return client.HIDRunScriptJob(
		ctx,
		&pb.HIDScriptRequest{
			ScriptPath: "/tmp/" + filename,
			TimeoutSeconds: timeoutSeconds,
		},
	)

}