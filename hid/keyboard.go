package hid

import (
	"errors"
	"io/ioutil"
	"os"
	"log"
)



type KeyboardOutReport struct {
	Modifiers byte
	//Reserved byte
	Keys [6]byte
}

func (rep KeyboardOutReport) Serialize() (out []byte) {
	out = []byte {
		rep.Modifiers,
		byte(0),
		rep.Keys[0],
		rep.Keys[1],
		rep.Keys[2],
		rep.Keys[3],
		rep.Keys[4],
		rep.Keys[5],
	}
	return
}

func (rep KeyboardOutReport) Deserialize(data []byte) (err error) {
	if len(data) != 8 {
		err = errors.New("Wrong data length, keyboard out report has to be 8 bytes in length")
	}
	rep = KeyboardOutReport{
		Modifiers: data[0],
		//data[1] should be empty, we ignore it
		Keys: [6]byte{
			data[2],
			data[3],
			data[4],
			data[5],
			data[6],
			data[7],
		},
	}
	return
}

func (rep KeyboardOutReport) WriteTo(filePath string) (err error) {
	return ioutil.WriteFile(filePath, rep.Serialize(), os.ModePerm) //Serialize Report and write to specified file
}



func NewKeyboardOutReport(modifiers byte, keys ...byte) (res KeyboardOutReport) {
	res = KeyboardOutReport{
		Keys: [6]byte {0, 0, 0, 0, 0, 0,},
	}
	res.Modifiers = modifiers
	for i, key := range keys {
		if i < 6 {
			res.Keys[i] = key
		}
	}
	return
}

func Test() {
	filepath := "/dev/hidg0"
	kbdRepEmpty := NewKeyboardOutReport(0)
	kbdRep_a := NewKeyboardOutReport(0, HID_KEY_A)
	kbdRep_A := NewKeyboardOutReport(HID_MOD_KEY_LEFT_SHIFT, HID_KEY_A)

	err := kbdRepEmpty.WriteTo(filepath)
	if err != nil {log.Fatal(err)}
	err = kbdRep_a.WriteTo(filepath)
	if err != nil {log.Fatal(err)}
	err = kbdRepEmpty.WriteTo(filepath)
	if err != nil {log.Fatal(err)}
	err = kbdRep_A.WriteTo(filepath)
	if err != nil {log.Fatal(err)}
	err = kbdRepEmpty.WriteTo(filepath)
	if err != nil {log.Fatal(err)}
	err = kbdRep_A.WriteTo(filepath)
	if err != nil {log.Fatal(err)}
	err = kbdRepEmpty.WriteTo(filepath)
	if err != nil {log.Fatal(err)}
}
