package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	fs := flag.NewFlagSet("upornot", flag.ExitOnError)

	targetIP := "127.0.0.1"
	fs.StringVar(&targetIP, "target", targetIP, "IP address to ping")

	interval := time.Second
	fs.DurationVar(&interval, "interval", interval, "interval between ping attempts")

	address := "127.0.0.1:4444"
	fs.StringVar(&address, "addr", address, "server address")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range sigs {
			log.Printf("received %v, exiting", sig)
			cancel()
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		logerr(serveHTTP(ctx, address), "http server")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logerr(ping(ctx, targetIP, interval), "ping")
	}()

	wg.Wait()
}

func logerr(err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}

func serveHTTP(ctx context.Context, address string) error {
	log.Printf("HTTP server started (%s)", address)
	defer log.Printf("HTTP server stopped (%s)", address)

	server := http.Server{Addr: address}

	done := make(chan error)
	go func() {
		defer close(done)
		done <- server.ListenAndServe()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		log.Printf("Shutdown HTTP server")
		return server.Shutdown(ctx)
	}
}

func ping(ctx context.Context, target string, interval time.Duration) error {
	log.Printf("Ping %s every %v", target, interval)

	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		go func() { logerr(pingOnce(ctx, target), target) }()

		select {
		case <-t.C:
			// next ping!
		case <-ctx.Done():
			return nil
		}
	}
}

func pingOnce(ctx context.Context, target string) error {
	start := time.Now()
	result := "indeterminate"
	defer func() { log.Printf("%s: %s in %v", target, result, time.Since(start)) }()

	// adapted from https://github.com/paulstuart/ping/blob/master/ping.go
	c, err := net.Dial("ip4:icmp", target)
	if err != nil {
		result = "connect error"
		return err
	}
	c.SetDeadline(time.Now().Add(2 * time.Second))
	defer c.Close()

	typ := icmpv4EchoRequest
	xid, xseq := os.Getpid()&0xffff, 1
	wb, err := (&icmpMessage{
		Type: typ,
		Code: 0,
		Body: &icmpEcho{
			ID:   xid,
			Seq:  xseq,
			Data: bytes.Repeat([]byte("Go Go Gadget Ping!!!"), 3),
		},
	}).Marshal()
	if err != nil {
		return err
	}
	//log.Printf("echo req")
	if _, err = c.Write(wb); err != nil {
		result = "send error"
		return err
	}

	var m *icmpMessage
	rb := make([]byte, 20+len(wb))
	for {
		if _, err = c.Read(rb); err != nil {
			result = "read error"
			return err
		}
		rb = ipv4Payload(rb)
		if m, err = parseICMPMessage(rb); err != nil {
			result = "parse error"
			return err
		}
		//log.Printf("echo resp %v", m.Type)
		switch m.Type {
		case icmpv4EchoRequest, icmpv6EchoRequest:
			continue
		}
		break
	}
	result = "OK"
	return nil
}

const (
	icmpv4EchoRequest = 8
	icmpv4EchoReply   = 0
	icmpv6EchoRequest = 128
	icmpv6EchoReply   = 129
)

type icmpMessage struct {
	Type     int             // type
	Code     int             // code
	Checksum int             // checksum
	Body     icmpMessageBody // body
}

type icmpMessageBody interface {
	Len() int
	Marshal() ([]byte, error)
}

// Marshal returns the binary enconding of the ICMP echo request or
// reply message m.
func (m *icmpMessage) Marshal() ([]byte, error) {
	b := []byte{byte(m.Type), byte(m.Code), 0, 0}
	if m.Body != nil && m.Body.Len() != 0 {
		mb, err := m.Body.Marshal()
		if err != nil {
			return nil, err
		}
		b = append(b, mb...)
	}
	switch m.Type {
	case icmpv6EchoRequest, icmpv6EchoReply:
		return b, nil
	}
	csumcv := len(b) - 1 // checksum coverage
	s := uint32(0)
	for i := 0; i < csumcv; i += 2 {
		s += uint32(b[i+1])<<8 | uint32(b[i])
	}
	if csumcv&1 == 0 {
		s += uint32(b[csumcv])
	}
	s = s>>16 + s&0xffff
	s = s + s>>16
	// Place checksum back in header; using ^= avoids the
	// assumption the checksum bytes are zero.
	b[2] ^= byte(^s & 0xff)
	b[3] ^= byte(^s >> 8)
	return b, nil
}

// parseICMPMessage parses b as an ICMP message.
func parseICMPMessage(b []byte) (*icmpMessage, error) {
	msglen := len(b)
	if msglen < 4 {
		return nil, errors.New("message too short")
	}
	m := &icmpMessage{Type: int(b[0]), Code: int(b[1]), Checksum: int(b[2])<<8 | int(b[3])}
	if msglen > 4 {
		var err error
		switch m.Type {
		case icmpv4EchoRequest, icmpv4EchoReply, icmpv6EchoRequest, icmpv6EchoReply:
			m.Body, err = parseICMPEcho(b[4:])
			if err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}

// imcpEcho represenets an ICMP echo request or reply message body.
type icmpEcho struct {
	ID   int    // identifier
	Seq  int    // sequence number
	Data []byte // data
}

func (p *icmpEcho) Len() int {
	if p == nil {
		return 0
	}
	return 4 + len(p.Data)
}

// Marshal returns the binary enconding of the ICMP echo request or
// reply message body p.
func (p *icmpEcho) Marshal() ([]byte, error) {
	b := make([]byte, 4+len(p.Data))
	b[0], b[1] = byte(p.ID>>8), byte(p.ID&0xff)
	b[2], b[3] = byte(p.Seq>>8), byte(p.Seq&0xff)
	copy(b[4:], p.Data)
	return b, nil
}

// parseICMPEcho parses b as an ICMP echo request or reply message body.
func parseICMPEcho(b []byte) (*icmpEcho, error) {
	bodylen := len(b)
	p := &icmpEcho{ID: int(b[0])<<8 | int(b[1]), Seq: int(b[2])<<8 | int(b[3])}
	if bodylen > 4 {
		p.Data = make([]byte, bodylen-4)
		copy(p.Data, b[4:])
	}
	return p, nil
}

func ipv4Payload(b []byte) []byte {
	if len(b) < 20 {
		return b
	}
	hdrlen := int(b[0]&0x0f) << 2
	return b[hdrlen:]
}
