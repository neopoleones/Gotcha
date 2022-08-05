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
	ApiRootPath  = "/api"
	ApiNotesPath = ApiRootPath + "/knd"

	ApiHeartbeat = newApiHandle("/heartbeat", true, "GET", "OPTIONS")
	ApiSignup    = newApiHandle("/authority/signup", true, "POST", "OPTIONS")
	ApiAuthorize = newApiHandle("/authority/signin", true, "POST", "OPTIONS")

	ApiGetBoards = newApiHandle("/boards", false, "GET", "OPTIONS")
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
	srv.Router.HandleFunc(ApiHeartbeat.Path, srv.heartbeatAPIHandler()).Methods(ApiHeartbeat.Methods...)
	srv.Router.HandleFunc(ApiSignup.Path, srv.signupHandler()).Methods(ApiSignup.Methods...)
	srv.Router.HandleFunc(ApiAuthorize.Path, srv.signinHandler()).Methods(ApiAuthorize.Methods...)

	// Authorization middleware enabled`
	noteSubRouter := srv.Router.PathPrefix(ApiNotesPath).Subrouter()
	noteSubRouter.Use(srv.authorizationMiddleware)
	noteSubRouter.HandleFunc(ApiGetBoards.Path, srv.getBoardsHandler()).Methods(ApiGetBoards.Methods...)
}

func (srv *GotchaAPIServer) error(w http.ResponseWriter, code int, err error) {
	if code == http.StatusInternalServerError && srv.state == stateRunning {
		srv.state = stateMangled
	}
	srv.respond(w, code, map[string]string{"error": err.Error()})
}

func (srv *GotchaAPIServer) respond(w http.ResponseWriter, code int, data any) {
	w.WriteHeader(code)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
