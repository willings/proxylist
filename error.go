package proxylist

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	ACCESS_DENIED = 1

	NO_PROVIDER = 100
)

type jsonError struct {
	ErrCode    int    `json:"errCode"`
	ErrMessage string `json:"errMessage"`
}

func (err *jsonError) toJson() string {
	ret, e := json.Marshal(err)
	if e == nil && ret != nil {
		return string(ret)
	} else {
		return ""
	}
}

func errJson(w http.ResponseWriter, errCode int) {
	w.Header().Set("Content-Type", "application/json")
	err := &jsonError{ErrCode: errCode, ErrMessage: ""}
	fmt.Fprint(w, err.toJson())
}

func formatError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/json")
	e := &jsonError{ErrCode:0, ErrMessage: err.Error()}
	fmt.Fprint(w, e.toJson())
}
