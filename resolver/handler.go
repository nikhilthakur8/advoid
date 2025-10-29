package resolver

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/upstreams"
)

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
func HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	for _, question := range r.Question {
		log.Printf("Received query for %s\n", question.Name)
		qname := strings.ToLower(strings.TrimSuffix(question.Name, "."))
		if blockedDomains[qname] {
			log.Printf("Blocked domain requested: %s\n", qname)
			m := new(dns.Msg)
			m.SetReply(r)
			m.Rcode = dns.RcodeNameError
			w.WriteMsg(m)
			return
		}
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

	for _, question := range msg.Question {
		log.Printf("Received DOH query for %s\n", question.Name)
		qName := strings.ToLower(strings.TrimSuffix(question.Name, "."))
		if blockedDomains[qName] {
			log.Printf("Blocked domain requested via DOH: %s\n", qName)
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
