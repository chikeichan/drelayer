package healthcheck

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Service struct {}

func (s *Service) Mount(r *mux.Router) {
	r.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}