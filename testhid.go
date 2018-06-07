package main

import(
	"./hid"

	"log"
	"fmt"
	"time"
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

	hidCtl, err := hid.NewHIDController("/dev/hidg0", "keymaps", "")

	fmt.Println("Initial sleep to test if we capture LED state changes from the past, as soon as we start waiting (needed at boot)")
	time.Sleep(3 * time.Second)

	//ToDo: Test multiple waits in separate goroutines

	//Test repeat single trigger on any LED
	fmt.Println("Waiting for any repeted LED state change (5 times frequently), wait timeout after 20 seconds...")
	trigger, err := hidCtl.Keyboard.WaitLEDStateChangeRepeated(hid.MaskAny, 5, time.Millisecond*500, 20*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}

	//Test single trigger on any LED
	fmt.Println("Waiting for any LED single state change, timeout after 15 seconds")
	trigger, err = hidCtl.Keyboard.WaitLEDStateChange(hid.MaskAny, 15*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}




	//Test single trigger on NUMLOCK LED (ignore CAPSLOCK, SCROLLLOCK etc.)
	fmt.Println("Waiting for NUMLOCK LED state change, timeout after 15 seconds")
	trigger, err = hidCtl.Keyboard.WaitLEDStateChange(hid.MaskNumLock, 15*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}

	//Test single trigger on NUMLOCK LED (ignore CAPSLOCK, SCROLLLOCK etc.)
	fmt.Println("Waiting for CAPSLOCK LED state change for 15 seconds")
	trigger, err = hidCtl.Keyboard.WaitLEDStateChange(hid.MaskCapsLock, 15*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}


	hidCtl.Keyboard.KeyDelay = 100
//	hidCtl.Keyboard.KeyDelayJitter = 200
	fmt.Printf("Available language maps:\n%v\n",hidCtl.Keyboard.ListLanguageMapNames())

	err = hidCtl.Keyboard.SetActiveLanguageMap("DE")
	if err != nil { fmt.Println(err)}

//	ascii := " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
//	special := "§°üÜöÖäÄµ€ß¹²³⁴⁵⁶⁷⁸⁹⁰¼½¬„“¢«»æſðđŋħĸł’¶ŧ←↓→øþ"
//	err = keyboard.SendString("Test:" + ascii + "\t" + special)
	if err != nil { fmt.Println(err)}

	script := `
		kString(" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~");
		kString("\n")
		kString("Waiting 500ms ...\n");
		delay(500)
		kString("... done\n");
		kString("§°üÜöÖäÄµ€ß¹²³⁴⁵⁶⁷⁸⁹⁰¼½¬„“¢«»æſðđŋħĸł’¶ŧ←↓→øþ");
		kString("\n")
		
		console.log("Log message from JS"); 
	`

	hidCtl.RunScriptAsync(script)

	fmt.Println("Running script ...")
	val,err := hidCtl.RunScript(script)
	if err != nil {log.Fatal(err)}
	fmt.Printf("Running script finished with result: %v ...\n", val)

	script2 := `
		for (i=0; i<10; i++)
		{
			console.log("Run " + i + ":");
			console.log("JS sleeping 1000ms");
			delay(1000);
		}
	`
	script3 := `
		for (i=0; i<30; i++)
		{
			console.log("JS script 3 Run " + i);
			delay(500);
		}
	`

	id, avm,_ := hidCtl.RunScriptAsync(script2)
	fmt.Printf("Satrted ASYNC VM %d ...\n", id)


	time.Sleep(1500 * time.Millisecond)
	id2 ,avm2, _ := hidCtl.RunScriptAsync(script3)
	fmt.Printf("Slept 1500 ms, started VM %d ...\n", id2)
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id, avm.IsWorking())
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id2, avm2.IsWorking())

	time.Sleep(1500 * time.Millisecond)
	fmt.Printf("Slept 1500 ms, cancelling VM %d and start waiting for VM %d...\n", id2, id)
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id, avm.IsWorking())
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id2, avm2.IsWorking())
	fmt.Printf("List of currently working VMs: %v\n", hidCtl.CurrentlyWorkingVmIDs())
	//avm2.Cancel()
	hidCtl.CancelAsyncScript(id2)

	//Waiting for canceled script
	_,err = avm2.WaitResult()
	if err != nil {
		fmt.Printf("Script 2 resulted in error: %v\n", err)
	}


	//Try to reuse avm2
	fmt.Println("Restarting script 3 on second VM")
	time.Sleep(2000 * time.Millisecond)
	avm2.RunAsync(script3)

	//Cancel all
	hidCtl.CancelAllVMs()

	val,err = avm.WaitResult()

	if err != nil {log.Fatal(err)}
	fmt.Printf("Running Script ASYNC finished with result: %v ...\n", val)
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id, avm.IsWorking())
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id2, avm2.IsWorking())


}