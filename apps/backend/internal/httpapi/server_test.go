package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/auth"
	"ux-securities/backend/internal/service/auth"
)

type testAuthService struct {
	registerErr error
	loginErr    error
	logoutCalls int
	logoutErr   error
	healthErr   error
}

func (s *testAuthService) Register(context.Context, domain.Registration) (domain.User, string, time.Time, error) {
	return domain.User{}, "", time.Time{}, s.registerErr
}
func (s *testAuthService) Login(context.Context, domain.Credentials) (domain.User, string, time.Time, error) {
	return domain.User{}, "", time.Time{}, s.loginErr
}
func (*testAuthService) Authenticate(context.Context, string) (domain.Actor, error) {
	return domain.Actor{}, nil
}
func (s *testAuthService) Logout(context.Context, string) error { s.logoutCalls++; return s.logoutErr }
func (s *testAuthService) Health(context.Context) error         { return s.healthErr }

func TestHealthReturnsUnavailableWhenReadinessProbeFails(t *testing.T) {
	handler := New(&testAuthService{healthErr: errors.New("db unavailable")}, false)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /health status = %d, want 503", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"status":"unavailable"`) {
		t.Fatalf("GET /health body = %s, want unavailable status", response.Body.String())
	}
}

func TestHealthReturnsOKWhenReadinessProbeSucceeds(t *testing.T) {
	handler := New(&testAuthService{}, false)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("GET /health = %d %s, want 200 ok", response.Code, response.Body.String())
	}
}

func TestLogoutWithoutCookieIsIdempotent(t *testing.T) {
	service := &testAuthService{}
	handler := New(service, false)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("POST logout status = %d, want 204", response.Code)
	}
	if service.logoutCalls != 0 {
		t.Fatalf("Logout called %d times without a session cookie, want 0", service.logoutCalls)
	}
}

func TestAuthUnexpectedErrorsReturnGenericInternalServerError(t *testing.T) {
	storageErr := errors.New("database unavailable")
	for _, test := range []struct {
		name    string
		path    string
		body    string
		service *testAuthService
	}{
		{name: "register", path: "/api/v1/auth/register", body: `{"username":"trader","email":"trader@example.com","password":"long-enough-pass"}`, service: &testAuthService{registerErr: storageErr}},
		{name: "login", path: "/api/v1/auth/login", body: `{"identity":"trader","password":"long-enough-pass"}`, service: &testAuthService{loginErr: storageErr}},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := New(test.service, false)
			response := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")
			handler.ServeHTTP(response, req)
			if response.Code != http.StatusInternalServerError {
				t.Fatalf("%s status = %d, want 500", test.name, response.Code)
			}
			if !strings.Contains(response.Body.String(), `"code":"internal_error"`) {
				t.Fatalf("%s body = %s, want generic internal_error", test.name, response.Body.String())
			}
		})
	}
}

func TestLoginInvalidIdentityReturnsUnauthorized(t *testing.T) {
	handler := New(&testAuthService{loginErr: auth.ErrCredentials}, false)
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"identity":"missing","password":"long-enough-pass"}`))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("login status = %d, want 401", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_credentials"`) {
		t.Fatalf("login body = %s, want invalid_credentials", response.Body.String())
	}
}
