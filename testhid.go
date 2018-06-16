package main

import(
	"./hid"

	"log"
	"fmt"
	"time"
	"math"
	"io/ioutil"
)

var (
	StringAscii = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	StringSpecial1 = "§°üÜöÖäÄµ€ß¹²³⁴⁵⁶⁷⁸⁹⁰¼½¬„“¢«»æſðđŋħĸł’¶ŧ←↓→øþ"

)

func TestMapCreation() {
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

	/*
	testmap, err := hid.loadKeyboardLanguageMapFromFile("keymaps/DE_ASCII.json")
	if err != nil { log.Fatal(err)}
	fmt.Println(testmap)
	*/
}

func TestComboPress(hidCtl *hid.HIDController) {
	testcombos := []string {"SHIFT 1", "ENTER", "ALT TAB", "ALT TABULATOR", "  WIN ", "GUI "}
	for _,comboStr := range testcombos {
		fmt.Printf("Pressing combo '%s'\n", comboStr)
		err := hidCtl.Keyboard.StringToPressKeyCombo(comboStr)
		if err == nil {
			fmt.Printf("... '%s' pressed sleeping 2s\n", comboStr)
			time.Sleep(2000 * time.Millisecond)
		} else {
			fmt.Printf("Error pressing combo '%s': %v\n", comboStr, err)
		}
	}
}

func TestLEDTriggers(hidCtl *hid.HIDController) {
	fmt.Println("Initial sleep to test if we capture LED state changes from the past, as soon as we start waiting (needed at boot)")
	time.Sleep(3 * time.Second)


	//Test repeat trigger on any LED
	fmt.Println("Waiting for any repeated LED state change (5 times frequently), wait timeout after 20 seconds...")
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
}

func TestMultiLEDTrigges(hidCtl *hid.HIDController, triggerMask byte) {
	//Test repeat trigger on given LED
	fmt.Printf("Waiting for repeated LED state change (5 times frequently) of mask %v, wait timeout after 20 seconds...\n", triggerMask)
	trigger, err := hidCtl.Keyboard.WaitLEDStateChangeRepeated(triggerMask, 5, time.Millisecond*500, 20*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}
}

func TestConcurrentLEDTriggers(hidCtl *hid.HIDController) {
	go TestMultiLEDTrigges(hidCtl, hid.MaskNumLock)
	go TestMultiLEDTrigges(hidCtl, hid.MaskCapsLock)
	go TestMultiLEDTrigges(hidCtl, hid.MaskCapsLock | hid.MaskScrollLock)
	time.Sleep(2*time.Second)
	go TestMultiLEDTrigges(hidCtl, hid.MaskAny)
	TestMultiLEDTrigges(hidCtl, hid.MaskKana)

}


func TestStringTyping(hidCtl *hid.HIDController) {
	fmt.Println("Typing:")
	fmt.Println(StringAscii)
	//err := hidCtl.Keyboard.StringToPressKeySequence("Test:" + StringAscii + "\t" + StringSpecial1)
	err := hidCtl.Keyboard.StringToPressKeySequence(StringAscii)
	if err != nil { fmt.Println(err)}
}

func TestCombinedScript(hidCtl *hid.HIDController) (err error) {

	testcript := `
		
		console.log("HID Script test for P4wnP1 rework");	//Print to internal console
		for (var i = 0; i<5; i++) {
			move(128, 0);
			delay(500);
			move(0, -100.1);
			delay(500);
			move(-100, 0);
			delay(500);
			move(0, 100);
			delay(500);
		}

		console.log("HID Script test for P4wnP1 rework");	//Print to internal console
		for (var i = 0; i<5; i++) {
			moveTo(0.0, 0.0);
			delay(500);
			moveTo(0.8, 0.0);
			delay(500);
			moveTo(0.8, 0.8);
			delay(500);
			moveTo(0.8, 0.8);
			delay(500);
		}
		waitLED(ANY)
		
		layout("US"); 										//Switch to US keyboard layout
		type("Some ASCII test text QWERTZ\n")				//Type text to target ('\n' translates to RETURN key)		

		delay(200); 										//sleep 200 milliseconds

		//waitLEDRepeat test
		var mask = NUM | SCROLL;
		var repeatCount = 5;
		var repeatIntervalMs = 800;
		var timeout = 20;
		
		//result = waitLEDRepeat(mask, repeatCount, repeatIntervalMs, timeout)
		result = waitLEDRepeat(mask, repeatCount)
		console.log("Result: " + JSON.stringify(result));	//Log result object as JSON to internal console

		waitLED(NUM | SCROLL, 2); 							//Wait for NumLock or ScrollLock LED change, abort after 2 seconds

		layout("DE"); 										//Switch to German keyboard layout
		type("Non ASCII: üÜöÖäÄ");							//Type non ASCII
		press("ENTER");										//Introduce linebreak by pressing RETURN directly
		press("RETURN");									//Alias

		counter = 4;										//set a var ...
		type("Pressing <ALT>+<TAB> "+ counter +" times\n");	//... and type it, along with a string
		for (var i=0; i<counter; i++) {
			press("ALT TAB");
			delay(500)
		}

		//Test LED change based branching
		result = waitLED(NUM | CAPS);						//Wait for change on NUM or CAPS LED only, without timeout, store result
		console.log("Result: " + JSON.stringify(result));	//Log result object as JSON to internal console
		if (result.NUM) {									//Branch depending on result of LED change 
			type("NUMLock LED changed\n");
		} else {
			type("Seems CAPSLock LED changed\n");			
		}
	`

	_,err = hidCtl.RunScript(testcript)
	if err != nil {panic(err)}

	return
}


func TestMouseNoScript(hidCtl *hid.HIDController) (err error) {
	hidCtl.Mouse.MoveStepped(100,0)
	hidCtl.Mouse.MoveStepped(0,-100)
	hidCtl.Mouse.MoveStepped(0,100)

	time.Sleep(2*time.Second)

	hidCtl.Mouse.SetButtons(true, false, false)
	for alpha := 0.0; alpha < 8*math.Pi; alpha+=(math.Pi/180) {
		cos := int16(math.Cos(6.0*alpha) * 5)
		sin := int16(math.Sin(alpha) * 5)

		hidCtl.Mouse.MoveStepped(sin,cos)
	}
	hidCtl.Mouse.SetButtons(false, false, false)

	return nil
}

func TestMouseCircle(hidCtl *hid.HIDController) {
	scriptMouse := `
		//circular mouse movement with rotating vector
		turns = 2
		degree = Math.PI/180.0
		scale = 4
		for (var alpha = 0; alpha < 2 * Math.PI * turns; alpha += degree) {
			vecx = Math.cos(alpha) * scale
			vecy = Math.sin(alpha) * scale

			moveStepped(vecx, vecy);
		}
	`

	_,err := hidCtl.RunScript(scriptMouse)
	if err != nil { panic(err)}
}

func main() {
	/*
	*/



	/*
	for x:=0; x<50; x++ {
		err := mouse.MoveTo(-12,int16(x))
		if err != nil { panic(err) }
		time.Sleep(100 * time.Millisecond)
	}
	*/


	hidCtl, err := hid.NewHIDController("/dev/hidg0", "keymaps", "/dev/hidg1")

//	hidCtl.StartScriptAsBackgroundJob("waitLED(ANY)")
//	time.Sleep(1*time.Second)
//	hidCtl, err = hid.NewHIDController("/dev/hidg0", "keymaps", "/dev/hidg1")

	if err != nil {panic(err)}
	hidCtl.Keyboard.KeyDelay = 100
	//	hidCtl.Keyboard.KeyDelayJitter = 200


	fmt.Printf("Available language maps:\n%v\n",hidCtl.Keyboard.ListLanguageMapNames())
	err = hidCtl.Keyboard.SetActiveLanguageMap("DE") //first loaded language map is set by default
	if err != nil { fmt.Println(err)}
	fmt.Printf("Chosen keyboard language mapping '%s'\n", hidCtl.Keyboard.ActiveLanguageLayout.Name)

	/* tests */



	//TestComboPress(hidCtl)
	//TestLEDTriggers(hidCtl)
	//TestStringTyping(hidCtl)
	//TestConcurrentLEDTriggers(hidCtl)
	//TestMouseNoScript(hidCtl)
	//TestCombinedScript(hidCtl)
	//TestMouseCircle(hidCtl)


	/*
	go func() {
		time.Sleep(3*time.Second)
		fmt.Printf("=======================\nClosing LED watcher\n======================\n")
		hidCtl.Keyboard.Close()
	}()

	scriptConcurrent := `
		console.log("Starting script with job ID: " + JID);
		delay(1000);
		console.log("Script with job ID: " + JID + " finished");
		res = waitLEDRepeat(ANY,5);
		console.log("****Finished Waiting for LED " + JSON.stringify(res));
	`


	for i:=0; i<25; i++ {
		fmt.Printf("Starting background script, run : %d\n",i)
		_,_,errBG := hidCtl.StartScriptAsBackgroundJob(scriptConcurrent)
		if errBG != nil { fmt.Printf("Error: %v\n", errBG)}
		time.Sleep(500 * time.Millisecond)

	}


	time.Sleep(2*time.Second)
	_,err = hidCtl.RunScript(scriptConcurrent)
	if err != nil {
		panic(err)
	}

	*/

	//try to load script file
	filepath := "./hidtest1.js"
	if scriptFile, err := ioutil.ReadFile(filepath); err != nil {
		log.Printf("Couldn't load HIDScript testfile: %s\n", filepath)
	} else {
		_,err = hidCtl.RunScript(string(scriptFile))
		if err != nil { panic(err)}
	}

	/*

	script := `
		for (i=0; i<10; i++) {
			press("CAPS")
			delay(500)
		}

		type(" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz{|}~");
		type("\n")
		type("Waiting 500ms ...\n");
		delay(500)
		type("... done\n");
		type("§°üÜöÖäÄµ€ß¹²³⁴⁵⁶⁷⁸⁹⁰¼½¬„“¢«»æſðđŋħĸł’¶ŧ←↓→øþ");
		type("\n")
		
		console.log("Log message from JS"); 
	`

	hidCtl.StartScriptAsBackgroundJob(script)

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

	id, avm,_ := hidCtl.StartScriptAsBackgroundJob(script2)
	fmt.Printf("Satrted ASYNC VM %d ...\n", id)


	time.Sleep(1500 * time.Millisecond)
	id2 ,avm2, _ := hidCtl.StartScriptAsBackgroundJob(script3)
	fmt.Printf("Slept 1500 ms, started VM %d ...\n", id2)
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id, avm.IsWorking())
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id2, avm2.IsWorking())

	time.Sleep(1500 * time.Millisecond)
	fmt.Printf("Slept 1500 ms, cancelling VM %d and start waiting for VM %d...\n", id2, id)
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id, avm.IsWorking())
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id2, avm2.IsWorking())
	fmt.Printf("List of currently working VMs: %v\n", hidCtl.GetRunningBackgroundJobs())
	//avm2.Cancel()
	hidCtl.CancelBackgroundJob(id2)

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
	hidCtl.CancelAllBackgroundJobs()

	val,err = avm.WaitResult()

	if err != nil {log.Fatal(err)}
	fmt.Printf("Running Script ASYNC finished with result: %v ...\n", val)
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id, avm.IsWorking())
	fmt.Printf("AsyncVM state isWorking of id %d: %v\n", id2, avm2.IsWorking())

	*/
}