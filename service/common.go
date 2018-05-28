package service

import (
	"os/exec"
	"strings"
)

func binaryAvailable(binname string) bool {
	cmd := exec.Command("which", binname)
	out,err := cmd.CombinedOutput()
	if err != nil { return false}
	if len(out) == 0 { return false }

	if strings.Contains(string(out), binname) {
		return true
	}
	return false
}
