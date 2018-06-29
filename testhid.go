package main

import(
	"./hid"

	"log"
	"fmt"
	"time"
	"math"
	"os"
	"runtime/trace"
	"os/signal"
	"syscall"
	_ "net/http/pprof"
	"net/http"
	"runtime"
	"context"
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
	trigger, err := hidCtl.Keyboard.WaitLEDStateChangeRepeated(context.Background(), hid.MaskAny, 5, time.Millisecond*500, 20*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}

	//Test single trigger on any LED
	fmt.Println("Waiting for any LED single state change, timeout after 15 seconds")
	trigger, err = hidCtl.Keyboard.WaitLEDStateChange(context.Background(),hid.MaskAny, 15*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}




	//Test single trigger on NUMLOCK LED (ignore CAPSLOCK, SCROLLLOCK etc.)
	fmt.Println("Waiting for NUMLOCK LED state change, timeout after 15 seconds")
	trigger, err = hidCtl.Keyboard.WaitLEDStateChange(context.Background(),hid.MaskNumLock, 15*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}

	//Test single trigger on NUMLOCK LED (ignore CAPSLOCK, SCROLLLOCK etc.)
	fmt.Println("Waiting for CAPSLOCK LED state change for 15 seconds")
	trigger, err = hidCtl.Keyboard.WaitLEDStateChange(context.Background(), hid.MaskCapsLock, 15*time.Second)
	if err != nil {
		fmt.Printf("Waiting aborted with error: %v\n", err)
	} else {
		fmt.Printf("Triggered by %+v\n", trigger)
	}
}

func TestMultiLEDTrigges(hidCtl *hid.HIDController, triggerMask byte) {
	//Test repeat trigger on given LED
	fmt.Printf("Waiting for repeated LED state change (5 times frequently) of mask %v, wait timeout after 20 seconds...\n", triggerMask)
	trigger, err := hidCtl.Keyboard.WaitLEDStateChangeRepeated(context.Background(),triggerMask, 5, time.Millisecond*500, 20*time.Second)
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

	_,err = hidCtl.RunScript(context.Background(),testcript)
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

	_,err := hidCtl.RunScript(context.Background(),scriptMouse)
	if err != nil { panic(err)}
}


// To profile for memory leaks and test clean cancellation of already running scripts on controller re-init
func TestControllerReInit() {
	//Test for memory leaks
	hidCtlTests := make([]*hid.HIDController,0)
	for i:=0; i<10;i++ {
		//create new controller
		fmt.Printf("****Creating HIDController %d\n", i)
		hidCtlTest,_ := hid.NewHIDController(context.Background(),"/dev/hidg0", "keymaps", "/dev/hidg1")

		//run script which utilizes LED read
		fmt.Printf("****Starting async LED reading script for HIDController %d\n", i)
		//script := "waitLEDRepeat(ANY);"
		script := "console.log('...started');delay(3000);console.log('...ended');"
		ctx := context.Background()
		for i:=0;i<4;i++ {
			job,err := hidCtlTest.StartScriptAsBackgroundJob(ctx, script)
			if err != nil {
				fmt.Printf("Error starting new job: %v\n",err)
			} else {
				fmt.Printf("New job started: %+v\n",job)
			}
		}


		time.Sleep(time.Second)

		//add to slice
		hidCtlTests = append(hidCtlTests, hidCtlTest)


	}
	hidCtlTests = make([]*hid.HIDController,0)
	runtime.GC()
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()



	//TestControllerReInit()



	hidCtl, err := hid.NewHIDController(context.Background(),"/dev/hidg0", "keymaps", "/dev/hidg1")


	if err != nil {panic(err)}
	hidCtl.Keyboard.KeyDelay = 100
	//	hidCtl.Keyboard.KeyDelayJitter = 200


	fmt.Printf("Available language maps:\n%v\n",hidCtl.Keyboard.ListLanguageMapNames())
	err = hidCtl.Keyboard.SetActiveLanguageMap("DE") //first loaded language map is set by default
	if err != nil { fmt.Println(err)}
	fmt.Printf("Chosen keyboard language mapping '%s'\n", hidCtl.Keyboard.ActiveLanguageLayout.Name)

	// tests
	//TestComboPress(hidCtl)
	//TestLEDTriggers(hidCtl)
	//TestStringTyping(hidCtl)
	//TestConcurrentLEDTriggers(hidCtl)
	//TestMouseNoScript(hidCtl)
	//TestCombinedScript(hidCtl)
	//TestMouseCircle(hidCtl)


	//add bg jobs waiting for LED
	jobList := make([]int,0)
	fmt.Println("Adding sleeping jobs with 5 seconds timeout context")
	ctxT,_ := context.WithTimeout(context.Background(), time.Second * 2)
	//script := "console.log('START ' + JID + ' on VM ' + VMID);delay(5000);console.log(JID + ' returned from 5s blocking delay');"
	script := "console.log('START ' + JID + ' on VM ' + VMID);waitLEDRepeat(ANY,5000);console.log(JID + ' returned from 5s blocking delay');"
	startTime := time.Now()
	for i:=1; i<4; i++ {
		job,err := hidCtl.StartScriptAsBackgroundJob(ctxT,script)
		if err != nil {
			fmt.Printf("Failed adding background job: %v\n", err)
		} else {
			// ad job to slice
			jobList = append(jobList, job.Id)
		}
	}
	//Wait for all jobs to finish
	fmt.Printf("Waiting for Job results for IDs: %+v\n", jobList)
	for _,jid := range jobList {
		job,err := hidCtl.GetBackgroundJobByID(jid)
		if err != nil {
			fmt.Printf("Job with ID %d not found, skipping...\n", jid)
			continue
		} else {
			fmt.Printf("Waiting for finish of job with ID %d \n", jid)
			jRes,jErr := hidCtl.WaitBackgroundJobResult(job)
			fmt.Printf("JID: %d, Result: %+v, Err: %v\n", jid, jRes, jErr)
		}
	}
	fmt.Printf("All results received after %v\n", time.Since(startTime))







	//try to load script file
	filepath := "./hidtest1.js"
	if scriptFile, err := ioutil.ReadFile(filepath); err != nil {
		log.Printf("Couldn't load HIDScript testfile: %s\n", filepath)
	} else {
		_,err = hidCtl.RunScript(context.Background(),string(scriptFile))
		if err != nil { panic(err)}
	}



	go http.ListenAndServe(":8080", nil)

	//use a channel to wait for SIGTERM or SIGINT
	fmt.Println("Waiting for keyboard interrupt")
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Printf("Signal (%v) received, closing ...\n", s)
	return
	//log.Fatalf("Signal (%v) received, closing \"Let Me In\" rebind DNS server\n", s)
}