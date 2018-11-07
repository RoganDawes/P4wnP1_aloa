package service

import (
	"errors"
	"fmt"
	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
	"sync"

	pb "github.com/mame82/P4wnP1_go/proto"
)

/*
ToDo: Entprellen + StopAllEdgeDetecting (when new TriggerActionSet is deployed)
ToDo: If a single item of the trigger action set changes, all GPIOTriggers have to be re-enumearted, to assure that there's no edge detection running for a trigger which doesn't exist anymore
 */

var (
	EGpioNotAvailable = errors.New("sub system GPIO not available")
	EGpioPinInvalid   = errors.New("invalid GPIO pin")
)

type GpioManager struct {
	availableGpioPins []gpio.PinIO
	availableGpioNames []string

	rootSvc *Service

	edgeDetectingMutex *sync.Mutex
	edgeDetecting map[gpio.PinIO]bool

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
	state,err := host.Init()
	if err != nil {
		gm.IsUsable = false
		return
	}

	gm.State = state
	gm.IsUsable = rpi.Present()

	gpios := gpioreg.All()
	for _,g := range gpios {
		if pinreg.IsConnected(g) {
			gm.availableGpioPins = append(gm.availableGpioPins,g)
			gm.availableGpioNames = append(gm.availableGpioNames, g.Name())
		}
	}

	gm.edgeDetecting = make(map[gpio.PinIO]bool)
	gm.edgeDetectingMutex = &sync.Mutex{}

	return gm
}

func (gm GpioManager) GetAvailableGpios() (res []gpio.PinIO, err error) {
	if gm.IsUsable {
		return gm.availableGpioPins,nil
	}
	return res,EGpioNotAvailable
}

func (gm GpioManager) GetAvailableGpioNames() (res []string, err error) {
	if gm.IsUsable {
		return gm.availableGpioNames,nil
	}
	return res,EGpioNotAvailable
}

func (gm *GpioManager) DeployGpioTrigger(in *pb.TriggerGPIOIn) (err error) {
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

	go func() {
		for gm.isEdgeDetecting(p) {
			p.WaitForEdge(-1)

			//Edge detected, check if still edge detecting before consuming
			if gm.isEdgeDetecting(p) {
				switch p.Read() {
				case gpio.High:
					fmt.Println("Gpio " + p.Name() + " changed to high")
					gm.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGpioIn(p.Name(), bool(gpio.High)))
				case gpio.Low:
					fmt.Println("Gpio " + p.Name() + " changed to low")
					gm.rootSvc.SubSysEvent.Emit(ConstructEventTriggerGpioIn(p.Name(), bool(gpio.Low)))
				}

			} else {
				//exit for loop
				break
			}
		}
		gm.edgeDetectingMutex.Lock()
		delete(gm.edgeDetecting, p)
		gm.edgeDetectingMutex.Unlock()
	}()


	return nil
}

func (gm *GpioManager) FireGpioAction(out *pb.ActionGPIOOut) (err error) {
	if !gm.IsUsable {
		return EGpioNotAvailable
	}

	p := gpioreg.ByName(out.GpioName)
	if p == nil {
		return EGpioPinInvalid
	}

	// If edge detection is already running, stop it
	if gm.isEdgeDetecting(p) {
		gm.stopEdgeDetectionForPin(p)
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

	p.Out(level)

	return nil
}



func (gm *GpioManager) isEdgeDetecting(p gpio.PinIO) bool {
	fmt.Println("Check edge detection for " + p.Name())
	gm.edgeDetectingMutex.Lock()
	defer gm.edgeDetectingMutex.Unlock()
	if _,exists := gm.edgeDetecting[p]; exists && gm.edgeDetecting[p] {
		fmt.Println("Edge detection for " + p.Name() + " running")
		return true
	}
	fmt.Println("Edge detection for " + p.Name() + " not running")
	return false
}

func (gm *GpioManager) stopEdgeDetectionForPin(p gpio.PinIO) {
	gm.edgeDetectingMutex.Lock()
	if _,exists := gm.edgeDetecting[p]; exists {
		gm.edgeDetecting[p] = false
	}
	gm.edgeDetectingMutex.Unlock()

	//write high/low till gpio is deleted from map, to assure pending edge detection ended
	for gm.isEdgeDetecting(p) {
		p.Out(gpio.High)
		if gm.isEdgeDetecting(p) {
			p.Out(gpio.Low)
		}
	}

}