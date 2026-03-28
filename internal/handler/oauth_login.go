package handler

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"
)

//go:embed templates/login.html
var loginFS embed.FS

var loginTmpl = template.Must(template.ParseFS(loginFS, "templates/login.html"))

type loginPageData struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Error               string
}

func (h *OAuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := loginPageData{
		ClientID:            r.URL.Query().Get("client_id"),
		RedirectURI:         r.URL.Query().Get("redirect_uri"),
		ResponseType:        r.URL.Query().Get("response_type"),
		Scope:               r.URL.Query().Get("scope"),
		State:               r.URL.Query().Get("state"),
		CodeChallenge:       r.URL.Query().Get("code_challenge"),
		CodeChallengeMethod: r.URL.Query().Get("code_challenge_method"),
		Error:               r.URL.Query().Get("error"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	loginTmpl.Execute(w, data)
}

func (h *OAuthHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid form data")
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	clientID := r.PostFormValue("client_id")
	redirectURI := r.PostFormValue("redirect_uri")
	responseType := r.PostFormValue("response_type")
	scope := r.PostFormValue("scope")
	state := r.PostFormValue("state")
	codeChallenge := r.PostFormValue("code_challenge")
	codeChallengeMethod := r.PostFormValue("code_challenge_method")

	result, err := h.auth.Login(r.Context(), email, password, r.RemoteAddr, r.UserAgent())
	if err != nil {
		loginURL := buildLoginURL(clientID, redirectURI, responseType, scope, state, codeChallenge, codeChallengeMethod, "Invalid email or password")
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "gate_session",
		Value:    result.Session.ID,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   h.httpsEnabled,
		SameSite: http.SameSiteLaxMode,
	})

	authorizeURL := buildAuthorizeURL(clientID, redirectURI, responseType, scope, state, codeChallenge, codeChallengeMethod)
	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

func getSessionCookie(r *http.Request) string {
	c, err := r.Cookie("gate_session")
	if err != nil {
		return ""
	}
	return c.Value
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gate_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	loginURL := "/api/v1/oauth/login?" + q.Encode()
	http.Redirect(w, r, loginURL, http.StatusFound)
}

func buildLoginURL(clientID, redirectURI, responseType, scope, state, codeChallenge, codeChallengeMethod, errMsg string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", responseType)
	q.Set("scope", scope)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", codeChallengeMethod)
	if errMsg != "" {
		q.Set("error", errMsg)
	}
	return "/api/v1/oauth/login?" + q.Encode()
}

func buildAuthorizeURL(clientID, redirectURI, responseType, scope, state, codeChallenge, codeChallengeMethod string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", responseType)
	q.Set("scope", scope)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", codeChallengeMethod)
	return "/api/v1/oauth/authorize?" + q.Encode()
}
