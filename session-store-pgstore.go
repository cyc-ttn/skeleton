package skeleton

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

// PgSessionStore implements SessionStore using PgStore as a backend.
type PgSessionStore struct {
	CookieName string
	Store      *pgstore.PGStore
}

// GorillaSession wraps gorilla's session so that it implements Session.
type GorillaSession struct {
	*sessions.Session
}

// SetValue sets a value in the session. The value does not be come
// permanent (across requests) util Save is called.
func (s *GorillaSession) SetValue(key string, val interface{}) {
	s.Values[key] = val
}

// GetValue returns a value from the session.
func (s *GorillaSession) GetValue(key string) interface{} {
	return s.Values[key]
}

// PgStore wraps `pgstore` so that it satisfies SessionStore.
func PgStore(db *sqlx.DB, cookieName string, keys ...string) (*PgSessionStore, error) {
	if len(keys) < 1 {
		return nil, errors.New("sessions requires at least one key")
	}

	// Base64 decode the session key.
	byteKeys := make([][]byte, 0, len(keys))
	for _, k := range keys {
		b, err := base64.StdEncoding.DecodeString(k)
		if err != nil {
			return nil, fmt.Errorf("could not decode key for session. %s", err)
		}
		byteKeys = append(byteKeys, b)
	}

	store, err := pgstore.NewPGStoreFromPool(db.DB, byteKeys...)
	if err != nil {
		return nil, err
	}

	return &PgSessionStore{
		Store:      store,
		CookieName: cookieName,
	}, nil
}

// Get should return the session corresponding to a single cookie, predefined
// by the application.
func (s *PgSessionStore) Get(r *http.Request) (Session, error) {
	sess, err := s.Store.Get(r, s.CookieName)
	if err != nil {
		return nil, err
	}
	return &GorillaSession{Session: sess}, nil
}

// Shutdown should run any procedures required on shutdown.
func (s *PgSessionStore) Shutdown() {
	s.Store.StopCleanup(s.Store.Cleanup(time.Minute * 5))
}
