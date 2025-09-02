package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	appErr "web-analyzer-go/internal/errors"
)

const requestIDHeader = "X-Request-ID"

func generateRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// RequestID ensures every request has an ID in context and response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(requestIDHeader)
		if reqID == "" {
			reqID = generateRequestID()
		}
		w.Header().Set(requestIDHeader, reqID)
		ctx := context.WithValue(r.Context(), appErr.RequestIDKey{}, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetReqID returns the request ID from context if present.
func GetReqID(ctx context.Context) string {
	return appErr.GetRequestID(ctx)
}
