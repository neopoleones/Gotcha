package apiserver

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type middlewareContextKey int

const (
	ctxVerifiedUserKey middlewareContextKey = iota
	ctxRequestIDKey
	ctxStatusCodeKey
)

func getIPAddress(req *http.Request) string {
	ipAddress := req.Header.Get("X-Real-Ip")
	if ipAddress == "" {
		ipAddress = req.Header.Get("X-Forwarded-For")
	}
	if ipAddress == "" {
		ipAddress = req.RemoteAddr
	}
	return ipAddress
}

func (srv *GotchaAPIServer) authorizationMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// Get the uuid and do some sanity checks
		session, err := srv.cookieStore.Get(request, sessionName)
		if err != nil {
			srv.error(writer, request, http.StatusUnauthorized, errUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"]
		if !ok {
			srv.error(writer, request, http.StatusUnauthorized, errUnauthorized)
			return
		}

		userUUID, err := uuid.Parse(userID.(string))
		if err != nil {
			srv.error(writer, request, http.StatusUnauthorized, err)
			return
		}

		user, err := srv.storage.User().FindUserByID(userUUID)
		if err != nil {
			srv.error(writer, request, http.StatusUnauthorized, errUnauthorized)
			return
		}

		wrappedContext := context.WithValue(request.Context(), ctxVerifiedUserKey, *user)
		handler.ServeHTTP(writer, request.WithContext(wrappedContext))
	})
}

func (srv *GotchaAPIServer) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			requestID := uuid.New().String()
			writer.Header().Set("Request-ID", requestID)
			keyContext := context.WithValue(request.Context(), ctxRequestIDKey, requestID)
			next.ServeHTTP(writer, request.WithContext(keyContext))
		})
}

func (srv *GotchaAPIServer) loggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var resultCode int
		requestID := request.Context().Value(ctxRequestIDKey)
		fields := logrus.Fields{
			"Path":       request.RequestURI,
			"Request-ID": requestID,
		}

		srv.logger.WithFields(fields).Infof("Connection from: %s", getIPAddress(request))
		wrappedContext := context.WithValue(request.Context(), ctxStatusCodeKey, &resultCode)
		startTime := time.Now()

		handler.ServeHTTP(writer, request.WithContext(wrappedContext))
		srv.logger.WithFields(fields).Infof("Served for %v. Status code: %d", time.Now().Sub(startTime), resultCode)
	})
}
