package export

import (
	"database/sql"
	"ddrp-relayer/protocol"
	"ddrp-relayer/user"
	"github.com/gorilla/mux"
	"net/http"
)

type Service struct {
	DB         *sql.DB
	Client     protocol.DDRPClient
	Signer     protocol.Signer
	ServiceKey string
}

func (s *Service) Export(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := ExportTLDs(s.DB, s.Client, s.Signer); err != nil {
			logger.Error("error exporting TLDs", "err", err)
		}
	}()
	w.WriteHeader(200)
}

func (s *Service) Mount(r *mux.Router) {
	r.Handle("/export", user.ServiceKeyAuthHandlerMW(s.ServiceKey, http.HandlerFunc(s.Export))).Methods("POST")
}
