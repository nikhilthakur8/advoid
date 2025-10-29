package upstreams

import "github.com/miekg/dns"

const upstream = "192.168.1.1:53"

var upstreams = []string{
	upstream,     // local router
	"1.1.1.1:53", // Cloudflare
	"8.8.8.8:53", // Google
	"9.9.9.9:53", // Quad9
}

func QueryUpstream(r *dns.Msg) *dns.Msg {
	c := new(dns.Client)
	for _, upstream := range upstreams {
		resp, _, err := c.Exchange(r, upstream)
		if err == nil && resp != nil {
			return resp
		}

		// switch to tcp
		c.Net = "tcp"
		resp, _, err = c.Exchange(r, upstream)
		if err == nil && resp != nil {
			return resp
		}
	}
	return nil
}
