package definitions

import "time"

type Rule struct {
	Domain string `json:"domain"`
	Active bool   `json:"active"`
}

type UserConfig struct {
	UserId    *string `json:"userId"`
	AllowList []Rule  `json:"allowList"`
	DenyList  []Rule  `json:"denyList"`
}

type DNSLog struct {
	Timestamp time.Time `json:"timestamp"`
	UserId    int       `json:"userId"`
	Domain    string    `json:"domain"`
	Type      string    `json:"type"`
	Action    bool      `json:"action"`
}
