package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/resolver"
)

func main() {
	log.Printf("I am running")
	dns.HandleFunc(".", resolver.HandleDNSRequest)
	r := gin.Default()
	r.POST("/", resolver.HandleDOHRequest)
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello from Nikhil Thakur")
	})
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
		log.Println("Starting DoH server on :8053 (HTTPS)")
		if err := r.Run(":8053"); err != nil {
			log.Fatalf("Failed to start DoH server: %v", err)
		}
	}()

	select {}
}
