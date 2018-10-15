// +build linux

package hid

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
	"math"
	"time"
)

const (
	BUTTON1 = byte(1 << 0)
	BUTTON2 = byte(1 << 1)
	BUTTON3 = byte(1 << 2)
)

//mat.Round() doesn't exist before  go 1.10
func round(f float64) float64 {
	return math.Floor(f + .5)
}

type Mouse struct {
	lastChangeWasAbsolute bool
	buttons [3]bool
	axis [2]int
	devicePath string

	deviceFile *os.File
}

func NewMouse(devicePath string) (mouse *Mouse, err error) {
	//ToDo: check existence of deviceFile (+ is writable)
	mouse = &Mouse{
		devicePath: devicePath,
	}

	mouse.deviceFile, err = os.OpenFile(devicePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil,err
	}
	return
}

func (m *Mouse) Close() {
	if m.deviceFile != nil {
		m.deviceFile.Close()
	}
}

func (m *Mouse) writeReport(filepath string) error {
	report, err := generateMouseReport(m.lastChangeWasAbsolute, m.buttons, m.axis)
	if err != nil { return err }
	//fmt.Printf("Writing %+v to %s\n", report, filepath)
	return ioutil.WriteFile(filepath, report, os.ModePerm) //Serialize Report and write to specified file
}

func (m *Mouse) writeReportToFile(file *os.File) (err error) {
	data, err := generateMouseReport(m.lastChangeWasAbsolute, m.buttons, m.axis)
	if err != nil { return err }
	//fmt.Printf("Writing %+v to %s\n", report, file)

	n, err := file.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	return err
}

func (m* Mouse) SetButtons(bt1,bt2,bt3 bool) (err error) {
	change := false
	if m.buttons[0] != bt1 {
		m.buttons[0] = bt1
		change = true
	}
	if m.buttons[1] != bt2 {
		m.buttons[1] = bt2
		change = true
	}
	if m.buttons[2] != bt3 {
		m.buttons[2] = bt3
		change = true
	}

	if change {
		m.lastChangeWasAbsolute = false
		m.axis[0] = 0 //No (repeated) movement on button change
		m.axis[1] = 0 //No (repeated) movement on button change
		return  m.writeReportToFile(m.deviceFile)
	} else {
		//no state change, no new mouse report
		return nil
	}
}

func (m* Mouse) Click(bt1,bt2,bt3 bool) (err error) {
	m.SetButtons(bt1,bt2,bt3)
	m.SetButtons(false,false,false ) //release all button (including other buttons in pressed state, before doing the click)
	return
}


func (m* Mouse) DoubleClick(bt1,bt2,bt3 bool) (err error) {
	m.Click(bt1,bt2,bt3)
	time.Sleep(100 * time.Millisecond) // delay between clicks
	m.Click(bt1,bt2,bt3)
	return
}

func (m* Mouse) Move(x,y int8) (err error) {
	m.axis[0] = int(x)
	m.axis[1] = int(y)
	m.lastChangeWasAbsolute = false
	return m.writeReportToFile(m.deviceFile)
}


func scaleAbs(fVal float64) int {
	ival := int(float64(0x7FFF) * fVal)
	//ival -= 32768
	if ival < -32768 { ival = -32768 }
	if ival > 32767 { ival = 32767 }
	return ival
}

func (m* Mouse) MoveTo(x,y float64) (err error) {
	m.axis[0] = scaleAbs(x)
	m.axis[1] = scaleAbs(y)
	m.lastChangeWasAbsolute = true
	return m.writeReportToFile(m.deviceFile)
}


func (m* Mouse) MoveStepped(x,y int16) (err error) {
	xf := float64(x)
	yf := float64(y)
	steps := math.Max(math.Abs(xf), math.Abs(yf))
	dx := xf / steps
	dy := yf / steps

	curX := int16(0)
	curY := int16(0)

	for curStep := 1; curStep <= int(steps); curStep++ {
		desiredX := int16(round(dx * float64(curStep)))
		desiredY := int16(round(dy * float64(curStep)))

		stepX := desiredX - curX
		stepY := desiredY - curY

		//start Lock here
		m.axis[0] = int(stepX)
		m.axis[1] = int(stepY)
		m.lastChangeWasAbsolute = false
		err =  m.writeReportToFile(m.deviceFile)
		if err != nil {
			m.axis[0] = 0
			m.axis[1] = 0
			//unlock
			return err
		}
		//unlock
		curX += stepX
		curY += stepY
	}
	//Lock
	m.axis[0] = 0
	m.axis[1] = 0
	//Unlock

	return nil
}



func generateMouseReport(absolute bool, buttons [3]bool, axis [2]int) (report []byte, err error) {
	var outdata [6]byte
	if absolute {
		outdata[0] = 0x02
	} else {
		outdata[0] = 0x01
	}
	if buttons[0] { outdata[1] |= BUTTON1 }
	if buttons[1] { outdata[1] |= BUTTON2 }
	if buttons[2] { outdata[1] |= BUTTON3 }
	if absolute {
		binary.LittleEndian.PutUint16(outdata[2:], uint16(axis[0]))
		binary.LittleEndian.PutUint16(outdata[4:], uint16(axis[1]))
	} else {
		outdata[2] = uint8(axis[0])
		outdata[3] = uint8(axis[1])
	}
	return outdata[:], nil
}



