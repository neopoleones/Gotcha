package apiserver

import (
	"encoding/json"
	"net/http"

	"Gotcha/internal/app/logging"
	"github.com/gorilla/mux"
)

var (
	apiRootPath = "/api"

	apiHeartbeat = newApiHandle("/heartbeat", "GET")
)

type apiHandle struct {
	path string
	methods []string
}

func newApiHandle(path string, methods ...string) apiHandle{
	return apiHandle{path: apiRootPath + path, methods: methods}
}

// GotchaAPIServer is a container of server-needed interfaces like logging, cookie store and router
type GotchaAPIServer struct {
	router *mux.Router
	logger logging.GotchaLogger
	cfg *GotchaConfiguration
	// storage store.storage
}

// newAPIServer returns an instance of GotchaAPIServer with registered handlers and middlewares
func newAPIServer(logger logging.GotchaLogger, cfg *GotchaConfiguration) *GotchaAPIServer {
	server := GotchaAPIServer{
		logger: logger,
		router: mux.NewRouter(),
		cfg: cfg,
	}

	// Register handlers & middlewares
	server.registerHandlers()

	return &server
}

func (srv *GotchaAPIServer) registerHandlers() {
	srv.router.HandleFunc(apiHeartbeat.path, srv.heartbeatAPIHandler()).Methods(apiHeartbeat.methods...)
}

func (srv *GotchaAPIServer) error(w http.ResponseWriter, code int, err error){
	srv.respond(w, code, map[string]string{"error": err.Error()})
}

func (srv *GotchaAPIServer) respond(w http.ResponseWriter, code int, data any) {
	w.WriteHeader(code)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}