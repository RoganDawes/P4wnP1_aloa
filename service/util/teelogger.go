package util

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"sync"
)

type TeeLogger struct {
	*sync.Mutex
	*log.Logger
	outs []io.Writer
	LogWriter io.Writer
}

type sublogger struct {
	*TeeLogger
}

/*
// struct to present an additional io.Writer, wrapping TeeLogger to ise its Print() method
func (sl sublogger) Write(p []byte) (n int, err error) {
	sl.TeeLogger.Print(string(p))
	return len(p), nil
}
*/

// struct to present an additional io.Writer, wrapping TeeLogger to ise its Print() method
func  (sl sublogger) Write(p []byte) (n int, err error) {
	//fmt.Printf("%s: %s", lw.Prefix, string(p))

	lineScanner := bufio.NewScanner(bytes.NewReader(p))
	lineScanner.Split(bufio.ScanLines)
	for lineScanner.Scan() {
		sl.TeeLogger.Print(string(lineScanner.Bytes()))
		//fmt.Printf("%s: %s\n", lw.Prefix, string(lineScanner.Bytes()))
	}


	return len(p),nil
}


func NewTeeLogger(addStdout bool) (res *TeeLogger) {
	res = &TeeLogger{
		Mutex: &sync.Mutex{},
		Logger: &log.Logger{},
		outs: make([]io.Writer,0),
	}

	if addStdout {
		res.AddOutput(os.Stdout)
	}
	// Create a sub-struct which presents a new writer used to prepend the PREFIX in front of written data (the TeeLogger's own io.Writer is used for tee'ing)
	res.LogWriter = sublogger{ TeeLogger: res }
	res.SetFlags(log.Ltime)
	res.SetOutput(res)
	log.Println()
	return res
}

func (tl *TeeLogger) AddOutput(out io.Writer) {
	tl.Lock()
	defer tl.Unlock()
	tl.outs = append(tl.outs, out)
}


func (tl *TeeLogger) Write(p []byte) (n int, err error) {
	outmsg := []byte(string(p))
	tl.Lock()
	defer tl.Unlock()

	for _,out := range tl.outs {
		for written:=0; written < len(outmsg); {
			len,err := out.Write(outmsg)
			if err != nil { return written, err }
			written += len
		}

	}

	return len(p), nil
}
