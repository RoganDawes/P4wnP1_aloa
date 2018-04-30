//package led
package main

import(
	"os"
	"log"
	"io/ioutil"
	"time"

	pb "../proto"
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

func initLed(led_on bool) (error) {
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
	return nil
}

func led_loop() {
	blink_count := 10
	for {
		for i := 0; i <= blink_count; i++ {
			ioutil.WriteFile(LED_BRIGHTNESS_PATH, []byte(LED_ON), os.ModePerm)
			time.Sleep(LED_DELAY_ON)
			ioutil.WriteFile(LED_BRIGHTNESS_PATH, []byte(LED_OFF), os.ModePerm)
			time.Sleep(LED_DELAY_OFF)
		}
		time.Sleep(LED_DELAY_PAUSE)
	}
}

func setLed(s pb.LEDSettings) (error) {
	
	
	return nil
}

func main() {
	initLed(false)
	
	log.Println("testing led")
	settings := pb.LEDSettings{}
	setLed(settings)
	
	go led_loop()
	time.Sleep(10 * time.Second)
}
