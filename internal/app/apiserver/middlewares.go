package apiserver

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type middlewareContextKey int

const (
	ctxVerifiedUserKey middlewareContextKey = iota
)

func (srv *GotchaAPIServer) authorizationMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// Get the uuid and do some sanity checks
		session, err := srv.cookieStore.Get(request, sessionName)
		if err != nil {
			srv.error(writer, http.StatusUnauthorized, errUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"]
		if !ok {
			srv.error(writer, http.StatusUnauthorized, errUnauthorized)
			return
		}

		userUUID, err := uuid.Parse(userID.(string))
		if err != nil {
			srv.error(writer, http.StatusUnauthorized, err)
			return
		}

		user, err := srv.storage.User().FindUserByID(userUUID)
		if err != nil {
			srv.error(writer, http.StatusUnauthorized, errUnauthorized)
			return
		}

		wrappedContext := context.WithValue(request.Context(), ctxVerifiedUserKey, *user)
		handler.ServeHTTP(writer, request.WithContext(wrappedContext))
	})
}
