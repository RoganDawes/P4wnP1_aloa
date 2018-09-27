package dwc2

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"unsafe"
)

var hbo = HostByteOrder()

func Hbo() binary.ByteOrder {
	return hbo
}

type Dwc2Netlink struct {
	sock_fd int
	nl_group int
}

func NewDwc2Nl(mcast_group int) (d* Dwc2Netlink) {
	return &Dwc2Netlink{
		nl_group: mcast_group,
	}
}

func (d *Dwc2Netlink) OpenNlKernelSock() (err error) {
	d.sock_fd,err = unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_USERSOCK)
	if err != nil { return }

	err = unix.Bind(d.sock_fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: 0,
		Pid: uint32(os.Getpid()),
	})
	if err != nil { return }

	err = syscall.SetsockoptInt(d.sock_fd, unix.SOL_NETLINK, unix.NETLINK_ADD_MEMBERSHIP, d.nl_group)

	return
}

func (d *Dwc2Netlink) Close() (err error) {
	return unix.Close(d.sock_fd)
}



func (d *Dwc2Netlink) socketReaderLoop() {
	fmt.Println("Readloop started")
	rcvBuf := make([]byte, 1024)
	for {
		fmt.Println("READER LOOP")

		n, err := syscall.Read(d.sock_fd, rcvBuf)
		if err != nil || n == 0 {
			d.Close()
			fmt.Println("Error reading from socket")
			break
		}
		nlMsg := make([]byte, n)
		copy(nlMsg, rcvBuf) // Copy over as many bytes as readen

		fmt.Printf("Received packet %+v\n", nlMsg)
		/*
		//fmt.Printf("Sending raw event packet to handler loop: %+v\n", nlMsg)
		select {
		case m.newRawPacket <- nlMsg:
			// do nothing
		case <-m.disposeMgmtConnection:
			// unblock and exit the loop if eventHandler is closed
			close(m.newRawPacket)
			break
		}
		*/
	}

	fmt.Println("Socket read loop exited")
}

func (d *Dwc2Netlink) SocketReaderLoop2() {
	fmt.Println("Readloop started")
	rcvBuf := make([]byte, os.Getpagesize())
	for {
		//fmt.Println("calling receive")

		// peek into rcv to fetch bytes available
		n,_,_ := unix.Recvfrom(d.sock_fd, rcvBuf, unix.MSG_PEEK)
		//fmt.Println("Bytes received: ", n)

		if len(rcvBuf) < n {
			fmt.Println("Receive buffer too small, increasing...")
			rcvBuf = make([]byte, len(rcvBuf)*2)
		} else {
			n,_,_ = unix.Recvfrom(d.sock_fd, rcvBuf, 0)

			nlMsgRaw := make([]byte, n)
			copy(nlMsgRaw, rcvBuf) // Copy over as many bytes as readen

			fmt.Printf("Received packet %+v\n", nlMsgRaw)
			msg := NlMsg{}
			msg.fromWire(nlMsgRaw)
			fmt.Printf("Received message %+v\n", msg)
		}
	}

	fmt.Println("Socket read loop exited")
}

func (d *Dwc2Netlink) Read() (res []byte) {
	rcvBuf := make([]byte, os.Getpagesize())
	for {
		//fmt.Println("calling receive")

		// peek into rcv to fetch bytes available
		n,_,_ := unix.Recvfrom(d.sock_fd, rcvBuf, unix.MSG_PEEK)
		//fmt.Println("Bytes received: ", n)

		if len(rcvBuf) < n {
			fmt.Println("Receive buffer too small, increasing...")
			rcvBuf = make([]byte, len(rcvBuf)*2)
		} else {
			break
		}
	}
	n,_,_ := unix.Recvfrom(d.sock_fd, rcvBuf, 0)

	nlMsgRaw := make([]byte, n)
	copy(nlMsgRaw, rcvBuf) // Copy over as many bytes as readen

	msg := NlMsg{}
	msg.fromWire(nlMsgRaw)
	return msg.Payload
}


type NlMsg struct {
	Length  uint32
	Type    uint16
	Flags   uint16
	SeqNum  uint32
	PortID  uint32 // == Process ID for first socket of Proc. If the message comes from kernel (like here) PortID is 0
	Payload []byte
}

// Note: data is formated in host byte order, unless the attribute NLA_F_NET_BYTEORDER is specified
func (m *NlMsg) fromWire(src []byte) {
	m.Length = hbo.Uint32(src[0:4])
	// truncate message to length
	src = src[:m.Length]

	m.Type = hbo.Uint16(src[4:6])
	m.Flags = hbo.Uint16(src[6:8])
	m.SeqNum = hbo.Uint32(src[8:12])
	m.PortID = hbo.Uint32(src[12:16])
	m.Payload = src[16:]
}

func HostByteOrder() (res binary.ByteOrder) {
	i := int(0x0100)
	ptr := unsafe.Pointer(&i)
	if 0x01 == *(*byte)(ptr) {
		return binary.BigEndian
	}
	return binary.LittleEndian
}
