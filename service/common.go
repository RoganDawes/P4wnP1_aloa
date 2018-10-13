package service

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
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

func ListFilesOfFolder(folderPath string, allowedExtensions ...string) (res []string, err error) {
	// assure all allowed extensions are prepended with a dot and converted to lower case
	for i,e := range allowedExtensions {
		if len(e) == 0 { continue }
		if []rune(e)[0] != '.' {
			allowedExtensions[i] = "." + allowedExtensions[i]
		}
		allowedExtensions[i] = strings.ToLower(allowedExtensions[i])
	}

	fcontent,err := ioutil.ReadDir(folderPath)
	if err != nil { return res,err }

	for _,fitem := range fcontent {
		if !fitem.IsDir() {
			extensionValid := false
			// seems to be a file
			if len(allowedExtensions) > 0 {
				//check if extension is valid
				itemExt := strings.ToLower(filepath.Ext(fitem.Name()))
				Inner:
				for _,validExt := range allowedExtensions {
					if validExt == itemExt {
						extensionValid = true
						break Inner
					}
				}
			} else {
				extensionValid = true
			}

			if extensionValid {
				res = append(res, fitem.Name())
			}

		}
	}

	return
}