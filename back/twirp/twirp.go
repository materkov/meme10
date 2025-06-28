package twirp

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Error) Error() string { return e.Msg }

func WriteError(w http.ResponseWriter, err error) {
	te, ok := err.(*Error)
	if !ok {
		te = &Error{Code: "internal", Msg: err.Error()}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(te)
}

type Server interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}
