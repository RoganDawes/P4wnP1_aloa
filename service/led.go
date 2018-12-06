package service

import(
	"os"
	"log"
	"io/ioutil"
	"time"
	"sync/atomic"

	pb "github.com/mame82/P4wnP1_aloa/proto"
)

const (
	LED_TRIGGER_PATH = "/sys/class/leds/led0/trigger"
	LED_BRIGHTNESS_PATH = "/sys/class/leds/led0/brightness"
	LED_TRIGGER_MANUAL = "none"
	LED_ON = "0"
	LED_OFF = "1"
	LED_DELAY_ON = 200 * time.Millisecond
	LED_DELAY_OFF = 200 * time.Millisecond
	LED_DELAY_PAUSE = 500 * time.Millisecond
)


type LedState struct {
	blink_count *uint32
}
/*
var (
	blink_count uint32 = 0
)
*/
func NewLed(led_on bool) (ledState *LedState, err error) {
	blinkCount := uint32(0)
	ledState = &LedState{ &blinkCount }

	//set trigger of LED to manual
	log.Println("Setting LED to manual trigger ...")
	ioutil.WriteFile(LED_TRIGGER_PATH, []byte(LED_TRIGGER_MANUAL), os.ModePerm)
	if led_on {
		log.Println("Setting LED to ON ...")
		ioutil.WriteFile(LED_BRIGHTNESS_PATH, []byte(LED_ON), os.ModePerm)
	} else {
		log.Println("Setting LED to OFF ...")
		ioutil.WriteFile(LED_BRIGHTNESS_PATH, []byte(LED_OFF), os.ModePerm)
	}

	go ledState.led_loop() // watcher loop

	ledState.SetLed(GetDefaultLEDSettings()) //set default setting
	return ledState,nil
}

func (leds *LedState) led_loop() {
	
	for {
		for i := uint32(0); i < atomic.LoadUint32(leds.blink_count); i++ {
			ioutil.WriteFile(LED_BRIGHTNESS_PATH, []byte(LED_ON), os.ModePerm)
			time.Sleep(LED_DELAY_ON)
			
			//Don't turn off led if blink_count >= 255 (solid)
			if 255 > atomic.LoadUint32(leds.blink_count) {
				ioutil.WriteFile(LED_BRIGHTNESS_PATH, []byte(LED_OFF), os.ModePerm)
				time.Sleep(LED_DELAY_OFF)
			}
		}
		time.Sleep(LED_DELAY_PAUSE)
	}
}

func (leds *LedState) SetLed(s *pb.LEDSettings) (error) {
	//log.Printf("setLED called with %+v", s)
	
	atomic.StoreUint32(leds.blink_count, s.BlinkCount)
	
	return nil
}

func (leds *LedState) GetLed() (res *pb.LEDSettings, err error) {
	return &pb.LEDSettings{BlinkCount: atomic.LoadUint32(leds.blink_count)}, nil
}

