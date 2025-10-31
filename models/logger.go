package models

type LogDNSQuery struct {
	Level       string  `json:"level"`
	Message     string  `json:"message"`
	Domain      string  `json:"domain,omitempty"`
	QueryType   string  `json:"query_type,omitempty"`
	ClientIP    string  `json:"client_ip,omitempty"`
	Resolver    string  `json:"resolver,omitempty"`
	Blocked     bool    `json:"blocked,omitempty"`
	ResolveTime float64 `json:"resolve_time_ms,omitempty"`
	Timestamp   string  `json:"timestamp"`
}
