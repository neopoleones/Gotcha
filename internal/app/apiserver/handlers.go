package apiserver

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	statusServerOK = "working"
)

func (srv *GotchaAPIServer) heartbeatAPIHandler() http.HandlerFunc {
	handlerRegisteredTime := time.Now()

	type ServerStatus struct {
		AppName string		   `json:"app_name"`
		Status string	       `json:"status"`
		Uptime time.Duration   `json:"uptime"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		serverStatus := ServerStatus {
			AppName: srv.cfg.AppName,
			Status: statusServerOK,
			Uptime: time.Now().Sub(handlerRegisteredTime),
		}

		err := json.NewEncoder(w).Encode(serverStatus)
		if err != nil {
			srv.error(w, http.StatusInternalServerError, err)
		}
	}
}