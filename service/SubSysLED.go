package service

import (
	pb "github.com/mame82/P4wnP1_aloa/proto"
	"io/ioutil"
	"log"
	"os"
	"sync/atomic"
	"time"
)

const (
	pLED_TRIGGER_PATH = "/sys/class/leds/led0/trigger"
	pLED_BRIGHTNESS_PATH = "/sys/class/leds/led0/brightness"
	pLED_TRIGGER_MANUAL = "none"
	pLED_ON = "0"
	pLED_OFF = "1"
	pLED_DELAY_ON = 200 * time.Millisecond
	pLED_DELAY_OFF = 200 * time.Millisecond
	pLED_DELAY_PAUSE = 500 * time.Millisecond
)

type LedState1 struct {
	Available bool
	IsRunning bool
	BlinkCount *uint32
}

type LedService struct {
	state *LedState1
}

func NewLedService() (res *LedService) {
	res = &LedService{
		state: &LedState1{},
	}
	bc := uint32(0)
	res.state.BlinkCount = &bc

	return res
}

func (l *LedService) led_loop() {
	ioutil.WriteFile(pLED_BRIGHTNESS_PATH, []byte(pLED_ON), os.ModePerm)

	for l.state.IsRunning{
		for i := uint32(0); i < atomic.LoadUint32(l.state.BlinkCount) && l.state.IsRunning; i++ {
			ioutil.WriteFile(pLED_BRIGHTNESS_PATH, []byte(pLED_ON), os.ModePerm)
			time.Sleep(pLED_DELAY_ON)

			//Don't turn off led if blink_count >= 255 (solid)
			if 255 > atomic.LoadUint32(l.state.BlinkCount) {
				ioutil.WriteFile(pLED_BRIGHTNESS_PATH, []byte(pLED_OFF), os.ModePerm)
				time.Sleep(pLED_DELAY_OFF)
			}
		}
		time.Sleep(pLED_DELAY_PAUSE)
	}

	ioutil.WriteFile(pLED_BRIGHTNESS_PATH, []byte(pLED_ON), os.ModePerm)
}

func (l *LedService) Start() error {
	//set trigger of LED to manual
	log.Println("Setting LED to manual trigger ...")
	ioutil.WriteFile(pLED_TRIGGER_PATH, []byte(pLED_TRIGGER_MANUAL), os.ModePerm)
	l.state.IsRunning = true
	go l.led_loop()

	return nil
}

func (l *LedService) Stop() {
	l.state.IsRunning = false
}

func (l *LedService) GetState() *LedState1 {
	return l.state
}

func (l *LedService) DeploySettings(sets *pb.LEDSettings) {
	atomic.StoreUint32(l.state.BlinkCount, sets.BlinkCount)
}

func (LedService) LoadSettings() {
	panic("implement me")
}

func (LedService) StoreSettings() {
	panic("implement me")
}
