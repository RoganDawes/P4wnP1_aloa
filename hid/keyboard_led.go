package hid

import (
	"sync"
	"os"
	"time"
	"log"
	"fmt"
	"context"
)

var (
	//privateCurrentLEDWatcher *HIDKeyboardLEDStateWatcher = nil
	privateCurrentKeyboardLEDWatcher *KeyboardLEDStateWatcher = nil
)

const (
	MaskNumLock    = 1 << 0
	MaskCapsLock   = 1 << 1
	MaskScrollLock = 1 << 2
	MaskCompose    = 1 << 3
	MaskKana       = 1 << 4
	MaskNone       = 1 << 7 //not really a mask, indicates no change
	MaskAny = MaskNumLock | MaskCapsLock | MaskScrollLock | MaskCompose | MaskKana
	MaskAnyOrNone = MaskNumLock | MaskCapsLock | MaskScrollLock | MaskCompose | MaskKana | MaskNone
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


type lockableListenerMap struct {
	sync.Mutex
	m map[*KeyboardLEDStateListener]bool
}

type KeyboardLEDStateWatcher struct {
	ledState        *HIDLEDState //latest global LED state
	listeners       *lockableListenerMap //map of registered listeners
	ledStateFile    *os.File
	hasInitialState bool //marks if the initial state is define (gets true after first LED state ha been read)

	//listener which should be registered are put into a zero length channel, to allow blocking till a listener is added
	//in case there isn't already a listener (avoid reading LED states, without having a single listener consuming them)
	listenersToAdd     chan *KeyboardLEDStateListener
	readerToDispatcher chan HIDLEDState
	ctx                context.Context
	cancelFunc         context.CancelFunc
	isUsable           bool
}


func NewLEDStateWatcher(ctx context.Context, devFilePath string) (res *KeyboardLEDStateWatcher,err error) {
	//try to open the devFile
	devFile, err := os.Open(devFilePath)
	if err != nil { return }

	if privateCurrentKeyboardLEDWatcher != nil {
		privateCurrentKeyboardLEDWatcher.Stop()
	}

	ctx,cancel := context.WithCancel(ctx)

	res = &KeyboardLEDStateWatcher{
		ledState:        &HIDLEDState{},
		hasInitialState: false,
		ledStateFile:    devFile,
		listeners: &lockableListenerMap{
			m: make(map[*KeyboardLEDStateListener]bool),
		},
		// Buffer at least one listener, to avoid blocking when one is added, the channel is only used to block the
		// dispatchLoop, in case there's no registered listener (by reading from the listenerToAdd chanel
		listenersToAdd:     make(chan *KeyboardLEDStateListener,1),
		readerToDispatcher: make(chan HIDLEDState), //communicates new LED states to from file reader loop to dispatcher loop, blocks till consumed
		ctx:                ctx,
		cancelFunc:         cancel,



		/*
		ledState: &HIDLEDState{},
		listeners: &listenersmap{m: make(map[*HIDKeyboardLEDListener]*HIDKeyboardLEDListener)},
		addListeners: make(chan *HIDKeyboardLEDListener,1), //Buffer at least one, to avoid blocking `CreateAndAddNewListener` (we only want to block `dispatchListeners` in case there's no listener)
		*/
	}

	go res.readLoop()
	go res.dispatchLoop()
	privateCurrentKeyboardLEDWatcher = res
	res.isUsable = true
	return
}


func (w *KeyboardLEDStateWatcher) RetrieveNewListener() (l *KeyboardLEDStateListener, err error) {
	if !w.isUsable { return nil, ErrNotAllowed }

	//create listener and assgin watcher as parent
	l = &KeyboardLEDStateListener{
		isMarkedForDeletion: false,
		changedLeds: make(chan HIDLEDState),
		//interrupt: make(chan struct{}),
		ledWatcher: w,
	}

	//addListener to map
	w.listenersToAdd <- l

	return l,nil
}


// - unregisters all listeners
// - listeners which are still processed receive an interrupt on their interrupt channel, it is the responsibility
// of the listener to deal with this interrupt and close the channel after the interrupt is consumed (wrting to the channel
// happens only once)
// - close ledStateFile
func (w *KeyboardLEDStateWatcher) Stop() (err error) {
	//ToDo: A crash occurs from time to time, if the underlying file is already gone (gadget destroyed), this has to be called before the file is removed
	err = w.ledStateFile.Close() //produces an error in readLoop which gets translated to an interrupt
	return nil
}



//reads the LED state from device file, till os.File object is closed
func (w *KeyboardLEDStateWatcher) readLoop() {
	fmt.Println("***Starting LED reader")

	defer w.ledStateFile.Close() //Assure File object is closed after this loop, if the closed file wasn't the reason for loop abort

	buf := make([]byte, 1)
	for {
		//fmt.Println("-----------\nLED READ LOOP\n-------------------")
		n,err := w.ledStateFile.Read(buf)
		if err != nil {
			log.Printf("Keyboard LED watcher: LED file seems to be closed %s: %v\n...interrupting dispatcher loop!\n", w.ledStateFile.Name(), err)
			//mark watcher as unusable
			w.isUsable = false

			//interrupt the dispatcher loop, by cancelling the context
			w.cancelFunc()

			break
		}
		for i:=0; i<n; i++ {
			newState := HIDLEDState{}
			newState.fillState(buf[i])

			w.readerToDispatcher <- newState
		}
	}
	fmt.Println("***Stopped LED reader")
}

func (w *KeyboardLEDStateWatcher) dispatchLoop() {
	fmt.Println("***Starting LED dispatcher")

	//try to consume a new LEDState (blocking wait if no interrupt)
	L:
	for {
		select {
			case newState:= <- w.readerToDispatcher:
				//fmt.Printf("**** HANDLING new LED state\n")
				//fmt.Printf("Old state: %+v\nNew state:%+v\n", w.ledState, newState)


				//Translate received LED state to state change (if first received state, everything is considered as change
				ledStateChange := w.ledState.Changes(newState)
				//fmt.Printf("Changed state: %+v\n", ledStateChange)
				if w.hasInitialState == false {
					ledStateChange.fillState(MaskAny)
					w.hasInitialState = true
				}

				//Store new LED state
				w.ledState = &newState //Note: as this method blocks, in case there's no LED state listener, global LED state is only update in case a listener is registered

				// check if there's at least one listener, if not block till a new one is registered
				// Blocking, in case there's no listener, doesn't align to the usual approach of event driven
				//
				if len(w.listeners.m) == 0 {
					//fmt.Println("Waiting fo at least one listener")
					w.listeners.Lock()
					w.listeners.m[<-w.listenersToAdd] = true
					//fmt.Println("At least one listener added")
					w.listeners.Unlock()
				}

				deleteList := make([]*KeyboardLEDStateListener,0)
				//fan out to every listener, block if one doesn't consume the change
				w.listeners.Lock()
				for l,_ := range w.listeners.m {
					// Beware the DeadLock:
					// If the listener decides to remove itself (sets l.isMarkedForDeletion) based on the ledStateChange
					// received on the l.changedLeds channel, without continuing consuming data, this could lead to a dead lock
					// on w.listeners.
					// Example: The listener consumes a single state change (by reading from l.changedLeds) does some time
					// consuming processing and, as a result of this time consuming processing, decides to set l.isMarkedForDeletion to
					// true. This means the listener consumes only the first state change from l.changedLeds. As isMarkedForDeletion
					// isn't set to true immediately, this loop, again, tries to write data to the channel (if changed LED states are
					// produced frequently or have been queued already), BUT WRITING BLOCKS as the listener is going to decide to mark
					// itself for deletion after he is done with processing the first state change. This means the next line of code blocks,
					// as the data written to the channel is never consumed by the listener (and the channel is marked
					// for deletion too late, which would prevent writing to it). This again means, that the outer for loop,
					// which iterates over the registered listeners, will never exit, ULTIMATELY LEAVING w.listeners IN LOCKED
					// STATE. This becomes a deadlock, as soon as another routine tries to remove a listener from w.listeners,
					// as this routine has to call w.listeners.Lock() itself. But exactly the case of trying to remove the listener
					// from w.listeners is the expected one, after setting isMarkedForDeletion to true.
					//
					// Solution: No matter where l.isMarkedForDeletion gets set to true, it is the responsibility of exactly
					// the same part of code, to consume all remaining channel data from l.changedLeds afterwards.
					if !l.isMarkedForDeletion {
						l.changedLeds <- ledStateChange
					} else {
						deleteList = append(deleteList, l)
					}
				}

				//delete listeners which aren't used anymore
				for _,l := range deleteList {
					delete(w.listeners.m, l)
				}
				w.listeners.Unlock()
			case <- w.ctx.Done():
				//inform all listeners about the interrupt
				w.listeners.Lock()

				for l,_ := range w.listeners.m {
					l.Unregister()
				}
				w.listeners.Unlock()

				//close Watcher channels
				close(w.listenersToAdd)

				//End the dispatcher loop
				break L
			case newListener := <-w.listenersToAdd:
				w.listeners.Lock()
				w.listeners.m[newListener] = true
				w.listeners.Unlock()
		}
	}
	fmt.Println("***Stopped LED dispatcher")
}

type KeyboardLEDStateListener struct {
	ledWatcher          *KeyboardLEDStateWatcher //the parent LEDWatcher, containing global ledState
	changedLeds         chan HIDLEDState //changedLeds represents the LEDs which change since last report as bitfield (MaskNumLock, MaskCapsLock ...)  the actual state has to be fetched from the respective field of the ledWatcher.ledState
	//interrupt			chan struct{}
	isMarkedForDeletion bool
}

func (l *KeyboardLEDStateListener) Unregister() {
	if l.isMarkedForDeletion {return}
	l.isMarkedForDeletion = true
	//consume remaining input data from channel
	L:
	for {
		select {
		case <-l.changedLeds:
			//do nothinng
		default:
			break L
		}

	}

	//close channels
	//close(l.interrupt)
	close(l.changedLeds)
}

/*
Waits for single LED state change
intendedChange: Mask values combined with logical or, to indicate which LEDs are allowed to trigger MaskNu
return value changed: Mask values combined with logical or, indicating which LED actually changed in order to stop waiting
 */
func (kbd *HIDKeyboard) WaitLEDStateChange(irqFunc <-chan func(), intendedChange byte, timeout time.Duration) (changed *HIDLEDState,err error) {
	//register state change listener
	l,err := kbd.LEDWatcher.RetrieveNewListener()
	if err!= nil { return nil,err }
	//defer kbd.LEDWatcher.removeListener(l)
	defer l.Unregister()

	startTime := time.Now()
	remaining := timeout

	intendedChangeStruct := HIDLEDState{}
	intendedChangeStruct.fillState(intendedChange)

	for {
		fmt.Println("LED change loop...")
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

			// special case - MaskNone is enabled we report back an LED change, even if nothing changed
			// this could be used to trigger on Keyboard re-attachment on Windows/som Linox distros, as a new
			// LED state is reported everytime the keyboard is attached, even if it doesn't differ from the old one
			if intendedChange & MaskNone > 0 {
				//fmt.Printf("LEDListener: no changes, reporting back anyway %+v\n", relevantChanges)
				return &relevantChanges, nil
			}

			//If here, there was a LED state change, but not one we want to use for triggering (continue outer loop, consuming channel data)
		case <-l.ledWatcher.ctx.Done():
			fmt.Println("...LEDWatcher aborted")
			return nil, ErrAbort
		case irq:=<-irqFunc:
			fmt.Println("...WaitLEDStateChange received Irq")
			irq()
			return nil, ErrIrq
		case <- time.After(remaining):
			return nil, ErrTimeout
		}
	}
}

func (kbd *HIDKeyboard) WaitLEDStateChangeRepeated(irqFunc <-chan func(), intendedChange byte, repeatCount int, minRepeatDelay time.Duration, timeout time.Duration) (changed *HIDLEDState,err error) {
	//register state change listener
	l,err := kbd.LEDWatcher.RetrieveNewListener()
	if err!= nil { return nil,err }
	//defer kbd.LEDWatcher.removeListener(l)
	defer l.Unregister()

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

				//log.Printf("\tRelevant LED changes after applying mask (interval %v) NUM: %v CAPS: %v SCROLL: %v COMPOSE: %v KANA: %v\n", minRepeatDelay, countNum, countCaps, countScroll, countCompose, countKana)


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
		case <-l.ledWatcher.ctx.Done():
			return nil, ErrAbort
		case irq:=<-irqFunc:
			irq()
			return nil, ErrIrq
		case <- time.After(remaining):
			return nil, ErrTimeout
		}
	}
}
