package user

import (
	"context"
	"database/sql"
	"ddrp-relayer/web"
	"errors"
	"net/http"
)

const (
	APITokenHeader   = "X-API-Token"
	ServiceKeyHeader = "X-Service-Key"
	AuthPrincipalKey = "authPrincipal"
)

func APITokenAuthHandlerMW(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(APITokenHeader)
		if token == "" {
			web.WriteError(w, r, errors.New("API token not specified"), 401, "invalid API token")
			return
		}

		user, err := GetByAPIToken(db, token)
		if err != nil {
			web.WriteError(w, r, err, 401, "invalid API key")
			return
		}
		ctx := context.WithValue(r.Context(), AuthPrincipalKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ServiceKeyAuthHandlerMW(realServiceKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serviceKey := r.Header.Get(ServiceKeyHeader)
		if serviceKey == "" {
			web.WriteError(w, r, errors.New("service key not specified"), 401, "invalid service key")
			return
		}
		if realServiceKey != serviceKey {
			web.WriteError(w, r, errors.New("invalid service key specified"), 401, "invalid service key")
			return
		}
		lgr := web.RequestLogger(r)
		lgr.Info("authorized service user")
		next.ServeHTTP(w, r)
	})
}

func AuthPrincipal(r *http.Request) *User {
	principal := r.Context().Value(AuthPrincipalKey)
	if principal == nil {
		panic("auth principal called in unauthenticated method")
	}
	return principal.(*User)
}
