package common

import "os"

func WriteFile(path string, mustNotExist bool, append bool, data []byte, perm os.FileMode) (error) {
	flag := os.O_CREATE | os.O_WRONLY
	if mustNotExist { flag |= os.O_EXCL }
	if append { flag |= os.O_APPEND } else { flag |= os.O_TRUNC }
	f, err := os.OpenFile(path, flag, perm)
	f.Stat()
	if err != nil { return err }
	defer f.Close()
	_,err = f.Write(data)
	return err
}

func ReadFile(path string, start int64, chunk []byte, perm os.FileMode) (n int, err error) {
	flag := os.O_RDONLY
	f, err := os.OpenFile(path, flag, perm)
	if err != nil { return 0,err }
	defer f.Close()
	return f.ReadAt(chunk, start)
}
