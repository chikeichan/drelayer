package web

import (
	"ddrp-relayer/log"
	"encoding/json"
	"io"
	"net/http"
)

const (
	MaxBodySizeBytes      = 5 * 1024
	ErrMsgBodyParseFailed = "json payload malformed"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func DeserializeParams(w http.ResponseWriter, r *http.Request, proto interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, MaxBodySizeBytes))
	if err := decoder.Decode(proto); err != nil {
		WriteError(w, r, err, 422, ErrMsgBodyParseFailed)
		return err
	}
	return nil
}

func WriteError(w http.ResponseWriter, r *http.Request, err error, statusCode int, publicMsg string) {
	lgr := RequestLogger(r)
	lgr.Error("error in request", "code", statusCode, "err", err)
	errBody := &ErrorResponse{
		Message: publicMsg,
	}
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(errBody); err != nil {
		lgr.Error("failed to encode error body", "err", err)
	}
}

func WriteJSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	lgr := RequestLogger(r)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		lgr.Error("failed to encode response body", "err", err)
		return
	}
}

func RequestLogger(r *http.Request) log.Logger {
	return serverLogger.Sub("id", r.Context().Value(RequestIDKey))
}
