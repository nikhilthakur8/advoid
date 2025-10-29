package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/miekg/dns"
	"github.com/nikhilthakur8/advoid/resolver"
)

func main() {
	log.Printf("I am running")

	dns.HandleFunc(".", resolver.HandleDNSRequest)

	// We are not starting traditional DNS servers for now (Due to no server)
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

	// DNS over HTTPS server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			resolver.HandleDOHRequest(w, r)
			return
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello from Nikhil Thakur"))
			return
		}
	})

	// DNS over TLS server
	cert, err := tls.LoadX509KeyPair("fullchain.pem", "privkey.pem")
	if err != nil {
		log.Fatalf("DOT is not working")
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	tcpServer := &dns.Server{
		Addr:      ":853",
		Net:       "tcp-tls",
		TLSConfig: tlsConfig,
	}

	go func() {
		log.Println("Starting DoT server on :853")
		if err := tcpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DoT server: %v", err)
		}
	}()

	go func() {
		log.Println("Starting DoH server on :8053 (HTTPS)")
		if err := http.ListenAndServe(":8053", nil); err != nil {
			log.Fatalf("Failed to start DoH server: %v", err)
		}
	}()

	select {}
}
