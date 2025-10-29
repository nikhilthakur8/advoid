package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/resolver"
)

func main() {
	dns.HandleFunc(".", resolver.HandleDNSRequest)
	r := gin.Default()
	r.POST("/", resolver.HandleDOHRequest)

	// updServer := &dns.Server{Addr: ":53", Net: "udp"}
	// tcpServer := &dns.Server{Addr: ":53", Net: "tcp"}

	// go func() {
	// 	log.Printf("Starting DNS server on :53(udp)")
	// 	if err := updServer.ListenAndServe(); err != nil {
	// 		log.Fatalf("Failed to start server: %v\n", err)
	// 	}
	// }()

	// go func() {
	// 	log.Printf("Starting DNS server on :53(tcp)")
	// 	if err := tcpServer.ListenAndServe(); err != nil {
	// 		log.Fatalf("Failed to start server: %v\n", err)
	// 	}
	// }()

	go func() {
		log.Println("Starting DoH server on :443 (HTTPS)")
		certFile := "/etc/letsencrypt/live/avoid.clouly.in/fullchain.pem"
		keyFile := "/etc/letsencrypt/live/avoid.clouly.in/privkey.pem"
		if err := r.RunTLS(":443", certFile, keyFile); err != nil {
			log.Fatalf("Failed to start DoH server: %v", err)
		}
	}()

	select {}
}
