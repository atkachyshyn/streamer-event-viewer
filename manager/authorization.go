package manager

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/atkachyshyn/streamer-event-viewer/shared"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

const (
	stateCallbackKey = "oauth-state-callback"
	oauthSessionName = "oauth-session"
	oauthTokenKey    = "oauth-token"
)

var (
	clientID = "ks1eshjn7pp016vrl566aq415k3r65"
	// Consider storing the secret in an environment variable or a dedicated storage system.
	clientSecret = "vfo50a5y3275tgupxmfsbdfqiqmib1"
	scopes       = []string{"user:read:email"}
	redirectURL  = "http://localhost:8080/redirect"
	oauth2Config *oauth2.Config
)

// AuthorizationHandler handlers
type AuthorizationHandler struct {
	logger                log.Logger
	store                 *sessions.CookieStore
	LoginHandler          http.Handler
	OAuth2CallbackHandler http.Handler
}

// NewAuthorizationHandlers returns authorization handlers
func NewAuthorizationHandlers(l log.Logger) AuthorizationHandler {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})

	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint:     twitch.Endpoint,
		RedirectURL:  redirectURL,
	}

	l.Log("method", "NewHandlers", "action", "return login and OAuth2 callback handlers")

	return AuthorizationHandler{
		logger:                l,
		LoginHandler:          errorHandling(middleware(HandleLogin), l),
		OAuth2CallbackHandler: errorHandling(middleware(HandleOAuth2Callback), l),
	}
}

// HandleLogin is a Handler that redirects the user to Twitch for login, and provides the 'state'
// parameter which protects against login CSRF.
func HandleLogin(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := shared.CookieStore.Get(r, oauthSessionName)
	if err != nil {
		// Return new session
		fmt.Errorf("method=HandleLogin Session doesn't exist, return new: %s", err)
		err = nil
	}

	session.Options.MaxAge = 0

	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return AnnotateError(err, "Couldn't generate a session!", http.StatusInternalServerError)
	}

	state := hex.EncodeToString(tokenBytes[:])

	session.AddFlash(state, stateCallbackKey)

	if err = session.Save(r, w); err != nil {
		fmt.Println("Session save err...")
		return
	}

	fmt.Println("HandleLogin redirecting...")

	http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)

	return
}

// HandleOAuth2Callback is a Handler for oauth's 'redirect_uri' endpoint;
func HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) (err error) {
	fmt.Println("Starting HandleOAuth2Callback...")

	session, err := shared.CookieStore.Get(r, oauthSessionName)
	if err != nil {
		// Return new session
		fmt.Errorf("method=HandleOAuth2Callback Session doesn't exist, return new: %s", err)
		err = nil
	}

	session.Options.MaxAge = 0

	// ensure we flush the csrf challenge even if the request is ultimately unsuccessful
	defer func() {
		if err := session.Save(r, w); err != nil {
			fmt.Errorf("error saving session: %s", err)
		}

		fmt.Println("HandleOAuth2Callback Session save...")
	}()

	switch stateChallenge, state := session.Flashes(stateCallbackKey), r.FormValue("state"); {
	case state == "", len(stateChallenge) < 1:
		err = errors.New("missing state challenge")
	case state != stateChallenge[0]:
		err = fmt.Errorf("invalid oauth state, expected '%s', got '%s'\n", state, stateChallenge[0])
	}

	if err != nil {
		return AnnotateError(
			err,
			"Couldn't verify your confirmation, please try again.",
			http.StatusBadRequest,
		)
	}

	token, err := oauth2Config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		return
	}

	// add the oauth token to session
	session.Values[oauthTokenKey] = token

	// fmt.Printf("Access token: %s\n", token.AccessToken)

	fmt.Println("HandleOAuth2Callback redirecting...", session.Values[oauthTokenKey].(*oauth2.Token).AccessToken)

	if err := session.Save(r, w); err != nil {
		fmt.Errorf("error saving session: %s", err)
	}

	http.Redirect(w, r, "http://localhost:3000/", http.StatusTemporaryRedirect)

	return
}

// HumanReadableError represents error information
// that can be fed back to a human user.
//
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type HumanReadableError interface {
	HumanError() string
	HTTPCode() int
}

// HumanReadableWrapper implements HumanReadableError
type HumanReadableWrapper struct {
	ToHuman string
	Code    int
	error
}

// HumanError returns human readable error
func (h HumanReadableWrapper) HumanError() string { return h.ToHuman }

// HTTPCode returns HTTP code of the error
func (h HumanReadableWrapper) HTTPCode() int { return h.Code }

// AnnotateError wraps an error with a message that is intended for a human end-user to read,
// plus an associated HTTP error code.
func AnnotateError(err error, annotation string, code int) error {
	if err == nil {
		return nil
	}
	return HumanReadableWrapper{ToHuman: annotation, error: err}
}

// Handler handles request
type Handler func(http.ResponseWriter, *http.Request) error

var middleware = func(h Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// parse POST body, limit request size
		if err = r.ParseForm(); err != nil {
			return AnnotateError(err, "Something went wrong! Please try again.", http.StatusBadRequest)
		}

		return h(w, r)
	}
}

// errorHandling is a middleware that centralises error handling.
// this prevents a lot of duplication and prevents issues where a missing
// return causes an error to be printed, but functionality to otherwise continue
// see https://blog.golang.org/error-handling-and-go
var errorHandling = func(handler func(w http.ResponseWriter, r *http.Request) error, l log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var errorString string = "Something went wrong! Please try again."
			var errorCode int = 500

			if v, ok := err.(HumanReadableError); ok {
				errorString, errorCode = v.HumanError(), v.HTTPCode()
			}

			l.Log("method", "errorHandling", "error", err.Error())
			w.Write([]byte(errorString))
			w.WriteHeader(errorCode)
			return
		}
	})
}
