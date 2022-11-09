package management

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

// ErrIncorrectPassword is returned on Login when the given password is
// incorrect.
var ErrIncorrectPassword = errors.New("incorrect password")

// Auth contains utilities that is used for authenticating the administrator to
// manage the system.
type Auth struct {
	// password is an administrator password.
	password string

	// sessionCache is a cache that stores valid access tokens after
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

// Login creates and returns a new access token if the provided password matches
// the password in the configuration. The newly generated access token is stored
// in the sessionCache so that it can be used for further requests.
func (a *Auth) Login(password string) (string, error) {
	if password != a.password {
		return "", ErrIncorrectPassword
	}

	accessToken, err := generateAccessToken()
	if err != nil {
		return "", err
	}
	if err := a.sessionCache.Add(accessToken, nil,
		cache.DefaultExpiration); err != nil {

		return "", fmt.Errorf("add session cache: %w", err)
	}

	return accessToken, nil
}

// AccessTokenValid returns true if the provided access token exists in the
// sessionCache, meaning that the access token can still be used for further
// requests.
func (a *Auth) AccessTokenValid(accessToken string) bool {
	_, ok := a.sessionCache.Get(accessToken)
	return ok
}

func generateAccessToken() (string, error) {
	accessTokenBytes := make([]byte, 16)
	if _, err := rand.Read(accessTokenBytes); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}
	return hex.EncodeToString(accessTokenBytes), nil
}
