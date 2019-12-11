// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_ListEditors(t *testing.T) {
	type args struct {
		ctx    context.Context
		params *v1.EditorQueryParams
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    []*v1.Editor
		wantErr bool
	}{
		{name: "success",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				params: nil,
				scopes: []string{"A", "B"},
			},
			want: []*v1.Editor{
				&v1.Editor{
					Name: "oracle",
				},
				&v1.Editor{
					Name: "Windows",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.ListEditors(tt.args.ctx, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ListEditors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareEditorsAll(t, "editors", tt.want, got)
			}
		})
	}
}

func compareEditorsAll(t *testing.T, name string, exp []*v1.Editor, act []*v1.Editor) {
	for i := range exp {
		idx := getEditorByName(exp[i].Name, act)
		if !assert.NotEqualf(t, -1, idx, "editor by name: %s not found in actual editors", exp[i].Name) {
			return
		}
		compareEditor(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
	}
}
func getEditorByName(name string, editors []*v1.Editor) int {
	for i := range editors {
		if name == editors[i].Name {
			return i
		}
	}
	return -1
}
func compareEditor(t *testing.T, name string, exp *v1.Editor, act *v1.Editor) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "editor is expected to be nil")
	}

	if exp.ID != "" {
		assert.Emptyf(t, act.ID, "%s.ID is expected to be nil", name)
	}
	assert.Equalf(t, exp.Name, act.Name, "%s.Name should be same", name)
}
