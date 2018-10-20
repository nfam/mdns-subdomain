package main

import (
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

type connection struct {
	*net.UDPAddr
	*net.UDPConn
	hostname   *string
	stopsServe chan struct{}
}

type pkt struct {
	*dns.Msg
	*net.UDPAddr
}

func listen(ifi *net.Interface, hostname *string) (*connection, error) {
	addr := &net.UDPAddr{
		IP:   net.ParseIP("224.0.0.251"),
		Port: 5353,
	}
	conn, err := net.ListenMulticastUDP("udp4", ifi, addr)
	if err != nil {
		return nil, err
	}
	return &connection{addr, conn, hostname, nil}, nil
}

func (c *connection) serve() {
	in := make(chan pkt, 32)
	stopsRead := make(chan struct{})
	go c.readloop(in, stopsRead)

	stopsServe := make(chan struct{})
	c.stopsServe = stopsServe
	for {
		select {
		case <-stopsServe:
			close(stopsRead)
			return
		case msg := <-in:
			if len(msg.Question) <= 0 {
				continue
			}

			rrs := c.query(msg.Question)
			if len(rrs) <= 0 {
				continue
			}

			msg.MsgHdr.Response = true // convert question to response
			msg.Answer = rrs
			msg.Extra = append(msg.Extra, c.findExtra(msg.Answer...)...)

			// nuke questions
			msg.Question = nil
			if err := c.writeMessage(msg.Msg, msg.UDPAddr); err != nil {
				log.Fatalf("Cannot send: %s", err)
			}
		}
	}
}

func (c *connection) stop() {
	stops := c.stopsServe
	c.stopsServe = nil
	if stops != nil {
		close(stops)
	}
}

func (c *connection) readloop(in chan pkt, stops chan struct{}) {
	for {
		select {
		case <-stops:
			return
		default:
			msg, addr, err := c.readMessage()
			if err != nil {
				// log dud packets
				log.Printf("Could not read message. %s", err)
				continue
			}
			if len(msg.Question) > 0 {
				in <- pkt{msg, addr}
			}
		}
	}
}

// consumes an mdns packet from the wire and decode it
func (c *connection) readMessage() (*dns.Msg, *net.UDPAddr, error) {
	buf := make([]byte, 1500)
	read, addr, err := c.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, err
	}
	msg := new(dns.Msg)
	if err := msg.Unpack(buf[:read]); err != nil {
		return nil, nil, err
	}
	return msg, addr, nil
}

// encodes an mdns msg and broadcast it on the wire
func (c *connection) writeMessage(msg *dns.Msg, addr *net.UDPAddr) error {
	buf, err := msg.Pack()
	if err != nil {
		return err
	}
	_, err = c.WriteToUDP(buf, addr)
	return err
}

// recursively probe for related records
func (c *connection) findExtra(r ...dns.RR) []dns.RR {
	extra := make([]dns.RR, 0)
	for _, rr := range r {
		var q dns.Question
		switch rr := rr.(type) {
		case *dns.PTR:
			q = dns.Question{
				Name:   rr.Ptr,
				Qtype:  dns.TypeANY,
				Qclass: dns.ClassINET,
			}
		case *dns.SRV:
			q = dns.Question{
				Name:   rr.Target,
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			}
		default:
			continue
		}
		if rr, err := c.answer(q); err == nil {
			extra = append(append(extra, rr), c.findExtra(rr)...)
		}
	}
	return extra
}

func (c *connection) query(qs []dns.Question) []dns.RR {
	rrs := make([]dns.RR, 0)
	for _, q := range qs {
		rr, err := c.answer(q)
		if err != nil {
			log.Println(err)
			continue
		}
		if rr == nil {
			continue
		}
		rrs = append(rrs, rr)
	}
	return rrs
}

func (c *connection) answer(q dns.Question) (dns.RR, error) {
	var (
		name string
		err  error
	)
	if c.hostname != nil {
		name = *c.hostname
	} else {
		name, err = hostname()
		if err != nil {
			return nil, err
		}
	}
	name += "."

	if q.Qtype == dns.TypeA || q.Qtype == dns.TypeANY {
		if strings.HasSuffix(q.Name, "."+name) {
			return dns.NewRR(q.Name + " IN CNAME " + name)
		}
	}
	return nil, nil
}
