package service

import (
	"sync"
	pb "github.com/mame82/P4wnP1_go/proto"
)

type GpioPinState byte
const (
	GPIO_PIN_STATE_DISABLED GpioPinState = iota
	GPIO_PIN_STATE_INPUT
	GPIO_PIN_STATE_OUTPUT
	//GPIO_PIN_STATE_PWM
	//GPIO_PIN_STATE_CLOCK
)

type GpioPin struct {
	*sync.Mutex
	Number pb.GPIONum
	InUse bool
	State GpioPinState
	EdgeDetectState pb.GPIOInEdge
	PullUpDown pb.GPIOInPullUpDown
}

var PI_GPIO_NUMS = []pb.GPIONum{
	pb.GPIONum_NUM_2, //SDA1
	pb.GPIONum_NUM_3, //SCL1
	pb.GPIONum_NUM_4,
	pb.GPIONum_NUM_5,
	pb.GPIONum_NUM_6,
	pb.GPIONum_NUM_7, //CE1
	pb.GPIONum_NUM_8, //CE0
	pb.GPIONum_NUM_9, //MISO
	pb.GPIONum_NUM_10, //MOSI
	pb.GPIONum_NUM_11, //SCLK
	pb.GPIONum_NUM_12,
	pb.GPIONum_NUM_13,
	pb.GPIONum_NUM_14, //TXD0
	pb.GPIONum_NUM_15, //RXD0
	pb.GPIONum_NUM_16,
	pb.GPIONum_NUM_17,
	pb.GPIONum_NUM_18,
	pb.GPIONum_NUM_19,
	pb.GPIONum_NUM_20,
	pb.GPIONum_NUM_21,
	pb.GPIONum_NUM_22,
	pb.GPIONum_NUM_23,
	pb.GPIONum_NUM_24,
	pb.GPIONum_NUM_25,
	pb.GPIONum_NUM_26,
	pb.GPIONum_NUM_27,
}
/*
var PI_GPIO_NUMS = []pb.GPIONum{
	pb.GPIONum_NUM_2, //SDA1
	pb.GPIONum_NUM_3, //SCL1
	pb.GPIONum_NUM_4,
	pb.GPIONum_NUM_17,
	pb.GPIONum_NUM_27,
	pb.GPIONum_NUM_22,
	pb.GPIONum_NUM_10, //MOSI
	pb.GPIONum_NUM_9, //MISO
	pb.GPIONum_NUM_11, //SCLK
	pb.GPIONum_NUM_5,
	pb.GPIONum_NUM_6,
	pb.GPIONum_NUM_13,
	pb.GPIONum_NUM_19,
	pb.GPIONum_NUM_26,
	pb.GPIONum_NUM_14, //TXD0
	pb.GPIONum_NUM_15, //RXD0
	pb.GPIONum_NUM_18,
	pb.GPIONum_NUM_23,
	pb.GPIONum_NUM_24,
	pb.GPIONum_NUM_25,
	pb.GPIONum_NUM_8, //CE0
	pb.GPIONum_NUM_7, //CE1
	pb.GPIONum_NUM_12,
	pb.GPIONum_NUM_16,
	pb.GPIONum_NUM_20,
	pb.GPIONum_NUM_21,
}
*/
type GpioManager struct {
	availableGpioPins []byte
}
