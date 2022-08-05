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
	apiRootPath  = "/api"
	apiNotesPath = apiRootPath + "/knd"

	apiHeartbeat = newApiHandle("/heartbeat", true, "GET", "OPTIONS")
	apiSignup    = newApiHandle("/authority/signup", true, "POST", "OPTIONS")
	apiAuthorize = newApiHandle("/authority/signin", true, "POST", "OPTIONS")

	apiGetBoards = newApiHandle("/boards", false, "GET", "OPTIONS")
)

type serverState int

const (
	stateRunning serverState = iota
	stateMangled
)

type apiHandle struct {
	path    string
	methods []string
}

func newApiHandle(path string, addBase bool, methods ...string) apiHandle {
	if addBase {
		path = apiRootPath + path
	}
	return apiHandle{path: path, methods: methods}
}

// GotchaAPIServer is a container of server-needed interfaces like logging, cookie store and router
type GotchaAPIServer struct {
	state       serverState
	router      *mux.Router
	logger      logging.GotchaLogger
	cfg         *GotchaConfiguration
	storage     storage.Storage
	cookieStore sessions.Store
	// storage store.storage
}

// newAPIServer returns an instance of GotchaAPIServer with registered handlers and middlewares
func newAPIServer(logger logging.GotchaLogger, cfg *GotchaConfiguration, storage storage.Storage, cookieStore sessions.Store) *GotchaAPIServer {
	server := GotchaAPIServer{
		logger:      logger,
		router:      mux.NewRouter(),
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
	srv.router.HandleFunc(apiHeartbeat.path, srv.heartbeatAPIHandler()).Methods(apiHeartbeat.methods...)
	srv.router.HandleFunc(apiSignup.path, srv.signupHandler()).Methods(apiSignup.methods...)
	srv.router.HandleFunc(apiAuthorize.path, srv.signinHandler()).Methods(apiAuthorize.methods...)

	// Authorization middleware enabled`
	noteSubRouter := srv.router.PathPrefix(apiNotesPath).Subrouter()
	noteSubRouter.Use(srv.authorizationMiddleware)
	noteSubRouter.HandleFunc(apiGetBoards.path, srv.getBoardsHandler()).Methods(apiGetBoards.methods...)
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
