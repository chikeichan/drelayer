package user

import (
	"database/sql"
	"ddrp-relayer/restmodels"
	"ddrp-relayer/store"
	"ddrp-relayer/web"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

type Service struct {
	Redis       *redis.Client
	DB          *sql.DB
	GHCfg       *oauth2.Config
	AllowSignup bool
	ServiceKey  string
}

func (s *Service) CreateUsernamePassword(w http.ResponseWriter, r *http.Request) {
	params := new(restmodels.CreateUserParams)
	if err := web.DeserializeParams(w, r, params); err != nil {
		return
	}
	if err := ValidateUserParams(params); err != nil {
		web.WriteError(w, r, err, 422, err.Error())
		return
	}
	var user *User
	err := store.WithTransaction(s.DB, func(tx *sql.Tx) error {
		u, err := CreateUsernamePassword(tx, params.Username, params.Tld, params.Email, params.Password)
		if err != nil {
			return errors.Wrap(err, "error creating user")
		}
		user = u
		return nil
	})
	if err != nil {
		web.WriteError(w, r, err, 500, "error creating user")
		return
	}
	w.WriteHeader(204)
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	var params restmodels.LoginParams
	if err := web.DeserializeParams(w, r, &params); err != nil {
		return
	}
	var (
		token        string
		tokenCreated time.Time
	)
	err := store.WithTransaction(s.DB, func(tx *sql.Tx) error {
		tok, tokCreated, err := Authenticate(tx, params.Username, params.Tld, params.Password)
		if err != nil {
			return errors.Wrap(err, "error authenticating user")
		}
		token = tok
		tokenCreated = tokCreated
		return nil
	})
	if errors.Is(err, ErrInvalidPassword) {
		web.WriteError(w, r, err, 401, "invalid username or password")
		return
	}
	if err != nil {
		web.WriteError(w, r, err, 500, "error authenticating user")
		return
	}
	web.WriteJSON(w, r, &restmodels.TokenResponse{
		Expiry: float64(tokenCreated.Add(24 * time.Hour).Unix()),
		Token:  token,
	})
}

func (s *Service) Mount(r *mux.Router) {
	if s.AllowSignup {
		r.HandleFunc("/users", s.CreateUsernamePassword)
	} else {
		r.Handle("/users", ServiceKeyAuthHandlerMW(s.ServiceKey, http.HandlerFunc(s.CreateUsernamePassword)))
	}
	r.HandleFunc("/login", s.Login).Methods("POST")
}
