package controllers

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/definitions"
	"github.com/nikhilthakur8/advoid/internal/filters"
	"github.com/nikhilthakur8/advoid/services"
	"github.com/nikhilthakur8/advoid/utils"
)

func HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	var userConfig definitions.UserConfig

	// extracting SNI(server name indication) from the tls connection (Gemini)
	if tlsWriter, ok := w.(interface{ ConnectionState() *tls.ConnectionState }); ok {
		state := tlsWriter.ConnectionState()
		if state != nil && state.ServerName != "" {
			sni := state.ServerName
			sniParts := strings.Split(sni, ".")
			if len(sniParts) > 3 {
				userId := sniParts[0]
				userConfig, _ = services.GetUserConfigFromCache(userId)
			}
		}
	}

	if filters.CheckForBlockedDomain(r.Question, userConfig) {
		log.Printf("Blocked domain requested via DOT: %v", r.Question[0].Name)
		m := new(dns.Msg)
		m.SetReply(r)
		m.Rcode = dns.RcodeNameError
		w.WriteMsg(m)
		return
	}

	resp := utils.QueryUpstream(r)
	if resp == nil {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	w.WriteMsg(resp)
}

func HandleDOHRequest(w http.ResponseWriter, r *http.Request) {

	// Get Context Value of userConfig
	ctxValue := r.Context().Value("userConfig")

	var userConfig definitions.UserConfig
	if ctxValue != nil {
		if uc, ok := ctxValue.(definitions.UserConfig); ok {
			userConfig = uc
		}
	}

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

	if filters.CheckForBlockedDomain(msg.Question, userConfig) {
		log.Printf("Blocked domain requested via DOH: %v", msg.Question[0].Name)
		m := new(dns.Msg)
		m.SetReply(&msg)
		m.Rcode = dns.RcodeNameError
		packedResp, err := m.Pack()
		if err != nil {
			http.Error(w, "Failed to pack response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/dns-message")
		w.WriteHeader(http.StatusOK)
		w.Write(packedResp)
		return
	}

	resp := utils.QueryUpstream(&msg)
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
