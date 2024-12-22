package nameserver

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
	"sync"
)

var (
	ip      = "0.0.0.0"
	dnsbase = "s.flm.me.uk"
	wg      = sync.WaitGroup{}
)

func Cache(queryname string) {
	txListener, err := net.ListenPacket("udp4", "0.0.0.0:34123")

	if err != nil {
		log.Fatalf("failed to listen on UDP 34123 (for tx) %s", err.Error())
	}

	rxListener, err := net.ListenPacket("udp4", fmt.Sprintf("%s:53", ip))
	if err != nil {
		log.Fatalf("failed to listen on UDP 53 %s", err.Error())
	}

	wg.Add(1)
	go DNSLoop(rxListener)

	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{
		Name:   fmt.Sprintf("%s.%s.", queryname, dnsbase),
		Qtype:  dns.TypeTXT,
		Qclass: dns.ClassINET,
	}
	dnspacket, _ := m1.Pack()

	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:53", ip))

	txListener.WriteTo(dnspacket, addr)

	wg.Wait()
}

func DNSLoop(socket net.PacketConn) {
	for {
		dnsin := make([]byte, 1500)
		inbytes, inaddr, err := socket.ReadFrom(dnsin)

		inmsg := &dns.Msg{}

		if unpackErr := inmsg.Unpack(dnsin[0:inbytes]); unpackErr != nil {
			log.Printf("Unable to unpack DNS request %s", err.Error())
			continue
		}

		if len(inmsg.Question) != 1 {
			log.Printf("More than one quesion in query (%d), droppin %+v", len(inmsg.Question), inmsg)
			continue
		}

		iqn := strings.ToLower(inmsg.Question[0].Name)

		fmt.Println(iqn)

		if !strings.Contains(iqn, dnsbase) {
			log.Printf("question is not for us '%s' vs expected '%s'", iqn, dnsbase)
			continue
		}

		outmsg := &dns.Msg{}

		queryname := strings.Replace(iqn, fmt.Sprintf(".%s.", dnsbase), "", 1)
		log.Printf("Inbound query for chunk '%+v'", queryname)

		ttl := uint32(2147483646)
		content := ""

		ostring := make([]string, 1)
		ostring[0] = content

		outmsg.Id = inmsg.Id
		outmsg = inmsg.SetReply(outmsg)

		outmsg.Answer = make([]dns.RR, 1)
		outmsg.Answer[0] = &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   iqn,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    ttl},
			Txt: ostring,
		}
		outputb, err := outmsg.Pack()

		if err != nil {
			log.Printf("unable to pack response to thing")
			continue
		}

		socket.WriteTo(outputb, inaddr)

		wg.Done()
	}
}
