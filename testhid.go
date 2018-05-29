package main

import(
	"./hid"

	"log"
	"fmt"
)

func Test() {
	filepath := "/dev/hidg0"

	kbdRepEmpty := hid.NewKeyboardOutReport(0)
	kbdRep_a := hid.NewKeyboardOutReport(0, hid.HID_KEY_A)
	kbdRep_A := hid.NewKeyboardOutReport(hid.HID_MOD_KEY_LEFT_SHIFT, hid.HID_KEY_A)

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


func main() {
	/*
	//Define test keyboard map
	mapDeASCII := hid.HIDKeyboardLanguageMap{
		Name: "DE",
		Description: "German ASCII to USB Keyboard report mapping",
		Mapping: map[string][]hid.KeyboardOutReport{},
	}
	mapDeASCII.Mapping["c"] = []hid.KeyboardOutReport{
		hid.NewKeyboardOutReport(0, hid.HID_KEY_C),
	}
	mapDeASCII.Mapping["C"] = []hid.KeyboardOutReport{
		hid.NewKeyboardOutReport(hid.HID_MOD_KEY_LEFT_SHIFT, hid.HID_KEY_C),
	}

	//Store map to file
	err := mapDeASCII.StoreToFile("/tmp/DE_ASCII.json")
	if err != nil { log.Fatal(err)}

	testmap, err := hid.LoadKeyboardLanguageMapFromFile("keymaps/DE_ASCII.json")
	if err != nil { log.Fatal(err)}
	fmt.Println(testmap)
	*/

	keyboard := hid.HIDKeyboard{}
	keyboard.DevicePath = "/dev/hidg0"
	keyboard.KeyDelay = 100
	keyboard.KeyDelayJitter = 200
	keyboard.LoadLanguageMapFromFile("keymaps/DE_ASCII.json")
	fmt.Printf("Available language maps:\n%v\n",keyboard.ListLanguageMapNames())

	err := keyboard.SetActiveLanguageMap("DE")
	if err != nil { fmt.Println(err)}

	ascii := " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	special := "§°üÜöÖäÄµ€ß¹²³⁴⁵⁶⁷⁸⁹⁰¼½¬„“¢«»æſðđŋħĸł’¶ŧ←↓→øþ"
	err = keyboard.SendString("Test:" + ascii + "\t" + special)
	if err != nil { fmt.Println(err)}
}