package hid

import (
	"sync"
	"os"
	"time"
	"fmt"
)


const (
	MaskNumLock    = 1 << 0
	MaskCapsLock   = 1 << 1
	MaskScrollLock = 1 << 2
	MaskCompose    = 1 << 3
	MaskKana       = 1 << 4
	MaskAny = MaskNumLock | MaskCapsLock | MaskScrollLock | MaskCompose | MaskKana
)


type HIDLEDState struct {
	NumLock bool
	CapsLock bool
	ScrollLock bool
	Compose bool
	Kana bool
}

type HIDKeyboardLEDStateWatcher struct {
	ledState *HIDLEDState
	listeners2 sync.Map
	listeners *listenersmap
	listenerNonZeroCondLock *sync.Mutex
	listenerNonZeroCond *sync.Cond
	hasInitialSate bool
}

type listenersmap struct {
	sync.Mutex
	m map[*HIDKeyboardLEDListener]*HIDKeyboardLEDListener
}

type HIDKeyboardLEDListener struct {
	ledWatcher          *HIDKeyboardLEDStateWatcher //the parent LEDWatcher, containing global ledState
	changedLeds         chan HIDLEDState //changedLeds represents the LEDs which change since last report as bitfield (MaskNumLock, MaskCapsLock ...)  the actual state has to be fetched from the respective field of the ledWatcher.ledState
	isMarkedForDeletion bool
}


func newHIDKeyboardLEDStateWatcher(devicePath string) (ledState *HIDKeyboardLEDStateWatcher,err error) {
	llock := &sync.Mutex{}
	ledState = &HIDKeyboardLEDStateWatcher{
		ledState: &HIDLEDState{},
		listeners: &listenersmap{m: make(map[*HIDKeyboardLEDListener]*HIDKeyboardLEDListener)},
		listenerNonZeroCondLock: llock,
		listenerNonZeroCond: &sync.Cond{L:llock},
	}

	//start go routine reading LED output reports from keyboard device
	//ToDo: this should happen only once

	go ledState.start(devicePath)

	return
}

func (watcher *HIDKeyboardLEDStateWatcher) CreateAndAddNewListener() (l *HIDKeyboardLEDListener) {
	l = &HIDKeyboardLEDListener{
		ledWatcher: watcher,
		changedLeds: make(chan HIDLEDState),
	}
	watcher.listeners.Lock()
	watcher.listeners.m[l] = l
	watcher.listenerNonZeroCondLock.Lock()
	watcher.listenerNonZeroCond.Broadcast()
	watcher.listenerNonZeroCondLock.Unlock()
	watcher.listeners.Unlock()
	return
}

func (watcher *HIDKeyboardLEDStateWatcher) removeListener(l *HIDKeyboardLEDListener) {
	l.isMarkedForDeletion = true //mark listener for deletion
	//clear remaining channel data to unlock producer (dispatchListeners loop)
L:
	for {
		select {
		case <-l.changedLeds:
			//do nothinng
		default:
			break L
		}

	}
	close(l.changedLeds) //close channel (ther should be no further read access)

}

func (watcher *HIDKeyboardLEDStateWatcher) start(devicePath string) {
	f, err := os.Open(devicePath)
	defer f.Close()
	if err != nil { panic(err) }
	buf := make([]byte, 1)

	//ToDo: implement cancel() and allow blocking read to be interrupted (select ??)
	for {
		n,err := f.Read(buf)
		if err != nil { panic(err) }
		for i:=0; i<n; i++ {
			watcher.dispatchListeners(buf[i]) //dispatchListeners implements the logic to remove listeners which are marked for remove and should be issued after every read of a single byte
		}
	}
}


func (watcher *HIDKeyboardLEDStateWatcher) dispatchListeners(state byte) {
	//log.Printf("New LED state %x\n", state)
	var nNum, nCaps, nScroll, nCompose, nKana bool

	if state &MaskNumLock > 0 { nNum = true }
	if state &MaskCapsLock > 0 { nCaps = true }
	if state &MaskScrollLock > 0 { nScroll = true }
	if state &MaskCompose > 0 { nCompose = true }
	if state &MaskKana > 0 { nKana = true }

	hasChanged := false
	ledsChanged := HIDLEDState{}
	if nNum != watcher.ledState.NumLock {
		watcher.ledState.NumLock = nNum
		ledsChanged.NumLock = true
		hasChanged = true
	}
	if nCaps != watcher.ledState.CapsLock {
		watcher.ledState.CapsLock = nCaps
		ledsChanged.CapsLock = true
		hasChanged = true
	}
	if nScroll != watcher.ledState.ScrollLock {
		watcher.ledState.ScrollLock = nScroll
		ledsChanged.ScrollLock = true
		hasChanged = true
	}
	if nCompose != watcher.ledState.Compose {
		watcher.ledState.Compose = nCompose
		ledsChanged.Compose = true
		hasChanged = true
	}
	if nKana != watcher.ledState.Kana {
		watcher.ledState.Kana = nKana
		ledsChanged.Kana = true
		hasChanged = true
	}

	if !watcher.hasInitialSate {
		//This is the first led state reported, so former state is undefined and we consider everything as a change
		watcher.ledState.NumLock = nNum
		watcher.ledState.CapsLock = nCaps
		watcher.ledState.ScrollLock = nScroll
		watcher.ledState.Compose = nCompose
		watcher.ledState.Kana = nKana

		ledsChanged.NumLock = true
		ledsChanged.CapsLock = true
		ledsChanged.ScrollLock = true
		ledsChanged.Compose = true
		ledsChanged.Kana = true

		hasChanged = true

		watcher.hasInitialSate = true //don't do this again
	}

	//check if we have listeners ready to remove
	rmList := make([]*HIDKeyboardLEDListener,0)
	for l,_ := range watcher.listeners.m {
		if l.isMarkedForDeletion {
			//add to remove list
			rmList = append(rmList, l)
		}
	}
	//remove listeners from map, which have been found ready to remove
	watcher.listeners.Lock()
	for _,l := range rmList { delete(watcher.listeners.m, l) }
	watcher.listeners.Unlock()

	//check if we have listeners left to consume the produced dispatchListeners, blocking wait otherwise
	for len(watcher.listeners.m) == 0 {
		//fmt.Println("Waiting for new LED Listener ...")
		watcher.listenerNonZeroCondLock.Lock()
		watcher.listenerNonZeroCond.Wait() //wait till at least one listener is present
		watcher.listenerNonZeroCondLock.Unlock()
		//fmt.Println("... done waiting , at least one LED state listener registered!")
	}


	//Inform listeners about change
	if hasChanged {
		//log.Printf("INFORMING LISTENERS Changed elements %+v for new LED sate: %+v\n", ledsChanged, watcher.ledState)
		watcher.listeners.Lock()


		//idx := 0
		for _,listener := range watcher.listeners.m {
			//log.Printf("Sending listener number %d led state change %+v\n", idx, ledsChanged)
			//idx+=1

			//only push to channel if listener isn't marked for deletion meanwhile
			if !listener.isMarkedForDeletion { listener.changedLeds <- ledsChanged }
		}
		watcher.listeners.Unlock()
		//log.Println("...END INFORMING LISTENERS")
	}
}

/*
Waits for single LED state change
intendedChange: Mask values combined with logical or, to indicate which LEDs are allowed to trigger MaskNu
return value changed: Mask values combined with logical or, indicating which LED actually changed in order to stop waiting
 */
func (kbd *HIDKeyboard) WaitLEDStateChange(intendedChange byte, timeout time.Duration) (changed *HIDLEDState,err error) {

	//ToDo: If neW LED state is 0 (single LED on and is turned off) any doesn't trigger

	//register state change listener
	l := kbd.LEDWatcher.CreateAndAddNewListener()
	defer kbd.LEDWatcher.removeListener(l)

	startTime := time.Now()
	remaining := timeout

	for {
		//calculate remaining timeout (error out if already reached
		passedBy := time.Since(startTime)
		if passedBy > timeout {
			return nil, ErrTimeout
		} else {
			remaining = timeout - passedBy
		}

		//Wait for report of LED change
		select {
		case ledsChanged := <- l.changedLeds:
			//we have a state change, check relevance
			fmt.Printf("LEDListener received state change on following LEDs %+v\n", ledsChanged)
			hasRelevantChange := false
			relevantChanges := &HIDLEDState{}

			if (ledsChanged.NumLock) && (intendedChange & MaskNumLock > 0)  {
				hasRelevantChange = true
				relevantChanges.NumLock = ledsChanged.NumLock
			}
			if (ledsChanged.CapsLock) && (intendedChange & MaskCapsLock > 0) {
				hasRelevantChange = true
				relevantChanges.CapsLock = ledsChanged.CapsLock
			}
			if (ledsChanged.ScrollLock) && (intendedChange & MaskScrollLock > 0) {
				hasRelevantChange = true
				relevantChanges.NumLock = ledsChanged.NumLock
			}
			if (ledsChanged.Compose) && (intendedChange & MaskCompose > 0) {
				hasRelevantChange = true
				relevantChanges.Compose = ledsChanged.Compose
			}
			if (ledsChanged.Kana) && (intendedChange & MaskKana > 0) {
				hasRelevantChange = true
				relevantChanges.Kana = ledsChanged.Kana
			}

			if hasRelevantChange {
				//We have an intended state change
				//fmt.Printf("LEDListener: the following changes have been relevant %+v\n", relevantChanges)
				return relevantChanges, nil
			}
			//If here, there was a LED state change, but not one we want to use for triggering
		case <- time.After(remaining):
			return nil, ErrTimeout
		}
	}

	//return 0,nil
}
