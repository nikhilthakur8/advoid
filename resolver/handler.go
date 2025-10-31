package resolver

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/controllers"
	"github.com/nikhilthakur8/advoid/models"
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

func logDNSRequest(domain, qtype, clientIP string, blocked bool, start time.Time) {
	duration := time.Since(start).Milliseconds()

	entry := models.LogDNSQuery{
		Level:       "info",
		Message:     "DNS Query Resolved",
		Domain:      domain,
		QueryType:   qtype,
		ClientIP:    clientIP,
		Resolver:    "DoH",
		Blocked:     blocked,
		ResolveTime: float64(duration),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	go func() {
		if err := controllers.LogDnsQuery(entry); err != nil {
			fmt.Println("Error logging DNS query:", err)
		}
	}()
}

func HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	start := time.Now()
	for _, question := range r.Question {
		log.Printf("Received query for %s\n", question.Name)
		qname := strings.ToLower(strings.TrimSuffix(question.Name, "."))
		if blockedDomains[qname] {
			log.Printf("Blocked domain requested: %s\n", qname)
			m := new(dns.Msg)
			m.SetReply(r)
			m.Rcode = dns.RcodeNameError
			w.WriteMsg(m)
			logDNSRequest(question.Name, dns.TypeToString[question.Qtype], w.RemoteAddr().String(), true, start)
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
	logDNSRequest(r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype], w.RemoteAddr().String(), false, start)
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
			logDNSRequest(question.Name, dns.TypeToString[question.Qtype], r.RemoteAddr, true, time.Now())
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
	logDNSRequest(msg.Question[0].Name, dns.TypeToString[msg.Question[0].Qtype], r.RemoteAddr, false, time.Now())
}
