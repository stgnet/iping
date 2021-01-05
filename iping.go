package iping

import (
	"bytes"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"os"
	"syscall"
	"time"
)

type Results struct {
	IP         net.IP
	Sent       int
	Received   int
	Response   []time.Duration
	ResponseMs []int64
	Average    time.Duration
}

type Options struct {
	Target string // address to ping
	Count  int    // number of pings to send (default 1)
	IfName string // optional name of interface to bind to
}

func (opt *Options) Ping() (result Results, err error) {

	ipAddr, err := net.ResolveIPAddr("ip", opt.Target)
	if err != nil {
		return
	}
	result.IP = net.ParseIP(ipAddr.IP.String())

	network := "ip4:icmp"
	icmpType := icmp.Type(ipv4.ICMPTypeEcho)
	icmpReply := icmp.Type(ipv4.ICMPTypeEchoReply)

	if result.IP.To4() == nil {
		network = "ip6:ipv6-icmp"
		icmpType = ipv6.ICMPTypeEchoRequest
		icmpReply = ipv6.ICMPTypeEchoReply
	}

	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		DualStack: true,
	}
	if opt.IfName != "" {
		dialer.Control = func(network, address string, c syscall.RawConn) (err error) {
			err1 := c.Control(func(fd uintptr) {
				err = bindInterface(int(fd), opt.IfName)
				if err != nil {
					return
				}
			})
			if err != nil {
				return err
			}
			return err1
		}
	}
	conn, err := dialer.Dial(network, result.IP.String())
	if err != nil {
		return
	}
	defer conn.Close()

	result.Sent = 0

	pid := os.Getpid() & 0xffff
	seq := pid ^ 0xffff

	count := opt.Count
	if count < 1 {
		count = 1
	}
	for count > 0 {
		count--
		seq++

		request := icmp.Message{
			Type: icmpType,
			Code: 0,
			Body: &icmp.Echo{
				ID:   pid,
				Seq:  seq,
				Data: bytes.Repeat([]byte("iPing!"), 10),
			},
		}
		var sendBuf []byte
		sendBuf, err = request.Marshal(nil)
		if err != nil {
			return
		}

		sent := time.Now()
		_, err = conn.Write(sendBuf)
		if err != nil {
			return
		}
		result.Sent++

		for {
			recvBuf := make([]byte, 1500)
			err = conn.SetReadDeadline(sent.Add(time.Second))
			if err != nil {
				return
			}
			_, err = conn.Read(recvBuf)
			if err != nil {
				// this is probably a timeout error, drop out of read loop
				err = nil
				break
			}
			elapsed := time.Now().Sub(sent)

			// bugfix: https://blog.benjojo.co.uk/post/linux-icmp-type-69
			if recvBuf[0] == 0x45 {
				// remove 20 byte IPv4 header in front of icmp
				recvBuf = recvBuf[20:]
			}

			reply, err := icmp.ParseMessage(icmpType.Protocol(), recvBuf)
			if err != nil {
				// fmt.Printf("Error: Packet parse failed: %v", pErr)
				continue
			}

			if reply.Type != icmpReply {
				continue
			}

			body, ok := reply.Body.(*icmp.Echo)
			if !ok {
				continue
			}
			if body.ID != pid || body.Seq != seq {
				// this packet is not our ping reply, ignore it
				continue
			}

			result.Response = append(result.Response, elapsed)
			result.ResponseMs = append(result.ResponseMs, elapsed.Milliseconds())
			result.Received++

			total := time.Duration(0)
			for _, t := range result.Response {
				total += t
			}
			if len(result.Response) > 0 {
				result.Average = total / time.Duration(len(result.Response))
			}
			err = nil

			break
		}
		if count > 0 {
			wait := time.Now().Add(time.Second).Sub(sent)
			if wait > 0 {
				time.Sleep(wait)
			}
		}
	}
	return
}
