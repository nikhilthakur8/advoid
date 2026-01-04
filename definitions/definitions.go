package definitions

type Rule struct {
	Domain string `json:"domain"`
	Active bool   `json:"active"`
}

type UserConfig struct {
	AllowList []Rule `json:"allowList"`
	DenyList  []Rule `json:"denyList"`
}

type DNSLog struct {
	Timestamp string `json:"timestamp"`
	UserId    string `json:"userId"`
	Domain    string `json:"domain"`
	Type      string `json:"type"`
	Action    string `json:"action"`
}