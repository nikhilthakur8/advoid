package filters

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/definitions"
	"github.com/nikhilthakur8/advoid/internal/logger"
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

func checkInList(domain string, list []definitions.Rule) bool {
	for _, i := range list {
		if i.Active == false {
			continue
		}

		cleanRuleDomain := strings.ToLower(strings.TrimSpace(i.Domain))
		if domain == cleanRuleDomain {
			return true
		}

		// Check for subdomains
		// like search.nextleet.com should match for rule domain nextleet.com
		if strings.HasSuffix(domain, "."+cleanRuleDomain) {
			return true
		}
	}
	return false
}

func CheckForBlockedDomain(questions []dns.Question, userConfig definitions.UserConfig) bool {
	for _, question := range questions {
		domain := strings.ToLower(strings.TrimSuffix(question.Name, "."))
		// Allow List have higher priority than Deny List and Blocklist
		if checkInList(domain, userConfig.AllowList) {
			dnsLogs := definitions.DNSLog{
				Timestamp: time.Now(),
				Domain:    question.Name,
				UserId:    userConfig.UserId,
				Action:    false,
				Type:      dns.TypeToString[question.Qtype],
			}
			logger.EmitLogs(dnsLogs)
			return false
		}

		// if domain is in blocklist or deny list, block it
		if blockedDomains[domain] || checkInList(domain, userConfig.DenyList) {
			dnsLogs := definitions.DNSLog{
				Timestamp: time.Now(),
				Domain:    question.Name,
				UserId:    userConfig.UserId,
				Action:    true,
				Type:      dns.TypeToString[question.Qtype],
			}
			logger.EmitLogs(dnsLogs)
			return true
		}

	}

	dnsLogs := definitions.DNSLog{
		Timestamp: time.Now(),
		Domain:    questions[0].Name,
		UserId:    userConfig.UserId,
		Action:    true,
		Type:      dns.TypeToString[questions[0].Qtype],
	}
	logger.EmitLogs(dnsLogs)

	return false
}
