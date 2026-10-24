package auth

import "errors"

var ErrIdentityConflict = errors.New("identity already in use")

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Registration struct {
	Username string
	Email    string
	Password string
}

type Credentials struct {
	Identity string
	Password string
}

type Session struct {
	Token     string
	ExpiresAt string
}

type Actor struct {
	UserID   int64
	Username string
	Email    string
}
