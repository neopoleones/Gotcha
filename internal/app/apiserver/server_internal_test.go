package apiserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"Gotcha/internal/app/apiserver"
	"Gotcha/internal/app/storage/teststore"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	cfg    = &apiserver.GotchaConfiguration{}
	logger = logrus.New()
)

func TestMain(m *testing.M) {
	cfg.AppName = "Gotcha test"
	os.Exit(m.Run())
}

func TestGotchaAPIServer_hearbeat(t *testing.T) {
	storage := teststore.New()
	encodedPayload := &bytes.Buffer{}
	sessionStore := sessions.NewCookieStore([]byte("TestKey"))
	srv := apiserver.NewAPIServer(logger, cfg, storage, sessionStore)

	// Check that server isn't mangled on create
	rec := httptest.NewRecorder()
	logger.Println(apiserver.ApiHeartbeat.Path)
	req, _ := http.NewRequest(http.MethodGet, apiserver.ApiHeartbeat.Path, encodedPayload)

	srv.Router.ServeHTTP(rec, req)

	status := apiserver.ServerStatus{}
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&status), "Result not in ServerStatus format")
	assert.Equal(t, status.Status, apiserver.StatusServerOK, "Status not OK")
	assert.Equal(t, status.AppName, cfg.AppName, "App name changed")
}

func TestGotchaAPIServer_signup(t *testing.T) {
	// ...
}
