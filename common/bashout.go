package common

// Functions/methods to move execution to bash (should only be used for things note yet implemented in go part of
// P4wnP1 or startup scripts)

import (
	"fmt"
	"os/exec"
)

type LogWriter struct {
	Prefix string
}

func (lw LogWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("%s: %s\n", lw.Prefix, string(p))
	return len(p),nil
}

func RunBashScript(scriptPath string) (err error) {
	cmd := exec.Command("/bin/bash", scriptPath)
	wStdout := LogWriter{scriptPath}
	wStderr := LogWriter{scriptPath + " error"}
	cmd.Stdout = wStdout
	cmd.Stderr = wStderr
	err = cmd.Start()
	if err != nil { return }
	return cmd.Wait()
}
