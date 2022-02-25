package client

import (
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

//go:generate mockgen -destination=mock/mock.go -package=mock gopkg.in/oauth2.v3 ClientStore
type store struct {
}

// NewStore returns oauth2.ClientStore
func NewStore() oauth2.ClientStore {
	return &store{}
}

// GetByID implements oauth2.ClientStore GetByID function
func (s *store) GetByID(id string) (oauth2.ClientInfo, error) {
	return &models.Client{}, nil
}
