package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/nikhilthakur8/advoid/services"
)

func UserConfigMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Example path: /1234/dns-query or /1234
		path := strings.TrimPrefix(r.URL.Path, "/") // remove leading "/"

		if path == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract first segment (user id)
		parts := strings.Split(path, "/")
		userID := parts[0]

		// Load user config
		userConfig, err := services.GetUserConfigFromCache(userID)
		if err != nil {
			// User not found fallback to generic
			next.ServeHTTP(w, r)
			return
		}

		// Attach userConfig to context
		ctx := context.WithValue(r.Context(), "userConfig", userConfig)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
