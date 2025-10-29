package resolver

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/upstreams"
)

var blockedDomains = map[string]bool{
	"ad.doubleclick.net":             true,
	"ads.youtube.com":                true,
	"googleads.g.doubleclick.net":    true,
	"pagead.l.doubleclick.net":       true,
	"pubads.g.doubleclick.net":       true,
	"partnerad.l.doubleclick.net":    true,
	"adservice.google.com":           true,
	"adservice.google.co.in":         true,
	"www.googleadservices.com":       true,
	"gads.g.doubleclick.net":         true,
	"securepubads.g.doubleclick.net": true,
	"static.doubleclick.net":         true,
}

func HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	qname := strings.ToLower(r.Question[0].Name)
	if blockedDomains[qname] {
		log.Printf("Blocked domain requested: %s\n", qname)
		m := new(dns.Msg)
		m.SetReply(r)
		m.Rcode = dns.RcodeNameError
		w.WriteMsg(m)
		return
	}
	for _, question := range r.Question {
		log.Printf("Received query for %s\n", question.Name)
	}

	resp := upstreams.QueryUpstream(r)
	if resp == nil {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}
	w.WriteMsg(resp)
}

func HandleDOHRequest(w http.ResponseWriter, r *http.Request) {
	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	var msg dns.Msg

	if err := msg.Unpack(req); err != nil {
		http.Error(w, "Invalid wire format", http.StatusBadRequest)
		return
	}

	resp := upstreams.QueryUpstream(&msg)
	if resp == nil {
		http.Error(w, "Upstream query failed", http.StatusInternalServerError)
		return
	}

	packedResp, err := resp.Pack()
	if err != nil {
		http.Error(w, "Failed to pack response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/dns-message")
	w.WriteHeader(http.StatusOK)
	w.Write(packedResp)
}
