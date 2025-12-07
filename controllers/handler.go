package controllers

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/definitions"
	"github.com/nikhilthakur8/advoid/services"
	"github.com/nikhilthakur8/advoid/utils"
)

// This is the list of blocked domains loaded from the blocklist file
var blockedDomains = make(map[string]bool)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	filePath := filepath.Join(cwd, "oisd_big_abp.txt")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening blocklist file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") {
			continue // skip comments and metadata
		}

		// Extract domain from ||domain^ style
		if after, ok := strings.CutPrefix(line, "||"); ok {
			line = after
		}
		if idx := strings.Index(line, "^"); idx != -1 {
			line = line[:idx]
		}

		// Skip complex patterns (regex, wildcards)
		if strings.ContainsAny(line, "/*") {
			continue
		}

		if line != "" {
			blockedDomains[line] = true
		}
	}
}

func CheckIsDomainInDenyList(domain string, userConfig definitions.UserConfig) bool {
	for _, i := range userConfig.DenyList {
		if i.Active == false {
			continue
		}
		deny := strings.ToLower(strings.TrimSpace(i.Domain))
		if domain == deny {
			return true
		}
		if strings.HasSuffix(domain, "."+deny) {
			return true
		}
	}
	return false
}

func CheckIsDomainInAllowList(domain string, userConfig definitions.UserConfig) bool {
	for _, i := range userConfig.AllowList {
		if i.Active == false {
			continue
		}
		allow := strings.ToLower(strings.TrimSpace(i.Domain))
		if domain == allow {
			return true
		}

		if strings.HasSuffix(domain, "."+allow) {
			return true
		}
	}
	return false
}

func CheckForBlockedDomain(questions []dns.Question, userConfig definitions.UserConfig) bool {
	for _, question := range questions {
		domain := strings.ToLower(strings.TrimSuffix(question.Name, "."))

		// Allow List have higher priority than Deny List and Blocklist
		if CheckIsDomainInAllowList(domain, userConfig) {
			return false
		}

		if blockedDomains[domain] {
			return true
		}

		if CheckIsDomainInDenyList(domain, userConfig) {
			return true
		}
	}
	return false
}

func HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	var userConfig definitions.UserConfig

	// extracting SNI(server name indication) from the tls connection (Gemini)
	if tlsWriter, ok := w.(interface{ ConnectionState() *tls.ConnectionState }); ok {
		state := tlsWriter.ConnectionState()
		sni := state.ServerName
		sniParts := strings.Split(sni, ".")
		if len(sniParts) > 3 {
			userId := sniParts[0]
			userConfig, _ = services.GetUserConfigFromCache(userId)
		}
	}

	if CheckForBlockedDomain(r.Question, userConfig) {
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

	if CheckForBlockedDomain(msg.Question, userConfig) {
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
