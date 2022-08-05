package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"Gotcha/internal/app/model"
)

const (
	StatusServerOK      = "working"
	StatusServerMangled = "mangled"

	sessionName = "gotcha_auth"
)

var (
	errInvalidUser    = errors.New("specified user options are invalid")
	errUserExists     = errors.New("user exists")
	errUnauthorized   = errors.New("unauthorized")
	errMixedIncorrect = errors.New("incorrect username or password") // hides out that user not exists
)

type ServerStatus struct {
	AppName string        `json:"app_name"`
	Status  string        `json:"status"`
	Uptime  time.Duration `json:"uptime"`
}

func (srv *GotchaAPIServer) heartbeatAPIHandler() http.HandlerFunc {
	handlerRegisteredTime := time.Now()

	return func(w http.ResponseWriter, r *http.Request) {
		state := "responds"
		switch srv.state {
		case stateRunning:
			state = StatusServerOK
		case stateMangled:
			state = StatusServerMangled
		}

		serverStatus := ServerStatus{
			AppName: srv.cfg.AppName,
			Status:  state,
			Uptime:  time.Now().Sub(handlerRegisteredTime),
		}

		err := json.NewEncoder(w).Encode(serverStatus)
		if err != nil {
			srv.error(w, http.StatusInternalServerError, err)
		}
	}
}

// Authorization

func (srv *GotchaAPIServer) signupHandler() http.HandlerFunc {
	type RegisterRequest struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		rReq := RegisterRequest{}
		if err := json.NewDecoder(request.Body).Decode(&rReq); err != nil {
			srv.error(writer, http.StatusBadRequest, err)
			return
		}

		tmpUser := model.User{
			Username: rReq.Username,
			Email:    rReq.Email,
			Password: rReq.Password,
		}

		if err := srv.storage.User().SaveUser(&tmpUser); err != nil {

			// Hide real error, because it contains sensitive information
			if strings.Contains(err.Error(), "duplicate") {
				err = errUserExists
			} else {
				err = errInvalidUser
			}
			srv.error(writer, http.StatusUnprocessableEntity, err)
			return
		}

		creationMessage := fmt.Sprintf("%s was created successfully", tmpUser.Username)
		srv.respond(writer, http.StatusOK, creationMessage)
	}
}

func (srv *GotchaAPIServer) signinHandler() http.HandlerFunc {
	type loginRequest struct {
		Sobriquet string `json:"sobriquet"`
		Password  string `json:"password"`
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		lReq := loginRequest{}
		if err := json.NewDecoder(request.Body).Decode(&lReq); err != nil {
			srv.error(writer, http.StatusBadRequest, err)
			return
		}

		user, err := srv.storage.User().FindUserBySobriquet(lReq.Sobriquet)
		if err != nil || !user.IsCorrectPassword(lReq.Password) {
			srv.error(writer, http.StatusUnauthorized, errMixedIncorrect)
			return
		}

		session, err := srv.cookieStore.Get(request, sessionName)
		if err != nil {
			srv.error(writer, http.StatusInternalServerError, err)
			return
		}
		session.Values["user_id"] = user.ID.String()
		if err := srv.cookieStore.Save(request, writer, session); err != nil {
			srv.error(writer, http.StatusInternalServerError, err)
		}
		srv.respond(writer, http.StatusOK, nil)
	}
}

// takes user from authorizationMiddleware
func (srv *GotchaAPIServer) getBoardsHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userWrapped := request.Context().Value(ctxVerifiedUserKey)
		if user, ok := userWrapped.(model.User); ok {

			// **PLACEHOLDER**, make teststorage and then continue coding the getBoardsHandler
			srv.respond(writer, http.StatusOK, user)
			return
		}
		srv.error(writer, http.StatusUnauthorized, errUnauthorized)
	}
}
