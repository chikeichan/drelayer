package social

import (
	"database/sql"
	"ddrp-relayer/restmodels"
	"ddrp-relayer/store"
	"ddrp-relayer/user"
	"ddrp-relayer/web"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
)

type Service struct {
	DB *sql.DB
}

func (s *Service) CreatePost(w http.ResponseWriter, r *http.Request) {
	params := new(restmodels.PostParams)
	if err := web.DeserializeParams(w, r, params); err != nil {
		return
	}
	if err := ValidatePostParams(params); err != nil {
		web.WriteError(w, r, err, 422, err.Error())
		return
	}
	currentUser := user.AuthPrincipal(r)

	var post *Post
	err := store.WithTransaction(s.DB, func(tx *sql.Tx) error {
		p, err := CreatePost(
			tx,
			currentUser.ID,
			params.Body,
			params.Title,
			params.Reference,
			params.Topic,
			params.Tags,
		)
		if err != nil {
			return errors.Wrap(err, "error creating post")
		}
		post = p
		return nil
	})
	if err != nil {
		web.WriteError(w, r, err, 500, "error creating post")
		return
	}

	web.WriteJSON(w, r, &restmodels.PostResponse{
		Body:      post.Body,
		GUID:      post.GUID,
		ID:        float64(post.ID),
		Reference: store.NilOrString(post.Reference),
		Refhash:   post.Refhash,
		Tags:      post.Tags,
		Timestamp: float64(post.CreatedAt.Unix()),
		Tld:       currentUser.TLD,
		Topic:     store.NilOrString(post.Topic),
		Username:  currentUser.Username,
	})
}

func (s *Service) CreateConnection(w http.ResponseWriter, r *http.Request) {
	params := new(restmodels.ConnectionParams)
	if err := web.DeserializeParams(w, r, params); err != nil {
		return
	}
	if err := ValidateConnectionParams(params); err != nil {
		web.WriteError(w, r, err, 422, err.Error())
		return
	}
	currentUser := user.AuthPrincipal(r)

	var conn *Connection
	err := store.WithTransaction(s.DB, func(tx *sql.Tx) error {
		var (
			c   *Connection
			err error
		)
		switch params.Type {
		case "FOLLOW":
			c, err = CreateFollow(tx, currentUser.ID, params.ConnecteeSubdomain, params.ConnecteeTld)
		case "BLOCK":
			c, err = CreateBlock(tx, currentUser.ID, params.ConnecteeSubdomain, params.ConnecteeTld)
		default:
			err = errors.New("invalid connection type")
		}
		if err != nil {
			return errors.Wrap(err, "error creating post")
		}
		conn = c
		return nil
	})
	if err != nil {
		web.WriteError(w, r, err, 500, "error creating connection")
		return
	}

	web.WriteJSON(w, r, &restmodels.ConnectionResponse{
		ConnecteeSubdomain: store.NilOrString(conn.ConnecteeSubdomain),
		ConnecteeTld:       conn.ConnecteeTLD,
		GUID:               conn.GUID,
		ID:                 float64(conn.ID),
		Refhash:            conn.Refhash,
		Timestamp:          float64(conn.CreatedAt.Unix()),
		Tld:                conn.TLD,
		Type:               params.Type,
		Username:           conn.Username,
	})
}

func (s *Service) CreateModeration(w http.ResponseWriter, r *http.Request) {
	params := new(restmodels.ModerationParams)
	if err := web.DeserializeParams(w, r, params); err != nil {
		return
	}
	if err := ValidateModerationParams(params); err != nil {
		web.WriteError(w, r, err, 422, err.Error())
		return
	}
	currentUser := user.AuthPrincipal(r)

	var mod *Moderation
	err := store.WithTransaction(s.DB, func(tx *sql.Tx) error {
		var (
			m   *Moderation
			err error
		)
		switch params.Type {
		case "LIKE":
			m, err = CreateLike(tx, currentUser.ID, params.Reference)
		case "PIN":
			m, err = CreatePin(tx, currentUser.ID, params.Reference)
		default:
			err = errors.New("invalid connection type")
		}
		if err != nil {
			return errors.Wrap(err, "error creating post")
		}
		mod = m
		return nil
	})
	if err != nil {
		web.WriteError(w, r, err, 500, "error creating connection")
		return
	}

	web.WriteJSON(w, r, &restmodels.ModerationResponse{
		GUID:      mod.GUID,
		ID:        float64(mod.ID),
		Reference: mod.Reference,
		Refhash:   mod.Refhash,
		Timestamp: float64(mod.CreatedAt.Unix()),
		Tld:       mod.TLD,
		Type:      params.Type,
		Username:  mod.Username,
	})
}

func (s *Service) Mount(r *mux.Router) {
	r.Handle("/posts", user.APITokenAuthHandlerMW(s.DB, http.HandlerFunc(s.CreatePost))).Methods("POST")
	r.Handle("/connections", user.APITokenAuthHandlerMW(s.DB, http.HandlerFunc(s.CreateConnection))).Methods("POST")
	r.Handle("/moderations", user.APITokenAuthHandlerMW(s.DB, http.HandlerFunc(s.CreateModeration))).Methods("POST")
}
