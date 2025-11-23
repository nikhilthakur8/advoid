package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/nikhilthakur8/advoid/services"
)

func UserConfigMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host

		// remove the Port if present
		hostParts := strings.Split(host, ":")
		hostName := hostParts[0]

		// spilt the host into parts to get the userId

		parts := strings.Split(hostName, ".")

		// CASES
		// 1. 1.dns.clouly.in  // userId specified
		// 2. dns.clouly.in // generic access

		if len(parts) < 3 {
			http.Error(w, "Invalid host format", http.StatusBadRequest)
			return
		}

		// General access
		// if len(parts) == 3 {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// userId := parts[0]
		userId := "2"

		userConfig, err := services.GetUserConfigFromCache(userId)
		
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "userConfig", userConfig)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

	})
}
