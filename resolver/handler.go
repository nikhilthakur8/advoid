package resolver

import (
	"io"
	"log"
	"net/http"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/upstreams"
)

func HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
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
