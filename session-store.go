package skeleton

import (
	"net/http"
)

// SessionStore describes functionality required for session management by the
// Server. It assumes that the underlying session architecture relies on
// github.com/gorilla/sessions.
type SessionStore interface {
	// Get should return the session corresponding to a single cookie, predefined
	// by the application. A session object should always be returned regardless
	// of the presence (or absense) or a session defining object (such as a
	// cookie).
	Get(r *http.Request) (Session, error)

	// Shutdown should run any procedures required on shutdown.
	Shutdown()
}

// Session describes a singular session object. For example, this could store
// the UserID related to a cookie.
type Session interface {
	// Save completes the session operation.
	Save(r *http.Request, w http.ResponseWriter) error

	// SetValue sets a value in the session. The value does not be come
	// permanent (across requests) util Save is called.
	SetValue(key string, val interface{})

	// GetValue returns a value from the session.
	GetValue(key string) interface{}
}
