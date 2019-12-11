// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// ListEditors implements Licence ListEditors function
func (r *LicenseRepository) ListEditors(ctx context.Context, params *v1.EditorQueryParams, scopes []string) ([]*v1.Editor, error) {
	q := `{
		Editors(func:eq(type,"editor")){
			ID  : uid,
			Name: editor.name,
		}
	}`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/Editor- ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/Editor - cannot complete query")
	}

	type data struct {
		Editors []*v1.Editor
	}

	editors := data{}

	if err := json.Unmarshal(resp.GetJson(), &editors); err != nil {
		logger.Log.Error("dgraph/Editor - ", zap.String("reason", err.Error()))
		return nil, fmt.Errorf("dgraph/Editor - cannot unmarshal Json object")
	}
	return editors.Editors, nil
}
