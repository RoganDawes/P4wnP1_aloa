// +build linux

package main

import (
	"errors"
	"fmt"
	"github.com/mame82/P4wnP1_go/service/peripheral"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
)

/*
while $true; do cat /sys/kernel/debug/20980000.usb/regdump | grep HPRT0; sleep 1; done

Host Port Control and Status register
=====================================
Reg value is dumpable via debugfs for DesignWare USB 2.0 host controller (dwc2)
The dump could be triggered via debugff by reading `/sys/kernel/debug/20980000.usb/regdump` which holds the value for the register in `HPRT0`.
The problem with the approach: The register dump has to be triggered by polling /sys/kernel/debug/20980000.usb/regdump and dumps ALL registers.
This seems to crash the UDC from time to time, in contrast to the comment of the kernel source (https://github.com/raspberrypi/linux/blob/rpi-4.14.y/drivers/usb/dwc2/debugfs.c#L402).

HPRT --> Host Port Control and Status Register, details on data struct:
https://www.cl.cam.ac.uk/~atm26/ephemeral/rpi/dwc_otg/doc/html/unionhprt0__data.html#a964274b5d22e89ca4490f66dff3c763


Meanings of struct fields:
https://www.intel.com/content/www/us/en/programmable/documentation/sfo1410144425160/sfo1410067646785/sfo1410069362932/sfo1410069623936/sfo1410069628257/sfo1410069409998.html

01741 typedef union hprt0_data {
01743         uint32_t d32;
01745         struct {
01746                 unsigned prtconnsts:1;
01747                 unsigned prtconndet:1;
01748                 unsigned prtena:1;
01749                 unsigned prtenchng:1;

01750                 unsigned prtovrcurract:1;
01751                 unsigned prtovrcurrchng:1;
01752                 unsigned prtres:1;
01753                 unsigned prtsusp:1;

01754                 unsigned prtrst:1;
01755                 unsigned reserved9:1;
01756                 unsigned prtlnsts:2;

01757                 unsigned prtpwr:1;
01758                 unsigned prttstctl:4;
01759                 unsigned prtspd:2;
01760 #define DWC_HPRT0_PRTSPD_HIGH_SPEED 0
01761 #define DWC_HPRT0_PRTSPD_FULL_SPEED 1
01762 #define DWC_HPRT0_PRTSPD_LOW_SPEED      2
01763                 unsigned reserved19_31:13;
01764         } b;
01765 } hprt0_data_t;

Test results (polling /sys/kernel/debug/20980000.usb/regdump and parsing "HPRT0")
=================================================================================

USB gadget running, but OTG adapter attached, NO device connected to the OTG adapter
--> prtlnsts = 0 (No data line on positive logic level)
--> prtpwr = 1 (the port is powered ... at least if there's no overcorrency)
--> prtconnsts = 0 (No device attached)

USB gadget running, but OTG adapter attached, device connected to the OTG adapter
--> prtlnsts = 0,1 or 2 (changes)
--> prtpwr = 1 (the port is powered ... at least if there's no overcorrency)
--> prtconnsts = 1 (A device is attached)
--> prtspd = 1 or 2 (for non high speed devices)

USB gadget running, no OTG adapter attached, but NOT connected to host as a peripheral
--> prtlnsts = 1 (Logic Level of D+ is 1, logic level of D- is 0)
--> prtpwr = 0
--> prtconnsts = 0 (No device attached)

USB gadget running, no OTG adapter attached, connected to host as a peripheral
--> prtlnsts = 0 (both, D+ and D-, are indicated as low)
--> prtpwr = 0
--> prtconnsts = 0 (No device attached)
--> in fact the whole HPRT0 is set to 0x00000000 if connected to a host in device mode



WARNING: Constant reading from the regdump file sometimes crashes UDC */

type USBConnectionState int

const (
	USB_CONNECTION_STATE_UNKNOWN                         = 0
	USB_CONNECTION_STATE_PERIPHERAL_ATTACHED_TO_HOST     = 1
	USB_CONNECTION_STATE_PERIPHERAL_NOT_ATTACHED_TO_HOST = 2
	USB_CONNECTION_STATE_OTG_NO_DEVICE_ATTACHED          = 3
	USB_CONNECTION_STATE_OTG_DEVICE_ATTACHED             = 4
)

func (cs USBConnectionState) String() string {
	names := [...]string{
		"Unknown",
		"Peripheral mode, connected to host",
		"Peripheral mode, not connected to host",
		"OTG mode, no device attached",
		"OTG mode, some device(s) attached",
	}

	if cs < USB_CONNECTION_STATE_UNKNOWN || cs > USB_CONNECTION_STATE_OTG_DEVICE_ATTACHED {
		return names[0]
	}

	return names[cs]
}

// Host Port Control and Status Register
type HprtData struct {
	RegVal uint32

	PrtConnSts bool //0
	PrtConnDet bool //1
	PrtEnA     bool //2
	PrtEnChng  bool //3

	PrtOvrCurrAct  bool //4
	PrtOvrCurrChng bool //5
	PrtRes         bool //6
	PrtSusp        bool //7

	PrtRst bool //8
	// Reserved9 bool //9
	PrtLnSts uint8 // 10 + 11

	PrtPwr    bool  //12
	PrtTstCtl uint8 //13,14,15,16

	PrtSpd uint8 //17,18

	ConState USBConnectionState
	// #define DWC_HPRT0_PRTSPD_HIGH_SPEED 0
	// #define DWC_HPRT0_PRTSPD_FULL_SPEED 1
	// #define DWC_HPRT0_PRTSPD_LOW_SPEED      2
	// unsigned reserved19_31:13;
}

func (tgt *HprtData) FromUint32(src uint32) {
	tgt.RegVal = src

	tgt.PrtConnSts = src&1 > 0
	tgt.PrtConnDet = src&(1<<1) > 0
	tgt.PrtEnA = src&(1<<2) > 0
	tgt.PrtEnChng = src&(1<<3) > 0

	tgt.PrtOvrCurrAct = src&(1<<4) > 0
	tgt.PrtOvrCurrChng = src&(1<<5) > 0
	tgt.PrtRes = src&(1<<6) > 0
	tgt.PrtSusp = src&(1<<7) > 0

	tgt.PrtRst = src&(1<<8) > 0
	tgt.PrtLnSts = uint8((src >> 10) & 0x3)

	tgt.PrtPwr = src&(1<<12) > 0
	tgt.PrtTstCtl = uint8((src >> 13) & 0xF)
	tgt.PrtSpd = uint8((src >> 17) & 0x3)

	switch {
	case tgt.PrtPwr && !tgt.PrtConnSts:
		tgt.ConState = USB_CONNECTION_STATE_OTG_NO_DEVICE_ATTACHED
	case tgt.PrtPwr && tgt.PrtConnSts:
		tgt.ConState = USB_CONNECTION_STATE_OTG_DEVICE_ATTACHED
	case tgt.RegVal == 0:
		tgt.ConState = USB_CONNECTION_STATE_PERIPHERAL_ATTACHED_TO_HOST
	case !tgt.PrtPwr && tgt.PrtLnSts > 0:
		tgt.ConState = USB_CONNECTION_STATE_PERIPHERAL_NOT_ATTACHED_TO_HOST
	default:
		tgt.ConState = USB_CONNECTION_STATE_UNKNOWN
	}
}

func ReadState() (regval *HprtData, err error) {
	cont, err := ioutil.ReadFile("/sys/kernel/debug/20980000.usb/regdump") // This read dumps all regs, including HPRT0
	if err != nil {
		return regval, err
	}
	scont := string(cont)

	reHprt0 := regexp.MustCompile("(?m)HPRT0 = 0x([0-9]{8})")
	strReg := ""
	if rRes := reHprt0.FindStringSubmatch(scont); len(rRes) > 1 {
		strReg = rRes[1]
	} else {
		return regval, errors.New("String for register HPRT0 couldn't be extracted from sysfs")
	}

	regValInt, err := strconv.ParseUint(strReg, 16, 32)
	if err != nil {
		return regval, errors.New(fmt.Sprintf("Couldn't parse HPRT0 value '%s' to int32\n", strReg))
	}

	regval = &HprtData{}
	regval.FromUint32(uint32(regValInt))

	return regval, nil
}



func main() {
	wo := peripheral.WaveshareOled{}
	wo.Start()
	/*
	dwc := dwc2.NewDwc2Nl(24)
	err := dwc.OpenNlKernelSock()

	if err == nil {
		dwc.SocketReaderLoop2()
	} else {
		fmt.Println("Err: ", err)
	}
	dwc.Close()
	*/


	/*
	i := 0
	for {
		res, err := ReadState()
		if err == nil {
			fmt.Printf("%d %+v\n ", i, res)
		} else {
			fmt.Println(err)
		}
		time.Sleep(time.Millisecond * 200)
		i++
	}
	*/

	/*
	w := service.NewDwc2ConnectWatcher()
	w.Start()
	*/

	fmt.Println("Stop with SIGTERM or SIGINT")
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	si := <-sig
	fmt.Printf("Signal (%v) received, ending P4wnP1_service ...\n", si)
	wo.Stop()
	//w.Stop()

}
