package token

import (
	"errors"

	"gopkg.in/oauth2.v3"
)

//go:generate mockgen -destination=mock/mock.go -package=mock gopkg.in/oauth2.v3 TokenStore
type store struct{}

// NewStore returns a custom implementation of oauth2.TokenStore
func NewStore() oauth2.TokenStore {
	return &store{}
}

// Create implements gopkg.in/oauth2 create function.
func (s *store) Create(info oauth2.TokenInfo) error {
	// We returning nil as the framework that we are using expects us to
	// store token in database or some other storage type.
	return nil
}

// RemoveByCode implements gopkg.in/oauth2 RemoveByCode function.
func (s *store) RemoveByCode(code string) error {
	// We returning nil as the framework that we are using expects us to
	// remove token from database or some other storage type.
	return nil
}

// RemoveByAccess implements gopkg.in/oauth2 RemoveByAccess function
func (s *store) RemoveByAccess(access string) error {
	// We returning nil as the framework that we are using expects us to
	// remove token from database or some other storage type.
	return nil
}

// RemoveByRefresh implements gopkg.in/oauth2 RemoveByRefresh function
func (s *store) RemoveByRefresh(refresh string) error {
	// We returning nil as the framework that we are using expects us to
	// remove token from database or some other storage type.
	return nil
}

// GetByCode implements gopkg.in/oauth2 GetByCode function
func (s *store) GetByCode(code string) (oauth2.TokenInfo, error) {
	return nil, errors.New("not supported")
}

// GetByAccess implements gopkg.in/oauth2 GetByAccess function
func (s *store) GetByAccess(access string) (oauth2.TokenInfo, error) {
	return nil, errors.New("not supported")
}

// GetByRefresh implements gopkg.in/oauth2 GetByRefresh function
func (s *store) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	return nil, errors.New("not supported")
}
