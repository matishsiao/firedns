package main

import (
	"github.com/matishsiao/dns"
	_"log"
	_"net"
	"strings"
)

type DNSHandler struct {
	zones *ZoneStore
}


func NewHandler(zones *ZoneStore) *DNSHandler {
	return &DNSHandler{zones:zones}
}
var pprofrun bool = false
func (h *DNSHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	// BIND does not support answering multiple questions so we won't
	
	//h.zones.m.RLock()
	//defer h.zones.m.RUnlock()
	counter.Total_counter++
	/*if counter.Total_counter == 10000 {
		StopCPUProfile()
	}*/
	/*if len(req.Question) != 1 {
		dns.HandleFailed(w, req)
		return
	} else {
		//Proxy
		if zone, name = h.zones.match(req.Question[0].Name, req.Question[0].Qtype); zone == nil {
			if recurseTo == "" {
				dns.HandleFailed(w, req)
				return
			} else {
				c := new(dns.Client)
			Redo:
				if in, _, err := c.Exchange(req, recurseTo); err == nil { // Second return value is RTT
					if in.MsgHdr.Truncated {
						c.Net = "tcp"
						goto Redo
					}

					w.WriteMsg(in)
					return
				} else {
					log.Printf("Recursive error: %+v %s\n", err,name)
					dns.HandleFailed(w, req)
					return
				}
			}
		}
	}*/

	m := new(dns.Msg)
	m.SetReply(req)
	m.Authoritative = true
	m.RecursionAvailable = true
	m.Compress = true
	bufsize := uint16(512)
	dnssec := false
	//tcp := false

	if req.Question[0].Qtype == dns.TypeANY {
		m.Authoritative = false
		m.Rcode = dns.RcodeRefused
		m.RecursionAvailable = false
		m.RecursionDesired = false
		m.Compress = false
		// if write fails don't care
		w.WriteMsg(m)
		return
	}

	if o := req.IsEdns0(); o != nil {
		bufsize = o.UDPSize()
		dnssec = o.Do()
	}
	if bufsize < 512 {
		bufsize = 512
	}
	// with TCP we can send 64K
	/*if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		bufsize = dns.MaxMsgSize - 1
		tcp = true
	}*/
	req.Question[0].Name = strings.ToLower(req.Question[0].Name)
	msgcache := GetZoneCache(QuestionKey(req.Question[0],dnssec))
	if msgcache == nil {
		//log.Println("name:",req.Question[0].Name)
		var zone *Zone
		var name string
		if zone, name = h.zones.match(req.Question[0].Name, req.Question[0].Qtype); zone == nil {
			m.Authoritative = false
			m.Rcode = dns.RcodeRefused
			m.RecursionAvailable = false
			m.RecursionDesired = false
			m.Compress = false
			// if write fails don't care
			w.WriteMsg(m)
			return
		}
		for _, r := range (*zone)[dns.RR_Header{Name: req.Question[0].Name, Rrtype: req.Question[0].Qtype, Class: req.Question[0].Qclass}] {
			m.Answer = append(m.Answer, r)
		}
		// Add Authority section for NS only
		//if req.Question[0].Qtype == dns.TypeNS {
			for _, r := range (*zone)[dns.RR_Header{Name: name, Rrtype: dns.TypeNS, Class: dns.ClassINET}] {
				m.Ns = append(m.Ns, r)
			
				// Resolve Authority if possible and serve as Extra
				for _, r := range (*zone)[dns.RR_Header{Name: r.(*dns.NS).Ns, Rrtype: dns.TypeA, Class: dns.ClassINET}] {
					m.Extra = append(m.Extra, r)
				}
				for _, r := range (*zone)[dns.RR_Header{Name: r.(*dns.NS).Ns, Rrtype: dns.TypeAAAA, Class: dns.ClassINET}] {
					m.Extra = append(m.Extra, r)
				}
			}
		//}
		SetZoneCache(QuestionKey(req.Question[0],dnssec),m)
		counter.Misscache_counter++
		w.WriteMsg(m)
	} else {
		counter.Cache_counter++
		msgcache.Id = m.Id
		w.WriteMsg(msgcache)
	}
	
	
	
}

func (h *DNSHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *DNSHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}

func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
