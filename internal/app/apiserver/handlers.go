package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/google/uuid"
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
			srv.error(w, r, http.StatusInternalServerError, err)
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
			srv.error(writer, request, http.StatusBadRequest, err)
			return
		}

		tmpUser := model.User{
			Username: rReq.Username,
			Email:    rReq.Email,
			Password: rReq.Password,
		}

		if err := srv.storage.User().SaveUser(&tmpUser); err != nil {

			// Hide real error, because it contains sensitive information
			if errors.Is(err, storage.ErrEntityDuplicate) {
				err = errUserExists
			} else {
				err = errInvalidUser
			}
			srv.error(writer, request, http.StatusUnprocessableEntity, err)
			return
		}

		creationMessage := fmt.Sprintf("%s was created successfully", tmpUser.Username)
		srv.respond(writer, request, http.StatusOK, creationMessage)
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
			srv.error(writer, request, http.StatusBadRequest, err)
			return
		}

		user, err := srv.storage.User().FindUserBySobriquet(lReq.Sobriquet)
		if err != nil || !user.IsCorrectPassword(lReq.Password) {
			srv.error(writer, request, http.StatusUnauthorized, errMixedIncorrect)
			return
		}

		session, err := srv.cookieStore.Get(request, sessionName)
		if err != nil {
			srv.error(writer, request, http.StatusInternalServerError, err)
			return
		}
		session.Values["user_id"] = user.ID.String()
		if err := srv.cookieStore.Save(request, writer, session); err != nil {
			srv.error(writer, request, http.StatusInternalServerError, err)
		}
		srv.respond(writer, request, http.StatusOK, nil)
	}
}

// takes user from authorizationMiddleware
func (srv *GotchaAPIServer) getBoardsHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userWrapped := request.Context().Value(ctxVerifiedUserKey)
		user, converted := userWrapped.(model.User)
		if !converted {
			srv.error(writer, request, http.StatusUnauthorized, errUnauthorized)
			return
		}
		boards, err := srv.storage.Board().GetRootBoardsOfUser(&user)
		if err != nil {
			srv.error(writer, request, http.StatusInternalServerError, err)
			return
		}
		srv.respond(writer, request, http.StatusOK, boards)
	}
}

func (srv *GotchaAPIServer) newRootBoardHandler() http.HandlerFunc {
	type newBoardRequest struct {
		Title string `json:"title"`
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		// Get user and title
		req := newBoardRequest{}
		userWrapped := request.Context().Value(ctxVerifiedUserKey)
		user, converted := userWrapped.(model.User)
		if !converted {
			srv.error(writer, request, http.StatusUnauthorized, errUnauthorized)
			return
		}

		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			srv.error(writer, request, http.StatusBadRequest, err)
			return
		}

		board, err := srv.storage.Board().NewRootBoard(&user, req.Title)
		if err != nil {
			srv.error(writer, request, http.StatusUnprocessableEntity, err)
			return
		}
		srv.respond(writer, request, http.StatusOK, board)
	}
}

func (srv *GotchaAPIServer) deleteRootBoardHandler() http.HandlerFunc {
	type deleteRequest struct {
		BoardID   uuid.UUID   `json:"board_id"`
		Relations []uuid.UUID `json:"relations"`
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		req := deleteRequest{}
		userWrapped := request.Context().Value(ctxVerifiedUserKey)
		user, converted := userWrapped.(model.User)

		if !converted {
			srv.error(writer, request, http.StatusUnauthorized, errUnauthorized)
			return
		}

		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			srv.error(writer, request, http.StatusBadRequest, err)
			return
		}

		// Perform delete operation
		if err := srv.storage.Board().DeleteRootBoard(req.BoardID, req.Relations, &user); err != nil {
			if err == storage.ErrSecurityError {
				srv.error(writer, request, http.StatusUnauthorized, err)
			} else if err == storage.ErrNotFound {
				srv.error(writer, request, http.StatusUnprocessableEntity, err)
			} else {
				srv.error(writer, request, http.StatusInternalServerError, err)
			}
			return
		}

		// Then user isn't an author
		srv.respond(writer, request, http.StatusOK, nil)
	}
}
