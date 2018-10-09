package mnetlink

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"math/rand"
	"os"
	"sync/atomic"
	"syscall"
	"time"
)

type Client struct {
	Family int
	sock_fd int
	seq uint32
	pid uint32
}

func NewNl(family int) (res *Client, err error) {
	res = &Client{
		Family: family,
	}

	// random start seq
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	res.seq = r.Uint32()
	// assign current PID as PortID
	res.pid = uint32(os.Getpid())


	return res,nil
}

func (c *Client) incSeq() {
	atomic.AddUint32(&c.seq,1)
}

func (c *Client) Open() (err error) {
	// if family is 0, choose NETLINK_USERSOCK by default
	if c.Family == 0 { c.Family = unix.NETLINK_USERSOCK }

	c.sock_fd,err = unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, c.Family)
	if err != nil { return }


	// bind socket
	err = unix.Bind(c.sock_fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: 0,
		Pid: uint32(os.Getpid()),
	})

	return
}

func (c *Client) Close() (err error) {
	unix.Close(c.sock_fd)
	return
}

func (c *Client) AddGroupMembership(groupid int) (err error) {
	err = syscall.SetsockoptInt(c.sock_fd, unix.SOL_NETLINK, unix.NETLINK_ADD_MEMBERSHIP, groupid)
	return
}

func (c *Client) DropGroupMembership(groupid int) (err error) {
	err = syscall.SetsockoptInt(c.sock_fd, unix.SOL_NETLINK, unix.NETLINK_DROP_MEMBERSHIP, groupid)
	return
}

func (c *Client) Sendmsg(p, oob []byte, to unix.Sockaddr, flags int) (err error) {
	err = unix.Sendmsg(c.sock_fd, p, oob, to, flags)

	return
}

func (c *Client) Send(msg Message) (err error) {
	// adjust seq
	msg.Seq = c.seq
	msg.Pid = c.pid


	raw,err := msg.MarshalBinary()
	if err != nil { return }

	addr := &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
	}

	//fmt.Printf("Sending raw2:\n%+v\n", hex.Dump(raw))
	err = c.Sendmsg(raw, nil, addr, 0)
	if err == nil { c.incSeq() }
	return err
}

func (c *Client) Receive() (msgs []Message, err error) {
	for {
		//fmt.Println("Reading")
		raw_in := c.Read()
		for len(raw_in) > unix.NLMSG_HDRLEN {
			msg := Message{}
			msg.UnmarshalBinary(raw_in)

			if msg.IsTypeError() {
				return nil,errors.New(fmt.Sprintf("Error response: %+v %+v", msg.GetData(), msg.GetErrNo()))
			}

			msgs = append(msgs, msg)
			//fmt.Printf("Received raw: \n%v\n", hex.Dump(raw_in))
			//fmt.Printf("Received %+v\nData:\n%+v\n", msg, hex.Dump(msg.data))

			// check if last message
			if msg.IsTypeDone() || !msg.HasFlagMulti() {
				return
			}

			raw_in = raw_in[AlignMsg(int(msg.Len)):]
		}

	}
	return
}
/*
func (c *Client) Receive2() (msgs []Message, err error) {
	for {
		//fmt.Println("Reading")
		raw_in := c.Read()
		msg := Message{}
		msg.UnmarshalBinary(raw_in)

		if msg.IsTypeError() {
			return nil,errors.New(fmt.Sprintf("Error response: %+v %+v", msg.GetData(), msg.GetErrNo()))

		}

		msgs = append(msgs, msg)
		fmt.Printf("Received raw: \n%v\n", hex.Dump(raw_in))
		fmt.Printf("Received %+v\nData:\n%+v\n", msg, hex.Dump(msg.data))

		// check if last message
		if msg.IsTypeDone() || !msg.HasFlagMulti() {
			break
		}
	}
	return
}
*/
func (c *Client) Read() (res []byte) {
	rcvBuf := make([]byte, os.Getpagesize())
	for {
		//fmt.Println("calling receive")

		// peek into rcv to fetch bytes available
		n,_,_ := unix.Recvfrom(c.sock_fd, rcvBuf, unix.MSG_PEEK)
		//fmt.Println("Bytes received: ", n)

		if len(rcvBuf) < n {
			fmt.Println("Receive buffer too small, increasing...")
			rcvBuf = make([]byte, len(rcvBuf)*2)
		} else {
			break
		}
	}
	n,_,_ := unix.Recvfrom(c.sock_fd, rcvBuf, 0)

	nlMsgRaw := make([]byte, n)
	copy(nlMsgRaw, rcvBuf) // Copy over as many bytes as readen

	return nlMsgRaw
	/*
	msg := NlMsg{}
	msg.fromWire(nlMsgRaw)
	return msg.Payload
	*/
}

