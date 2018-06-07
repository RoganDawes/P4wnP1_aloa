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

func (s HIDLEDState) AnyOn() bool {
	return s.NumLock || s.CapsLock || s.ScrollLock || s.Compose || s.Kana
}

func (s HIDLEDState) Mask(mask HIDLEDState) (result HIDLEDState) {
	result.NumLock = s.NumLock && mask.NumLock
	result.CapsLock = s.CapsLock && mask.CapsLock
	result.ScrollLock = s.ScrollLock && mask.ScrollLock
	result.Compose = s.Compose && mask.Compose
	result.Kana = s.Kana && mask.Kana
	return
}

func (s HIDLEDState) Changes(other HIDLEDState) (result HIDLEDState) {
	result.NumLock = s.NumLock != other.NumLock
	result.CapsLock = s.CapsLock != other.CapsLock
	result.ScrollLock = s.ScrollLock != other.ScrollLock
	result.Compose = s.Compose != other.Compose
	result.Kana = s.Kana != other.Kana
	return
}




type HIDKeyboardLEDStateWatcher struct {
	ledState *HIDLEDState //global LED state
	listeners *listenersmap //map of registered listeners
	addListeners chan *HIDKeyboardLEDListener //channel which takes new listeners, used to block LED state dispatching in case there isn't at least one listenr in listeners
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
		//listenerNonZeroCond: &sync.Cond{L:&sync.Mutex{}},
		addListeners: make(chan *HIDKeyboardLEDListener,1), //Buffer at least one, to avoid blocking `CreateAndAddNewListener` (we only want to block `dispatchListeners` in case there's no listener)
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
	watcher.addListeners <- l
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


	close(l.changedLeds) //close channel (there should be no further read access)

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

	ledsChanged := watcher.ledState.Changes(newState)
	watcher.ledState.NumLock = newState.NumLock
	watcher.ledState.CapsLock = newState.CapsLock
	watcher.ledState.ScrollLock = newState.ScrollLock
	watcher.ledState.Compose = newState.Compose
	watcher.ledState.Kana = newState.Kana

	//fmt.Printf("Dispatcher LEDS changed: %+v\n", ledsChanged)
	hasChanged := ledsChanged.AnyOn()

	if !watcher.hasInitialSate {
		//This is the first led state reported, so former state is undefined and we consider everything as a change

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
		//log.Printf("INFORMING LISTENERS Changed elements %+v for new LED sate: %+v\n", ledsChanged, watcher.ledState)
		watcher.listeners.Lock()

		//check if we have listeners left to consume the produced LED state, blocking wait otherwise
		for len(watcher.listeners.m) == 0 {
			//fmt.Println("Waiting for new LED Listener ...")
			l := <-watcher.addListeners //get first listener with blocking wait
			watcher.listeners.m[l] = l
			//fmt.Println("... done waiting , at least one LED state listener registered!")
		}
		//add remaining listeners
		L:
			for {
				select {
				case l:= <- watcher.addListeners:
					watcher.listeners.m[l] = l
				default:
					break L
				}
			}


		for _,listener := range watcher.listeners.m {
			//log.Printf("Sending listener the led state change %+v\n", ledsChanged)
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

	intendedChangeStruct := HIDLEDState{}
	intendedChangeStruct.fillState(intendedChange)

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
			relevantChanges := ledsChanged.Mask(intendedChangeStruct)

			if relevantChanges.AnyOn() {
				//We have an intended state change
				//fmt.Printf("LEDListener: the following changes have been relevant %+v\n", relevantChanges)
				return &relevantChanges, nil
			}
			//If here, there was a LED state change, but not one we want to use for triggering (continue outer loop, consuming channel data)
		case <- time.After(remaining):
			return nil, ErrTimeout
		}
	}
}

func (kbd *HIDKeyboard) WaitLEDStateChangeRepeated(intendedChange byte, repeatCount int, minRepeatDelay time.Duration, timeout time.Duration) (changed *HIDLEDState,err error) {
	//register state change listener
	l := kbd.LEDWatcher.CreateAndAddNewListener()
	defer kbd.LEDWatcher.removeListener(l)

	startTime := time.Now()
	remaining := timeout

	intendedChangeStruct := HIDLEDState{}
	intendedChangeStruct.fillState(intendedChange)

	lastNum,lastCaps,lastScroll,lastCompose,lastKana := startTime,startTime,startTime,startTime,startTime
	countNum,countCaps,countScroll,countCompose,countKana := 0,0,0,0,0


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
			relevantChanges := ledsChanged.Mask(intendedChangeStruct)
//			fmt.Printf("LEDs changed:     %v\n", ledsChanged)
//			fmt.Printf("Changes relevant: %v\n", relevantChanges)

			if relevantChanges.AnyOn() {
				now := time.Now()
				//log.Printf("Duration: NUM %v, CAPS %v, SCROLL %v, COMPOSE %v, KANA %v\n", now.Sub(lastNum), now.Sub(lastCaps), now.Sub(lastScroll), now.Sub(lastCompose), now.Sub(lastKana))

				//We have an intended state change, check if any was in intended delay and increment counter in case
				if relevantChanges.NumLock {
					if now.Sub(lastNum) < minRepeatDelay {
						countNum += 1
					} else {
						countNum = 0
					}
					lastNum = now
				}
				if relevantChanges.CapsLock {
					if now.Sub(lastCaps) < minRepeatDelay {
						countCaps += 1
					} else {
						countCaps = 0
					}
					lastCaps = now
				}
				if relevantChanges.ScrollLock {
					if now.Sub(lastScroll) < minRepeatDelay {
						countScroll += 1
					} else {
						countScroll = 0
					}
					lastScroll = now
				}
				if relevantChanges.Compose {
					if now.Sub(lastCompose) < minRepeatDelay {
						countCompose += 1
					} else {
						countCompose = 0
					}
					lastCompose = now
				}
				if relevantChanges.Kana {
					if now.Sub(lastKana) < minRepeatDelay {
						countKana += 1
					} else {
						countKana = 0
					}
					lastKana = now
				}

				//log.Printf("Counters: NUM %d, CAPS %d, SCROLL %d, COMPOSE %d, KANA %d\n", countNum, countCaps, countScroll, countCompose, countKana)

				//check counters
				result := &HIDLEDState{}
				if countNum >= repeatCount { result.NumLock = true }
				if countCaps >= repeatCount { result.CapsLock = true }
				if countScroll >= repeatCount { result.ScrollLock = true }
				if countCompose >= repeatCount { result.Compose = true }
				if countKana >= repeatCount { result.Kana = true }

				if result.AnyOn() {
					return result, nil
				}
				//return &relevantChanges, nil
			}
			//If here, there was a LED state change, but not one we want to use for triggering (continue outer loop, consuming channel data)
		case <- time.After(remaining):
			return nil, ErrTimeout
		}
	}
}
