package export

import (
	"database/sql"
	"ddrp-relayer/protocol"
	apiv1 "ddrp-relayer/protocol/v1"
	"ddrp-relayer/user"
	"github.com/gorilla/mux"
	"net/http"
)

type Service struct {
	DB         *sql.DB
	Client     apiv1.DDRPv1Client
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
