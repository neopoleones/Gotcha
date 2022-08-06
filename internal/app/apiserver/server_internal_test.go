package apiserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"Gotcha/internal/app/apiserver"
	"Gotcha/internal/app/model"
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

func TestGotchaAPIServer_heartbeat(t *testing.T) {
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
	storage := teststore.New()
	testUser := model.TestUser(t)
	sessionStore := sessions.NewCookieStore([]byte("TestKey"))
	srv := apiserver.NewAPIServer(logger, cfg, storage, sessionStore)

	// For duplicate test
	anotherUser := model.TestUser(t)
	anotherUser.Email += "2"
	anotherUser.Username += "2"
	_ = storage.User().SaveUser(anotherUser)

	testCases := []struct {
		caseName, testFailedDescription string
		payload                         any
		expectedCode                    int
	}{
		{
			caseName: "Valid registration",
			payload: map[string]string{
				"email":    testUser.Email,
				"username": testUser.Username,
				"password": testUser.Password,
			},
			testFailedDescription: "Registration of valid user rejected",
			expectedCode:          http.StatusOK,
		},
		{
			caseName:              "Corrupted body",
			payload:               "I am a sigma's megaHacker",
			testFailedDescription: "Server somehow registered a user from the corrupted body",
			expectedCode:          http.StatusBadRequest,
		},
		{
			caseName: "Duplicate user",
			payload: map[string]string{
				"email":    anotherUser.Email,
				"username": anotherUser.Username,
				"password": anotherUser.Password,
			},
			testFailedDescription: "Server somehow registered a duplicate user",
			expectedCode:          http.StatusUnprocessableEntity,
		},
		{
			caseName: "Incorrect user",
			payload: map[string]string{
				"email":    "abc",
				"username": "test",
				"password": "pass",
			},
			testFailedDescription: "Server somehow registered an incorrect user",
			expectedCode:          http.StatusUnprocessableEntity,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(tc *testing.T) {
			buf := bytes.Buffer{}
			_ = json.NewEncoder(&buf).Encode(&testCase.payload)
			req, _ := http.NewRequest(http.MethodPost, apiserver.ApiSignup.Path, &buf)
			rec := httptest.NewRecorder()

			srv.Router.ServeHTTP(rec, req)
			assert.Equal(tc, rec.Code, testCase.expectedCode, testCase.testFailedDescription)
		})
	}
}

func TestGotchaAPIServer_signin(t *testing.T) {
	storage := teststore.New()
	testUser := model.TestUser(t)
	_ = storage.User().SaveUser(testUser)

	sessionStore := sessions.NewCookieStore([]byte("TestKey"))
	srv := apiserver.NewAPIServer(logger, cfg, storage, sessionStore)

	testCases := []struct {
		caseName, failedTestDescription string
		expectedCode                    int
		payload                         any
	}{
		{
			caseName:              "Valid authorization",
			failedTestDescription: "Failed to login with correct credentials",
			expectedCode:          http.StatusOK,
			payload: map[string]string{
				"sobriquet": testUser.Email,
				"password":  testUser.Password,
			},
		},
		{
			caseName:              "Corrupted body",
			failedTestDescription: "Server somehow authorized a null user",
			expectedCode:          http.StatusBadRequest,
			payload:               "MegaHacker",
		},
		{
			caseName:              "Incorrect password",
			failedTestDescription: "Server somehow authorized a user with incorrect password",
			expectedCode:          http.StatusUnauthorized,
			payload: map[string]string{
				"sobriquet": testUser.Email,
				"password":  testUser.Password + "abc",
			},
		},
		{
			caseName:              "Incorrect sobriquet",
			failedTestDescription: "Server somehow authorized a user with incorrect sobriquet",
			expectedCode:          http.StatusUnauthorized,
			payload: map[string]string{
				"sobriquet": testUser.Email + "abc",
				"password":  testUser.Password,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(tc *testing.T) {
			buf := bytes.Buffer{}
			_ = json.NewEncoder(&buf).Encode(&testCase.payload)

			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, apiserver.ApiAuthorize.Path, &buf)
			srv.Router.ServeHTTP(rec, req)

			assert.Equal(tc, rec.Code, testCase.expectedCode, testCase.failedTestDescription)
		})
	}
}
