package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

const requestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
