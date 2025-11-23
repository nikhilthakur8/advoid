package definitions

type Rule struct {
	Domain string `json:"domain"`
	Active bool   `json:"active"`
}

type UserConfig struct {
	AllowList []Rule `json:"allowList"`
	DenyList  []Rule `json:"denyList"`
}
