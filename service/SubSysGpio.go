// +build linux

package service

import (
	"context"
	"fmt"
	"github.com/mame82/P4wnP1_go/service/pgpio"
	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
	"sync"
	"time"
	"errors"
	pb "github.com/mame82/P4wnP1_go/proto"
)

var (
	EGpioNotAvailable = errors.New("sub system GPIO not available")
	EGpioPinInvalid   = errors.New("invalid GPIO pin")
)

type GpioManager struct {
	availableGpioPins  []*pgpio.P4wnp1PinIO
	availableGpioPinsMap  map[string]*pgpio.P4wnp1PinIO
	availableGpioNames []string

	rootSvc *Service

	edgeDetectingMutex *sync.Mutex
	edgeDetecting      map[gpio.PinIO]bool

	IsUsable bool

	*periph.State
}

func (gm *GpioManager) Start() {

}

func (gm *GpioManager) Stop() {
}

func NewGpioManager(rootSvc *Service) (res *GpioManager) {
	gm := &GpioManager{
		rootSvc: rootSvc,
	}
	state, err := host.Init()
	if err != nil {
		gm.IsUsable = false
		return
	}

	gm.State = state
	gm.IsUsable = rpi.Present()

	gm.availableGpioPinsMap = make(map[string]*pgpio.P4wnp1PinIO)
	gpios := gpioreg.All()
	for _, g := range gpios {
		if pinreg.IsConnected(g) {
			ppin := pgpio.NewP4wnp1PinIO(g)
			gm.availableGpioPins = append(gm.availableGpioPins, ppin)
			gm.availableGpioPinsMap[g.Name()] = ppin
			gm.availableGpioNames = append(gm.availableGpioNames, g.Name())
		}
	}


	gm.edgeDetecting = make(map[gpio.PinIO]bool)
	gm.edgeDetectingMutex = &sync.Mutex{}

	return gm
}

func (gm *GpioManager) GetAvailableGpios() (res []*pgpio.P4wnp1PinIO, err error) {
	if gm.IsUsable {
		return gm.availableGpioPins, nil
	}
	return res, EGpioNotAvailable
}

func (gm *GpioManager) GetAvailableGpioNames() (res []string, err error) {
	if gm.IsUsable {
		return gm.availableGpioNames, nil
	}
	return res, EGpioNotAvailable
}



func (gm *GpioManager) DeployGpioTrigger(in *pb.TriggerGPIOIn) (err error) {
	if !gm.IsUsable {
		return EGpioNotAvailable
	}

	p,present := gm.availableGpioPinsMap[in.GpioName]
	if !present {
		return EGpioPinInvalid
	}

	fmt.Printf("Deploying trigger for GPIO: %+v\n", p)

	pull := gpio.Float
	switch in.PullUpDown {
	case pb.GPIOInPullUpDown_DOWN:
		pull = gpio.PullDown
	case pb.GPIOInPullUpDown_UP:
		pull = gpio.PullUp
	}

	edge := gpio.BothEdges
	switch in.GpioInEdge {
	case pb.GPIOInEdge_FALLING:
		edge = gpio.FallingEdge
	case pb.GPIOInEdge_RISING:
		edge = gpio.RisingEdge
	}

	p.In(pull, edge)

	debounceDelay := time.Duration(in.DebounceMillis) * time.Millisecond
	//ToDo: remove next line, as this is a fixed test delay of 100 ms
	debounceDelay = 100 * time.Millisecond

	go func() {
		fmt.Println("Starting edge detection for pin " + p.Name())
		detectErr := error(nil)
		for detectErr == nil {
			var detectedLevel gpio.Level
			detectedLevel,detectErr = p.ExtWaitForEdge(context.Background(), debounceDelay)

			fmt.Printf("... done wait for edge %s level: %v\n", p.Name(), detectedLevel)

			//Edge detected, check if still edge detecting before consuming

			switch detectedLevel {
				case gpio.High:
					fmt.Println("Gpio " + p.Name() + " changed to high")
					gm.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGpioIn(p.Name(), bool(gpio.High)))

				case gpio.Low:
					fmt.Println("Gpio " + p.Name() + " changed to low")
					gm.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGpioIn(p.Name(), bool(gpio.Low)))
				}
		}
		fmt.Println("!!!! STOPPED edge loop for pin " + p.Name())
	}()

	return nil
}

/*
func (gm *GpioManager) DeployGpioTriggerOld(in *pb.TriggerGPIOIn) (err error) {
	if !gm.IsUsable {
		return EGpioNotAvailable
	}

	p := gpioreg.ByName(in.GpioName)
	if p == nil {
		return EGpioPinInvalid
	}

	pull := gpio.Float
	switch in.PullUpDown {
	case pb.GPIOInPullUpDown_DOWN:
		pull = gpio.PullDown
	case pb.GPIOInPullUpDown_UP:
		pull = gpio.PullUp
	}

	edge := gpio.BothEdges
	switch in.GpioInEdge {
	case pb.GPIOInEdge_FALLING:
		edge = gpio.FallingEdge
	case pb.GPIOInEdge_RISING:
		edge = gpio.RisingEdge
	}

	// If edge detection is already running, stop it
	if gm.isEdgeDetecting(p) {
		gm.stopEdgeDetectionForPin(p)
	}

	p.In(pull,edge)

	gm.edgeDetectingMutex.Lock()
	gm.edgeDetecting[p] = true
	gm.edgeDetectingMutex.Unlock()

	debounceDelay := time.Duration(in.DebounceMillis) * time.Millisecond
	//ToDo: remove next line, as this is a fixed test delay of 100 ms
	debounceDelay = 100 *time.Millisecond

	go func() {
		fmt.Println("Starting edge detection for pin " + p.Name())
		lastHit := time.Now()
		for gm.isEdgeDetecting(p) {
			fmt.Println("Wait for edge " + p.Name() + " ...")
			waitSuccess := p.WaitForEdge(-1)
			fmt.Printf("... done wait for edge %s waitSucces: %v\n", p.Name(), waitSuccess)
			if !waitSuccess {
				break
			}
			now := time.Now()
			bounce := true
			sinceLastHit := now.Sub(lastHit)
			if sinceLastHit > debounceDelay {
				bounce = false
				// could do `lastHit = now` if timer only should be reset for "no bounce"
			}
			lastHit = now

			//Edge detected, check if still edge detecting before consuming
			if gm.isEdgeDetecting(p) {
				switch p.Read() {
				case gpio.High:
					fmt.Println("Gpio " + p.Name() + " changed to high")
					if !bounce {
						gm.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGpioIn(p.Name(), bool(gpio.High)))
					} else {
						fmt.Println("... ignored, as in debounceDelay")
					}
				case gpio.Low:
					fmt.Println("Gpio " + p.Name() + " changed to low")
					if !bounce {
						gm.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGpioIn(p.Name(), bool(gpio.Low)))
					} else {
						fmt.Println("... ignored, as in debounceDelay")
					}
				}

			} else {
				//exit for loop
				break
			}
		}
		fmt.Println("STOPPED edge detection for pin " + p.Name())
		gm.edgeDetectingMutex.Lock()
		delete(gm.edgeDetecting, p)
		gm.edgeDetectingMutex.Unlock()
	}()


	return nil
}
*/

func (gm *GpioManager) FireGpioAction(out *pb.ActionGPIOOut) (err error) {
	fmt.Println("FireGPIOAction for", out.GpioName)
	if !gm.IsUsable {
		return EGpioNotAvailable
	}

	p,present := gm.availableGpioPinsMap[out.GpioName]
	if !present {
		return EGpioPinInvalid
	}

	level := gpio.Low
	switch out.Value {
	case pb.GPIOOutValue_HIGH:
		level = gpio.High
	case pb.GPIOOutValue_TOGGLE:
		if p.Read() == gpio.Low {
			level = gpio.High
		}
	}

	fmt.Printf("Setting %s to out level %v...\n", p.Name(), level)

	p.Out(level)
	fmt.Println("..setting level done")

	return nil
}

func (gm *GpioManager) ResetPins() {
	fmt.Println("Resetting all pins")
	for _, pin := range gm.availableGpioPins {
		if pin.Edge() != gpio.NoEdge {
			pin.In(gpio.Float, gpio.NoEdge)
		}
		fmt.Println("... halting pin " + pin.Name())
		pin.Halt()
	}

}

/*
func (gm *GpioManager) isEdgeDetecting(p gpio.PinIO) bool {
	gm.edgeDetectingMutex.Lock()
	defer gm.edgeDetectingMutex.Unlock()
	if _,exists := gm.edgeDetecting[p]; exists && gm.edgeDetecting[p] {
		fmt.Println("Edge detection for " + p.Name() + " running")
		return true
	}
	fmt.Println("Edge detection for " + p.Name() + " not running")
	return false
}
*/

/*
func (gm *GpioManager) stopEdgeDetectionForPin(p gpio.PinIO) {
	fmt.Println("Stopping edge detection for pin " + p.Name())
	gm.edgeDetectingMutex.Lock()
	if _,exists := gm.edgeDetecting[p]; exists {
		gm.edgeDetecting[p] = false
	}
	gm.edgeDetectingMutex.Unlock()


	//p.Halt()
	p.In(gpio.Float, gpio.BothEdges)

	isNotDeleted := func() bool {
		gm.edgeDetectingMutex.Lock()
		defer gm.edgeDetectingMutex.Unlock()
		_,exists := gm.edgeDetecting[p]
		return exists
	}


	//write high/low till gpio is deleted from map, to assure pending edge detection ended
	for isNotDeleted() {
		fmt.Println("... stopping " + p.Name() + " setting PullDown to trigger final edge")
		p.In(gpio.PullDown, gpio.BothEdges)
		if isNotDeleted() {
			fmt.Println("... stopping " + p.Name() + " setting PullUp to trigger final edge")
			p.In(gpio.PullUp, gpio.BothEdges)
		}
	}

	p.In(gpio.Float, gpio.NoEdge)
}
*/
