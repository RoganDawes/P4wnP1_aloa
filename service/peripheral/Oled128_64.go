package peripheral

import (
	"fmt"
	"github.com/warthog618/gpio"
	"time"
)

/*
// https://www.waveshare.com/1.3inch-oled-hat.htm

KEY1 			P21 		Button 1/GPIO
KEY2 			P20 		Button 2/GPIO
KEY3 			P16 		Button 3/GPIO
Joystick Up 	P6 			Joystick Up
Joystick Down 	P19 		Joystick Down
Joystick Left 	P5 			Joystick Left
Joystick Right 	P26 		Joystick Right
Joystick Press 	P13 		Joystick Press
SCLK 			P11/SCLK 	SPI clock input
MOSI 			P10/MOSI 	SPI data input
DC 				P24 		Data/Command selection (high for data, low for command)
CS 				P8/CE0 		Chip selection, low active
RST 			P25 		Reset, low active
 */

var timeLastEvent time.Time
const (
	WAVESHARE_OLED_GPIO_KEY_1 = gpio.GPIO21
	WAVESHARE_OLED_GPIO_KEY_2 = gpio.GPIO20
	WAVESHARE_OLED_GPIO_KEY_3 = gpio.GPIO16

	WAVESHARE_OLED_GPIO_JOY_UP    = gpio.GPIO6
	WAVESHARE_OLED_GPIO_JOY_DOWN  = gpio.GPIO19
	WAVESHARE_OLED_GPIO_JOY_LEFT  = gpio.GPIO5
	WAVESHARE_OLED_GPIO_JOY_RIGHT = gpio.GPIO26
	WAVESHARE_OLED_GPIO_JOY_PRESS = gpio.GPIO13
)

var waveshareRelevantEvents = []uint8{
	WAVESHARE_OLED_GPIO_JOY_DOWN,
	WAVESHARE_OLED_GPIO_JOY_LEFT,
	WAVESHARE_OLED_GPIO_JOY_RIGHT,
	WAVESHARE_OLED_GPIO_JOY_UP,
	WAVESHARE_OLED_GPIO_JOY_PRESS,
	WAVESHARE_OLED_GPIO_KEY_1,
	WAVESHARE_OLED_GPIO_KEY_2,
	WAVESHARE_OLED_GPIO_KEY_3,
}

type WaveshareOledEvent int

const (
	PRESSED_UP WaveshareOledEvent = iota
	PRESSED_DOWN
	PRESSED_LEFT
	PRESSED_RIGHT
	PRESSED_STICK

	PRESSED_KEY1
	PRESSED_KEY2
	PRESSED_KEY3
)

var pin2function = map[uint8]WaveshareOledEvent{
	WAVESHARE_OLED_GPIO_KEY_1:     PRESSED_KEY1,
	WAVESHARE_OLED_GPIO_KEY_2:     PRESSED_KEY2,
	WAVESHARE_OLED_GPIO_KEY_3:     PRESSED_KEY3,
	WAVESHARE_OLED_GPIO_JOY_UP:    PRESSED_UP,
	WAVESHARE_OLED_GPIO_JOY_DOWN:  PRESSED_DOWN,
	WAVESHARE_OLED_GPIO_JOY_LEFT:  PRESSED_LEFT,
	WAVESHARE_OLED_GPIO_JOY_RIGHT: PRESSED_RIGHT,
	WAVESHARE_OLED_GPIO_JOY_PRESS: PRESSED_STICK,
}

var function2string = map[WaveshareOledEvent]string{
	PRESSED_KEY1:  "KEY 1",
	PRESSED_KEY2:  "KEY 2",
	PRESSED_KEY3:  "KEY 3",
	PRESSED_UP:    "UP",
	PRESSED_DOWN:  "DOWN",
	PRESSED_LEFT:  "LEFT",
	PRESSED_RIGHT: "RIGHT",
	PRESSED_STICK: "STICK",
}

type WaveshareOled struct {
	pins       []*gpio.Pin
	gpioOpened bool
}

func (w *WaveshareOled) GetHandler(event WaveshareOledEvent) (func(*gpio.Pin)) {
	timeLastEvent = time.Now()
	return func(pin *gpio.Pin) {
		if str,okay :=  function2string[event]; okay {
			if time.Since(timeLastEvent) > 200*time.Millisecond {
				fmt.Printf("%s pressed\n", str)
				timeLastEvent = time.Now()

				// send to output channel if needed here
			}
			//fmt.Printf("%+v\n", pin)
		} else {
			fmt.Println("Unknown GPIO input")
		}
	}
}

func (w *WaveshareOled) Start() (err error) {
	err = gpio.Open()
	if err != nil {
		return
	}

	w.pins = make([]*gpio.Pin, len(waveshareRelevantEvents))
	for idx, evtPinNum := range waveshareRelevantEvents {
		pin := gpio.NewPin(evtPinNum)
		w.pins[idx] = pin
		pin.Input()
		pin.PullUp()
		pin.Watch(gpio.EdgeFalling, w.GetHandler(pin2function[evtPinNum]))
	}
	return nil
}

func (w *WaveshareOled) Stop() {
	for _, p := range w.pins {
		p.Unwatch()
		p.PullNone()
	}
	gpio.Close()
}
