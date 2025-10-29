package resolver

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/upstreams"
)

func HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	for _, question := range r.Question {
		log.Printf("Received query for %s\n", question.Name)
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

func HandleDOHRequest(c *gin.Context) {
	req, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}
	var msg dns.Msg

	if err := msg.Unpack(req); err != nil {
		c.String(http.StatusBadRequest, "Invalid wire format")
		return
	}

	resp := upstreams.QueryUpstream(&msg)
	if resp == nil {
		c.String(http.StatusInternalServerError, "Upstream query failed")
		return
	}

	packedResp, err := resp.Pack()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to pack response")
		return
	}

	c.Data(http.StatusOK, "application/dns-message", packedResp)
}
