package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/HMasataka/gate/internal/domain"
)

type stubSessionStore struct {
	session *domain.Session
	err     error
}

func (s *stubSessionStore) Create(_ context.Context, _ *domain.Session) error   { return nil }
func (s *stubSessionStore) Get(_ context.Context, _ string) (*domain.Session, error) {
	return s.session, s.err
}
func (s *stubSessionStore) Delete(_ context.Context, _ string) error            { return nil }
func (s *stubSessionStore) DeleteByUserID(_ context.Context, _ string) error    { return nil }
func (s *stubSessionStore) ListByUserID(_ context.Context, _ string) ([]domain.Session, error) {
	return nil, nil
}

func TestAuthorize_NoCookie_RedirectsToLogin(t *testing.T) {
	h := &OAuthHandler{
		sessions: &stubSessionStore{err: errors.New("not found")},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?client_id=abc&redirect_uri=http://example.com/cb&response_type=code", nil)
	rec := httptest.NewRecorder()

	h.Authorize(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}

	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/api/v1/oauth/login?") {
		t.Errorf("expected redirect to /api/v1/oauth/login, got %s", loc)
	}

	u, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("client_id") != "abc" {
		t.Errorf("expected client_id=abc in redirect, got %s", u.Query().Get("client_id"))
	}
}

func TestAuthorize_InvalidSession_RedirectsToLogin(t *testing.T) {
	h := &OAuthHandler{
		sessions: &stubSessionStore{err: errors.New("session expired")},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?client_id=abc&redirect_uri=http://example.com/cb&response_type=code", nil)
	req.AddCookie(&http.Cookie{Name: "gate_session", Value: "invalid-session"})
	rec := httptest.NewRecorder()

	h.Authorize(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}

	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/api/v1/oauth/login?") {
		t.Errorf("expected redirect to /api/v1/oauth/login, got %s", loc)
	}

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "gate_session" && c.MaxAge < 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected gate_session cookie to be cleared")
	}
}

func TestGetSessionCookie_NoCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := getSessionCookie(req); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestGetSessionCookie_WithCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "gate_session", Value: "session-123"})
	if got := getSessionCookie(req); got != "session-123" {
		t.Errorf("expected session-123, got %s", got)
	}
}

func TestLoginPage_RendersTemplate(t *testing.T) {
	h := &OAuthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/login?client_id=abc&redirect_uri=http://example.com/cb&response_type=code&error=bad+creds", nil)
	rec := httptest.NewRecorder()

	h.LoginPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html content type, got %s", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "bad creds") {
		t.Error("expected error message in body")
	}
	if !strings.Contains(body, `value="abc"`) {
		t.Error("expected client_id hidden field in body")
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	got := buildAuthorizeURL("c1", "http://example.com/cb", "code", "openid", "st", "ch", "S256")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/api/v1/oauth/authorize" {
		t.Errorf("expected /api/v1/oauth/authorize, got %s", u.Path)
	}
	if u.Query().Get("client_id") != "c1" {
		t.Errorf("expected client_id=c1, got %s", u.Query().Get("client_id"))
	}
}

func TestClearSessionCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	clearSessionCookie(rec)

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "gate_session" && c.MaxAge < 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected gate_session cookie to be cleared with negative MaxAge")
	}
}

// Ensure session fixture is valid
var _ = &domain.Session{
	ID:        "sess-1",
	UserID:    "user-1",
	ExpiresAt: time.Now().Add(time.Hour),
}
