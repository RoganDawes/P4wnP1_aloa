// +build linux

package pgpio

import (
	"context"
	"errors"
	"fmt"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"time"
)

var (
	EEdgeDetectNotRunning = errors.New("edge detection not running")
	EEdgeDetectAborted    = errors.New("edge detection aborted")
)

type steadyTimeLevel struct {
	SteadyTime time.Duration
	Level      gpio.Level
}

type P4wnp1PinIO struct {
	piPin gpio.PinIO

	/*
	edgeDetectLoopContext *context.Context //nil if no internal edge detect loop is running
	edgeDetectLoopCancel context.CancelFunc //nil if no internal edge detect loop is running
	edgeDetectionMutex    *sync.Mutex
	*/

	edge                 gpio.Edge
	pull                 gpio.Pull
	edgeDetectionAbort   bool
	edgeDetectionChannel chan steadyTimeLevel //Every time a valid edge is detected, this channel returns the duration since the last valid edge (==steadyTime. could be used for debounce) and the gpio.Level during edge event
	waitForEdgeStopped   bool
	lastEdgeTime         time.Time
}

func (p *P4wnp1PinIO) startEdgeDetection(pull gpio.Pull, edge gpio.Edge, preserveUnconsumed bool) (err error) {
	fmt.Println("starting edge detection for", p.Name(), "...")

	p.lastEdgeTime = time.Now()
	if preserveUnconsumed {
		p.edgeDetectionChannel = make(chan steadyTimeLevel, 0) // don't allow buffering --> block read loop if detected edge isn't consumed
	} else {
		p.edgeDetectionChannel = make(chan steadyTimeLevel, 1) // allow buffering a single value
	}

	p.edgeDetectionAbort = false
	p.waitForEdgeStopped = false

	go func() {
	Loop:
		for !p.edgeDetectionAbort {
			// Note: it seems to be impossible to force WaitForEdge to produce a 'false' result if no timeout is given.
			// Providing a timeout isn't suitable for this use case (continuous edge detection without interruption till
			// stopEdgeDetection() is called). A possible solution would be to call WaitForEdge in a loop with a timeout
			// and checking for the abort condition in every iteration, but this involves a tradeoff between responsiveness
			// to the stop method (long timeouts) and higher system load (short timeout).
			//
			// To cope with that, the state of `p.edgeDetectionAbort` is checked after every return of `WaitForEdge(-1)`.
			// In case no edges are detected after stopEdgeDetection has been called, this would block forever. In order to
			// deal with that, the stopEdgeDetection method changes the pull state of the GPIO resistor between PullUp and
			// PullDown, till this loop ends and indicates termination by setting p.waitForEdgeStopped to true.
			//
			// This approach is only tested on Pi0w and likely behaves different on other drivers or doesn't work at all
			// if no internal pull up/pull down resistors are present. The PinIO.Halt() method unfortunately doesn't interrupt
			// WaitForEdge.
			waitSuccess := p.piPin.WaitForEdge(-1)
			fmt.Println("Edge event received")
			if !waitSuccess {
				// don't consume the edge, stopEdgeDetection has been called meanwhile --> abort the loop
				fmt.Println("WaitForEdge failed, aborting EdgeDetection")
				break Loop
			}
			if p.edgeDetectionAbort {
				// don't consume the edge, stopEdgeDetection has been called meanwhile --> abort the loop
				fmt.Println("Ignored edge and because stopEdgeDetection was called, aborting edgeDetection")
				break Loop
			}

			//if here, we got a un-debounced and unvalidated edge and thus measure the duration since the last detected edge
			now := time.Now()
			timeSinceLastEdge := now.Sub(p.lastEdgeTime)
			p.lastEdgeTime = now

			var level gpio.Level
			switch p.edge {
			case gpio.RisingEdge:
				level = gpio.High // don't read from pin, assume correct detection
			case gpio.FallingEdge:
				level = gpio.Low // don't read from pin, assume correct detection
			case gpio.BothEdges:
				level = p.piPin.Read()
			default:
				// this includes gpio.NoEdge and shouldn't be set if this loop is running, so we abort the whole loop
				break Loop
			}

			// Send the steadyTimeLeve along the channel
			stl := steadyTimeLevel{
				Level:      level,
				SteadyTime: timeSinceLastEdge,
			}

			if preserveUnconsumed {
				p.edgeDetectionChannel <- stl //write steadyTimeLevel to channel and block loop till read by ExtWaitForEdge
				// ToDo: fix debounce agnostic decision
				// Note: this has influence on timeSince last edge. The debounce duration has to be known here, to avoid blocking
				// write to the channel (and possibly increasing timeSinceLastEdge), in case the event occurred during debounce duration.
			} else {
				//pop old stl from channel if needed
				for len(p.edgeDetectionChannel) > 0 {
					<-p.edgeDetectionChannel
				}
				p.edgeDetectionChannel <- stl //put latest stl to channel
			}
		}

		p.waitForEdgeStopped = true

		//indicate that edge detection has successfully terminated by closing the edgeDetectionChannel channel (and set to nil to use as indicator for isEdgeDetectionRunning, without reading from this channel)
		close(p.edgeDetectionChannel)
		p.edgeDetectionChannel = nil

		//assure edgeDetection is stopped
		fmt.Println("... edge detection for", p.Name(), "stopped")
	}()

	fmt.Println("... edge detection for", p.Name(), "started")
	return
}

func (p *P4wnp1PinIO) isEdgeDetectionRunning() bool {
	if p.edgeDetectionChannel == nil {
		//fmt.Println("edge detection not running")
		return false
	}
	//fmt.Println("edge detection running")
	return true
}

func (p *P4wnp1PinIO) stopEdgeDetection() error {
	if !p.isEdgeDetectionRunning() {
		return EEdgeDetectNotRunning
	}

	fmt.Println("stopping edge detection for", p.Name(), "...")
	p.edgeDetectionAbort = true

	// hackish approach to end WaitForEdge, by toggling Pull resistors (see comments in startEdgeDetection for details)
	for !p.waitForEdgeStopped {
		p.piPin.In(gpio.PullDown, p.edge)
		if !p.waitForEdgeStopped {
			p.piPin.In(gpio.PullUp, p.edge)
		}
	}

	// disable edge detection interrupt and restore pull resistor state
	p.piPin.In(p.pull, gpio.NoEdge)

	// wait till stop success is indicated
	if p.edgeDetectionChannel != nil { // if channel still exists
		<-p.edgeDetectionChannel //wait for close
	}

	return nil
}

func (p P4wnp1PinIO) Edge() gpio.Edge {
	return p.edge
}

func (p *P4wnp1PinIO) ExtWaitForEdge(ctx context.Context, debounceDuration time.Duration) (level gpio.Level, err error) {
	useDebounce := debounceDuration > 0

	for { // the select statement is wrapped into a for loop, to account for edges not fulfilling the debounceDuration criteria
		select {
		case steadyTimeLevel, channelOpen := <-p.edgeDetectionChannel:
			//channel closed, so we return error
			if !channelOpen {
				return level, EEdgeDetectNotRunning
			}
			if !useDebounce {
				// no debounce needed, ultimately return the edge change
				return steadyTimeLevel.Level, nil
			} else {
				// we have to assure that no other edge change event occurs in debounce duration (steady level) before returning
				if steadyTimeLevel.SteadyTime >= debounceDuration {
					//detected edge fulfills debounce criteria
					return steadyTimeLevel.Level, nil
				}
				// else ignore edge
				fmt.Printf("Ignored detected edge as invalidated by debounce: %v\n", steadyTimeLevel)
			}
		case <-ctx.Done():
			return level, EEdgeDetectAborted
		}
	}
}

/*
func (p P4wnp1PinIO) String() string {
	return p.piPin.String()
}
*/
func (p P4wnp1PinIO) Halt() error {
	return p.piPin.Halt()
}

func (p P4wnp1PinIO) Name() string {
	return p.piPin.Name()
}

func (p P4wnp1PinIO) Number() int {
	return p.piPin.Number()
}

func (p P4wnp1PinIO) Function() string {
	return p.piPin.Function()
}

func (p *P4wnp1PinIO) In(pull gpio.Pull, edge gpio.Edge) (err error) {
	p.stopEdgeDetection()
	// bring up edgeDetection go routine (again) if needed
	err = p.piPin.In(pull, edge)
	if err != nil {
		return
	}
	p.pull = pull
	p.edge = edge
	if edge != gpio.NoEdge {
		p.startEdgeDetection(pull, edge, false) //don't preserve unconsumed events
	}
	return err
}

func (p P4wnp1PinIO) Read() gpio.Level {
	return p.piPin.Read()
}

func (p P4wnp1PinIO) WaitForEdge(timeout time.Duration) bool {
	return p.piPin.WaitForEdge(timeout)
}

func (p P4wnp1PinIO) Pull() gpio.Pull {
	return p.piPin.Pull()
}

func (p P4wnp1PinIO) DefaultPull() gpio.Pull {
	return p.piPin.DefaultPull()
}

func (p *P4wnp1PinIO) Out(l gpio.Level) error {
	//stop edge detection, if needed
	p.stopEdgeDetection()
	return p.piPin.Out(l)
}

func (p *P4wnp1PinIO) PWM(duty gpio.Duty, f physic.Frequency) error {
	//stop edge detection, if needed
	p.stopEdgeDetection()
	return p.piPin.PWM(duty, f)
}

func NewP4wnp1PinIO(p gpio.PinIO) *P4wnp1PinIO {
	return &P4wnp1PinIO{piPin: p}
}
