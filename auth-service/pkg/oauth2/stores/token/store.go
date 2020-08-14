// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

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

// Create implements gopkg.in/oauth2 create fucntion.
func (s *store) Create(info oauth2.TokenInfo) error {
	// We returning nil as the framework that we are using expects us to
	// store token in database or some other storage type.
	return nil
}

// RemoveByCode implements gopkg.in/oauth2 RemoveByCode fucntion.
func (s *store) RemoveByCode(code string) error {
	// We returning nil as the framework that we are using expects us to
	// remove token from database or some other storage type.
	return nil
}

// RemoveByAccess implements gopkg.in/oauth2 RemoveByAccess fucntion
func (s *store) RemoveByAccess(access string) error {
	// We returning nil as the framework that we are using expects us to
	// remove token from database or some other storage type.
	return nil
}

// RemoveByRefresh implements gopkg.in/oauth2 RemoveByRefresh fucntion
func (s *store) RemoveByRefresh(refresh string) error {
	// We returning nil as the framework that we are using expects us to
	// remove token from database or some other storage type.
	return nil
}

// GetByCode implements gopkg.in/oauth2 GetByCode fucntion
func (s *store) GetByCode(code string) (oauth2.TokenInfo, error) {
	return nil, errors.New("not supported")
}

// GetByAccess implements gopkg.in/oauth2 GetByAccess fucntion
func (s *store) GetByAccess(access string) (oauth2.TokenInfo, error) {
	return nil, errors.New("not supported")
}

// GetByRefresh implements gopkg.in/oauth2 GetByRefresh fucntion
func (s *store) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	return nil, errors.New("not supported")
}
