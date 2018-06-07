package hid

import (
	"sync"
	"os"
	"time"
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

func (s *HIDLEDState) fillState(stateByte byte) {
	if stateByte & MaskNumLock > 0 { s.NumLock = true  }
	if stateByte & MaskCapsLock > 0 { s.CapsLock = true }
	if stateByte & MaskScrollLock > 0 { s.ScrollLock = true }
	if stateByte & MaskCompose > 0 { s.Compose = true }
	if stateByte & MaskKana > 0 { s.Kana = true }
	return
}


type HIDKeyboardLEDStateWatcher struct {
	ledState *HIDLEDState //global LED state
	listeners *listenersmap //map of registered listeners
	listenerNonZeroCond *sync.Cond //condition to signaling in case at least one listener is present (dispatcher blocks if no listener is registered and receiving latest state changes)
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


func newHIDKeyboardLEDStateWatcher(devicePath string) (watcher *HIDKeyboardLEDStateWatcher,err error) {
	//llock := &sync.Mutex{}
	watcher = &HIDKeyboardLEDStateWatcher{
		ledState: &HIDLEDState{},
		listeners: &listenersmap{m: make(map[*HIDKeyboardLEDListener]*HIDKeyboardLEDListener)},
		listenerNonZeroCond: &sync.Cond{L:&sync.Mutex{}},
	}

	//start go routine reading LED output reports from keyboard device
	//ToDo: this should happen only once

	go watcher.start(devicePath)

	return
}

func (watcher *HIDKeyboardLEDStateWatcher) CreateAndAddNewListener() (l *HIDKeyboardLEDListener) {
	l = &HIDKeyboardLEDListener{
		ledWatcher: watcher,
		changedLeds: make(chan HIDLEDState),
	}
	watcher.listeners.Lock()
	watcher.listeners.m[l] = l
	watcher.listenerNonZeroCond.L.Lock()
	watcher.listenerNonZeroCond.Broadcast()
	watcher.listenerNonZeroCond.L.Unlock()
	watcher.listeners.Unlock()
	return
}

func (watcher *HIDKeyboardLEDStateWatcher) removeListener(l *HIDKeyboardLEDListener) {
	l.isMarkedForDeletion = true //mark listener for deletion to avoid that dispatcher write to channel
	//consume remaining channel data to unlock dispatchListeners loop
L:
	for {
		select {
		case <-l.changedLeds:
			//do nothinng
		default:
			break L
		}

	}

	//lock listener map and delete listener
	watcher.listeners.Lock()
	delete(watcher.listeners.m, l)
	watcher.listeners.Unlock()


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
	newState := HIDLEDState{}
	newState.fillState(state)


	hasChanged := false
	ledsChanged := HIDLEDState{}
	if newState.NumLock != watcher.ledState.NumLock {
		watcher.ledState.NumLock = newState.NumLock
		ledsChanged.NumLock = true
		hasChanged = true
	}
	if newState.CapsLock != watcher.ledState.CapsLock {
		watcher.ledState.CapsLock = newState.CapsLock
		ledsChanged.CapsLock = true
		hasChanged = true
	}
	if newState.ScrollLock != watcher.ledState.ScrollLock {
		watcher.ledState.ScrollLock = newState.ScrollLock
		ledsChanged.ScrollLock = true
		hasChanged = true
	}
	if newState.Compose != watcher.ledState.Compose {
		watcher.ledState.Compose = newState.Compose
		ledsChanged.Compose = true
		hasChanged = true
	}
	if newState.Kana != watcher.ledState.Kana {
		watcher.ledState.Kana = newState.Kana
		ledsChanged.Kana = true
		hasChanged = true
	}

	if !watcher.hasInitialSate {
		//This is the first led state reported, so former state is undefined and we consider everything as a change
		watcher.ledState.NumLock = newState.NumLock
		watcher.ledState.CapsLock = newState.CapsLock
		watcher.ledState.ScrollLock = newState.ScrollLock
		watcher.ledState.Compose = newState.Compose
		watcher.ledState.Kana = newState.Kana

		ledsChanged.NumLock = true
		ledsChanged.CapsLock = true
		ledsChanged.ScrollLock = true
		ledsChanged.Compose = true
		ledsChanged.Kana = true

		hasChanged = true

		watcher.hasInitialSate = true //don't do this again
	}



	//Inform listeners about change
	if hasChanged {
		//check if we have listeners left to consume the produced dispatchListeners, blocking wait otherwise
		for len(watcher.listeners.m) == 0 {
			//fmt.Println("Waiting for new LED Listener ...")
			watcher.listenerNonZeroCond.L.Lock()
			watcher.listenerNonZeroCond.Wait() //wait till at least one listener is present
			watcher.listenerNonZeroCond.L.Unlock()
			//fmt.Println("... done waiting , at least one LED state listener registered!")
		}


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
	//If there's no change, we ignore the new state
}

/*
Waits for single LED state change
intendedChange: Mask values combined with logical or, to indicate which LEDs are allowed to trigger MaskNu
return value changed: Mask values combined with logical or, indicating which LED actually changed in order to stop waiting
 */
func (kbd *HIDKeyboard) WaitLEDStateChange(intendedChange byte, timeout time.Duration) (changed *HIDLEDState,err error) {
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
			//fmt.Printf("LEDListener received state change on following LEDs %+v\n", ledsChanged)
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
			//If here, there was a LED state change, but not one we want to use for triggering (continue outer loop, consuming channel data)
		case <- time.After(remaining):
			return nil, ErrTimeout
		}
	}
}
