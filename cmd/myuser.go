package cmd

import (
	"crypto"

	"github.com/go-acme/lego/v4/registration"
)

// MyUser represents an ACME user for Let's Encrypt certificate generation.
// Taken directly from the lego example, since we need very minimal support
// https://go-acme.github.io/lego/usage/library/
// Moved from: utils.go
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

// NewMyUser creates a new ACME user with the given email and private key.
// Moved from: utils.go
func NewMyUser(email string, key crypto.PrivateKey) *MyUser {
	return &MyUser{
		Email: email,
		key:   key,
	}
}

// GetEmail returns the user's email address for ACME registration.
// Moved from: utils.go
func (u *MyUser) GetEmail() string {
	return u.Email
}

// GetRegistration returns the user's ACME registration resource.
// Moved from: utils.go
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the user's private key for ACME operations.
// Moved from: utils.go
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
