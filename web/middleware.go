package web

import (
	"context"
	"crypto/rand"
	"ddrp-relayer/log"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var serverLogger = log.WithModule("server")

const (
	XForwardedFor = "X-Forwarded-For"
	RequestIDKey  = "requestID"
)

func RequestIDMW() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), RequestIDKey, newRequestID())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggingMW() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			l := RequestLogger(r)

			var remoteAddr string
			forwardedFor := r.Header.Get(XForwardedFor)
			if forwardedFor != "" {
				remoteAddr = forwardedFor
			} else {
				remoteAddr = r.RemoteAddr
			}

			l.Info("received request", "remote_addr", remoteAddr, "uri", r.RequestURI, "method", r.Method)
			next.ServeHTTP(w, r)
			l.Info("completed request", "duration", time.Since(start))
		})
	}
}

func newRequestID() string {
	buf := make([]byte, 20)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
