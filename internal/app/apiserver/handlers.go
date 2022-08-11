package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"Gotcha/internal/app/model"
	"Gotcha/internal/app/storage"
	"github.com/asaskevich/govalidator"
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
	errNotPermitted   = errors.New("not permitted")
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
			Uptime:  time.Since(handlerRegisteredTime),
		}

		srv.respond(w, r, http.StatusOK, serverStatus)
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
		Title string `json:"title" valid:"required"`
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
		if passedValidation, err := govalidator.ValidateStruct(req); !passedValidation {
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
		BoardID   uuid.UUID   `json:"board_id" valid:"required"`
		Relations []uuid.UUID `json:"relations" valid:"required"`
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

		if passedValidation, err := govalidator.ValidateStruct(req); !passedValidation {
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

func (srv *GotchaAPIServer) permitBoard() http.HandlerFunc {
	type permitRequest struct {
		Description string    `json:"description" valid:"required"`
		BoardID     uuid.UUID `json:"board_id"    valid:"required"`
		UserID      uuid.UUID `json:"user_id"     valid:"required"`
		Permission  string    `json:"permission"  valid:"required"`
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		req := permitRequest{}
		var permission model.PrivilegeType

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

		if passedValidation, err := govalidator.ValidateStruct(req); !passedValidation {
			srv.error(writer, request, http.StatusBadRequest, err)
			return
		}

		// Switch over permission type
		switch req.Permission {
		case "ro":
			permission = model.PrivilegeReadOnly
		case "rw":
			permission = model.PrivilegeReadWrite
		default:
			srv.error(writer, request, http.StatusBadRequest, errors.New("incorrect permission"))
			return
		}

		board, err := srv.storage.Board().GetBoardInfo(req.BoardID)
		// Trigger on incorrect boards: nested & unreal
		if err != nil || board.Base.ID != req.BoardID {
			srv.error(writer, request, http.StatusBadRequest, errors.New("incorrect board"))
			return
		}

		for _, rel := range board.U2BRelations {
			bp, err := srv.storage.Board().GetPrivilegeFromRelation(rel)
			if err != nil {
				// We sure that board is correct, but relations are broken
				srv.error(writer, request, http.StatusInternalServerError, err)
				return
			}

			// Only owner can permit other users to work with table
			if bp.UserID == user.ID && bp.Privilege == model.PrivilegeAuthor {
				relation, err := srv.storage.Board().CreateRelation(bp.BoardID, req.UserID, req.Description, permission)
				if err != nil {
					continue
				}

				// Added the relation
				response := fmt.Sprintf("Granted user:%v %s access to board:%v as %v",
					req.UserID, req.Permission, req.BoardID, relation)
				srv.respond(writer, request, http.StatusOK, response)
				return
			}
		}
		srv.error(writer, request, http.StatusForbidden, errNotPermitted)
	}
}

func (srv *GotchaAPIServer) listUsersHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userWrapped := request.Context().Value(ctxVerifiedUserKey)
		user, converted := userWrapped.(model.User) // Just check if user is specified

		if !converted {
			srv.error(writer, request, http.StatusUnauthorized, errUnauthorized)
			return
		}

		users, err := srv.storage.User().GetAllUsers(&user)
		if err != nil {
			srv.error(writer, request, http.StatusInternalServerError, err)
			return
		}
		srv.respond(writer, request, http.StatusOK, users)
	}
}
