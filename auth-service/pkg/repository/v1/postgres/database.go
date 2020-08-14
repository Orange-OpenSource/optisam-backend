// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import "database/sql"

// Default implemets ../v1.Repository interface
type Default struct {
	db *sql.DB
}

// NewRepository returns an implementation of Repository interface.
func NewRepository(db *sql.DB) *Default {
	return &Default{
		db: db,
	}
}
