package apiserver

import (
	"encoding/json"
	"net/http"

	"Gotcha/internal/app/logging"
	"Gotcha/internal/app/storage"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	ApiRootPath   = "/api"
	ApiBoardsPath = ApiRootPath + "/boards"

	ApiHeartbeat = newApiHandle("/heartbeat", true, "GET")
	ApiSignup    = newApiHandle("/authority/signup", true, "POST")
	ApiAuthorize = newApiHandle("/authority/signin", true, "POST")

	ApiGetBoards       = newApiHandle("/all", false, "GET")
	ApiNewRootBoard    = newApiHandle("/root", false, "POST")
	ApiDeleteRootBoard = newApiHandle("/root", false, "DELETE")
	ApiPermitBoard     = newApiHandle("/permit", false, "POST")
)

type serverState int

const (
	stateRunning serverState = iota
	stateMangled
)

type ApiHandle struct {
	Path    string
	Methods []string
}

func newApiHandle(path string, addBase bool, methods ...string) ApiHandle {
	if addBase {
		path = ApiRootPath + path
	}
	return ApiHandle{Path: path, Methods: methods}
}

// GotchaAPIServer is a container of server-needed interfaces like logging, cookie store and Router
type GotchaAPIServer struct {
	state       serverState
	Router      *mux.Router
	logger      logging.GotchaLogger
	cfg         *GotchaConfiguration
	storage     storage.Storage
	cookieStore sessions.Store
	// storage store.storage
}

// NewAPIServer returns an instance of GotchaAPIServer with registered handlers and middlewares
func NewAPIServer(logger logging.GotchaLogger, cfg *GotchaConfiguration, storage storage.Storage, cookieStore sessions.Store) *GotchaAPIServer {
	server := GotchaAPIServer{
		logger:      logger,
		Router:      mux.NewRouter(),
		cfg:         cfg,
		storage:     storage,
		state:       stateRunning,
		cookieStore: cookieStore,
	}

	// Register handlers & middlewares
	server.registerHandlers()

	return &server
}

func (srv *GotchaAPIServer) registerHandlers() {
	// Authorization not required
	srv.Router.Use(srv.setRequestID)
	srv.Router.Use(srv.loggingMiddleware)
	srv.Router.HandleFunc(ApiHeartbeat.Path, srv.heartbeatAPIHandler()).Methods(ApiHeartbeat.Methods...)
	srv.Router.HandleFunc(ApiSignup.Path, srv.signupHandler()).Methods(ApiSignup.Methods...)
	srv.Router.HandleFunc(ApiAuthorize.Path, srv.signinHandler()).Methods(ApiAuthorize.Methods...)

	// Authorization middleware enabled`
	noteSubRouter := srv.Router.PathPrefix(ApiBoardsPath).Subrouter()
	noteSubRouter.Use(srv.authorizationMiddleware)
	noteSubRouter.HandleFunc(ApiGetBoards.Path, srv.getBoardsHandler()).Methods(ApiGetBoards.Methods...)
	noteSubRouter.HandleFunc(ApiNewRootBoard.Path, srv.newRootBoardHandler()).Methods(ApiNewRootBoard.Methods...)
	noteSubRouter.HandleFunc(ApiDeleteRootBoard.Path, srv.deleteRootBoardHandler()).Methods(ApiDeleteRootBoard.Methods...)
	noteSubRouter.HandleFunc(ApiPermitBoard.Path, srv.permitBoard()).Methods(ApiPermitBoard.Methods...)
}

func (srv *GotchaAPIServer) error(w http.ResponseWriter, request *http.Request, code int, err error) {
	if code == http.StatusInternalServerError && srv.state == stateRunning {
		srv.state = stateMangled
	}
	srv.respond(w, request, code, map[string]string{"error": err.Error()})
}

func (srv *GotchaAPIServer) respond(w http.ResponseWriter, request *http.Request, code int, data any) {
	// HELLCODE: Save status code in context for logger
	*(request.Context().Value(ctxStatusCodeKey).(*int)) = code

	w.WriteHeader(code)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
