package management

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/patrickmn/go-cache"
)

// Auth contains utilities that is used for authenticating the administrator to
// manage the system.
type Auth struct {
	// password is an administrator password.
	password string

	// sessionCache is a cache that stores valid session IDs after
	// authentication succeeds.
	sessionCache *cache.Cache
}

// NewAuth is a constructor for the Auth struct. It uses config parameter for
// setting up the properties of the Auth struct.
func NewAuth(config Config) *Auth {
	return &Auth{
		password: config.Password,
		sessionCache: cache.New(
			time.Second*time.Duration(config.SessionTimeout),
			time.Minute*10,
		),
	}
}

// Login creates and returns a new session ID if the provided password matches
// the password in the configuration. The newly generated session ID is stored
// in the sessionCache so that it can be used for further requests.
func (a *Auth) Login(password string) (string, bool) {
	if password != a.password {
		return "", false
	}

	sessionID := generateSessionID()
	a.sessionCache.Add(sessionID, nil, cache.DefaultExpiration)

	return sessionID, true
}

// SessionValid returns true if the provided session ID exists in the
// sessionCache, meaning that the session ID can still be used for further
// requests.
func (a *Auth) SessionValid(sessionID string) bool {
	_, ok := a.sessionCache.Get(sessionID)
	return ok
}

func generateSessionID() string {
	sessionIDBytes := make([]byte, 16)
	rand.Read(sessionIDBytes)

	return hex.EncodeToString(sessionIDBytes)
}
