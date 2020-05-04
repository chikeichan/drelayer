package healthcheck

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Service struct{}

func (s *Service) Mount(r *mux.Router) {
	r.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}
