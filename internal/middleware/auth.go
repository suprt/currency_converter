package middleware

import "net/http"

func APIKeyAuth(validkey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-KEY")
			if key != validkey {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
