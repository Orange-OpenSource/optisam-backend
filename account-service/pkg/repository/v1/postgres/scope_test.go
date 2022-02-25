package postgres

import (
	"context"
	"fmt"
	v1 "optisam-backend/account-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountRepository_CreateScope(t *testing.T) {
	type args struct {
		ctx       context.Context
		scopeName string
		scopeCode string
		userID    string
		scopeType string
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		verify  func(a *AccountRepository) error
		wantErr bool
	}{
		{
			name: "Success",
			r:    NewAccountRepository(db),
			args: args{
				ctx:       context.Background(),
				scopeName: "France",
				scopeCode: "O1",
				userID:    "admin@test.com",
				scopeType: "GENERIC",
			},
			verify: func(a *AccountRepository) error {
				// Get scope table and match the Scope
				// Get the root group using ID and match the array.

				scope, err := a.ScopeByCode(context.Background(), "O1")
				if err != nil {
					return err
				}

				expectedScope := v1.Scope{
					ScopeCode: "O1",
					ScopeName: "France",
					CreatedBy: "admin@test.com",
					ScopeType: "GENERIC",
				}

				compareScopesData(t, "CreateScope", &expectedScope, scope)

				group, err := a.GroupInfo(context.Background(), 1)

				if err != nil {
					return err
				}

				isExists := isScopeAdded(group.Scopes, "O1")

				if !isExists {
					return fmt.Errorf("Scope is not there in root group")
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer deleteScopes(context.Background(), []string{"O1"})
			if err := tt.r.CreateScope(tt.args.ctx, tt.args.scopeName, tt.args.scopeCode, tt.args.userID, tt.args.scopeType); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.CreateScope() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}

}

func TestAccountRepository_ListScopes(t *testing.T) {
	scope1 := &v1.Scope{
		ScopeCode: "O1",
		ScopeName: "France",
		CreatedBy: "admin@test.com",
		ScopeType: "GENERIC",
	}
	scope2 := &v1.Scope{
		ScopeCode: "O2",
		ScopeName: "India",
		CreatedBy: "admin@test.com",
		ScopeType: "GENERIC",
	}
	scope3 := &v1.Scope{
		ScopeCode: "O3",
		ScopeName: "England",
		CreatedBy: "admin@test.com",
		ScopeType: "GENERIC",
	}
	scope4 := &v1.Scope{
		ScopeCode: "O4",
		ScopeName: "SriLanka",
		CreatedBy: "admin@test.com",
		ScopeType: "GENERIC",
	}
	type args struct {
		ctx        context.Context
		scopeCodes []string
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		want    []*v1.Scope
		setup   func(a *AccountRepository) (func() error, error)
		wantErr bool
	}{
		{
			name: "SUCCESS",
			r:    NewAccountRepository(db),
			args: args{
				ctx:        context.Background(),
				scopeCodes: []string{"O1", "O2"},
			},
			setup: func(a *AccountRepository) (func() error, error) {
				err := a.CreateScope(context.Background(), scope1.ScopeName, scope1.ScopeCode, scope1.CreatedBy, scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				err = a.CreateScope(context.Background(), scope2.ScopeName, scope2.ScopeCode, scope2.CreatedBy, scope2.ScopeType)
				if err != nil {
					return nil, err
				}

				group, err := a.CreateGroup(context.Background(), "admin@test.com", &v1.Group{
					Name:               "India",
					FullyQualifiedName: "ROOT.India",
					ParentID:           1,
					Scopes:             []string{"O2"},
				})
				if err != nil {
					return nil, err
				}

				return func() error {
					err := deleteScopes(context.Background(), []string{scope1.ScopeCode, scope2.ScopeCode})
					if err != nil {
						return err
					}
					err = a.DeleteGroup(context.Background(), group.ID)
					if err != nil {
						return err
					}

					return nil

				}, nil
			},
			want: []*v1.Scope{
				{
					ScopeCode:  "O1",
					ScopeName:  "France",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType: "GENERIC",
				},
				{
					ScopeCode:  "O2",
					ScopeName:  "India",
					ScopeType: "GENERIC",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT", "India"},
				},
			},
		},
		{
			name: "SUCCESS - Scope is not in scope table",
			r:    NewAccountRepository(db),
			args: args{
				ctx:        context.Background(),
				scopeCodes: []string{"O3"},
			},
			setup: func(a *AccountRepository) (func() error, error) {
				err := a.CreateScope(context.Background(), scope1.ScopeName, scope1.ScopeCode, scope1.CreatedBy, scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				err = a.CreateScope(context.Background(), scope2.ScopeName, scope2.ScopeCode, scope2.CreatedBy ,scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				group, err := a.CreateGroup(context.Background(), "admin@test.com", &v1.Group{
					Name:               "India",
					FullyQualifiedName: "ROOT.India",
					ParentID:           1,
					Scopes:             []string{"O2"},
				})
				if err != nil {
					return nil, err
				}

				return func() error {
					err := deleteScopes(context.Background(), []string{scope1.ScopeCode, scope2.ScopeCode})
					if err != nil {
						return err
					}
					err = a.DeleteGroup(context.Background(), group.ID)
					if err != nil {
						return err
					}

					return nil

				}, nil
			},
			want: []*v1.Scope{},
		},
		{
			name: "SUCCESS - Scope is not there in the group",
			r:    NewAccountRepository(db),
			args: args{
				ctx:        context.Background(),
				scopeCodes: []string{"O1", "O2", "O3", "O4"},
			},
			setup: func(a *AccountRepository) (func() error, error) {
				err := createScopes(context.Background(), scope3, scope4)
				if err != nil {
					return nil, err
				}
				err = a.CreateScope(context.Background(), scope1.ScopeName, scope1.ScopeCode, scope1.CreatedBy,scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				err = a.CreateScope(context.Background(), scope2.ScopeName, scope2.ScopeCode, scope2.CreatedBy,scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				group, err := a.CreateGroup(context.Background(), "admin@test.com", &v1.Group{
					Name:               "India",
					FullyQualifiedName: "ROOT.India",
					ParentID:           1,
					Scopes:             []string{"O2"},
				})
				if err != nil {
					return nil, err
				}

				return func() error {
					err := deleteScopes(context.Background(), []string{scope1.ScopeCode, scope2.ScopeCode, scope3.ScopeCode, scope4.ScopeCode})
					if err != nil {
						return err
					}
					err = a.DeleteGroup(context.Background(), group.ID)
					if err != nil {
						return err
					}

					return nil

				}, nil
			},
			want: []*v1.Scope{
				{
					ScopeCode:  "O1",
					ScopeName:  "France",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType: "GENERIC",
				},
				{
					ScopeCode:  "O2",
					ScopeName:  "India",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT", "India"},
					ScopeType: "GENERIC",
				},
				{
					ScopeCode: "O3",
					ScopeName: "England",
					CreatedBy: "admin@test.com",
					ScopeType: "GENERIC",
				},
				{
					ScopeCode: "O4",
					ScopeName: "SriLanka",
					CreatedBy: "admin@test.com",
					ScopeType: "GENERIC",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.r)
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.ListScopes(tt.args.ctx, tt.args.scopeCodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.ListScopes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareScopesDataAll(t, "ListScopes", tt.want, got)
			}

		})
	}
}

func TestAccountRepository_ScopeByCode(t *testing.T) {
	scope1 := &v1.Scope{
		ScopeCode: "O1",
		ScopeName: "France",
		CreatedBy: "admin@test.com",
		ScopeType: "GENERIC",
	}
	scope2 := &v1.Scope{
		ScopeCode: "O2",
		ScopeName: "India",
		CreatedBy: "admin@test.com",
		ScopeType: "GENERIC",
	}
	type args struct {
		ctx       context.Context
		scopeCode string
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func(a *AccountRepository) (func() error, error)
		want    *v1.Scope
		wantErr bool
	}{
		{
			name: "SUCCESS",
			r:    NewAccountRepository(db),
			args: args{
				ctx:       context.Background(),
				scopeCode: "O1",
			},
			setup: func(a *AccountRepository) (func() error, error) {
				err := a.CreateScope(context.Background(), scope1.ScopeName, scope1.ScopeCode, scope1.CreatedBy,scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				err = a.CreateScope(context.Background(), scope2.ScopeName, scope2.ScopeCode, scope2.CreatedBy,scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				return func() error {
					return deleteScopes(context.Background(), []string{scope1.ScopeCode, scope2.ScopeCode})
				}, nil
			},
			want: scope1,
		},
		{
			name: "SUCCESS - With scope nil",
			r:    NewAccountRepository(db),
			args: args{
				ctx:       context.Background(),
				scopeCode: "O3",
			},
			setup: func(a *AccountRepository) (func() error, error) {
				err := a.CreateScope(context.Background(), scope1.ScopeName, scope1.ScopeCode, scope1.CreatedBy,scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				err = a.CreateScope(context.Background(), scope2.ScopeName, scope2.ScopeCode, scope2.CreatedBy,scope1.ScopeType)
				if err != nil {
					return nil, err
				}

				return func() error {
					return deleteScopes(context.Background(), []string{scope1.ScopeCode, scope2.ScopeCode})
				}, nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.r)
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.ScopeByCode(tt.args.ctx, tt.args.scopeCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.ScopeByCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareScopesData(t, "ListScopes", tt.want, got)
			}
		})
	}
}

func isScopeAdded(scopes []string, scopeCode string) bool {
	for _, scope := range scopes {
		if scope == scopeCode {
			return true
		}
	}

	return false
}

func compareScopesDataAll(t *testing.T, name string, expected []*v1.Scope, actual []*v1.Scope) {
	if expected == nil && actual == nil {
		return
	}

	if expected == nil {
		assert.Nil(t, actual, "Scopes are expected to be nil")
	}

	for i := range expected {
		if idx := scopeIndex(expected[i], actual); idx != -1 {
			compareScopesData(t, fmt.Sprintf("%s[%d]", name, i), expected[i], actual[idx])
		}
	}
}

func scopeIndex(scope *v1.Scope, scopes []*v1.Scope) int {
	for i := range scopes {
		if scope.ScopeCode == scopes[i].ScopeCode {
			return i
		}
	}
	return -1
}

func compareScopesData(t *testing.T, name string, exp, act *v1.Scope) {
	if exp == nil && act == nil {
		return
	}

	if exp == nil {
		assert.Nil(t, act, "Scope is expected to be nil")
	}

	if act == nil {
		assert.Nil(t, exp, "Scope was not expected to be nil")
	}

	assert.Equalf(t, exp.ScopeCode, act.ScopeCode, "%s %s.ScopeCode are not same", name, exp.ScopeCode)
	assert.Equalf(t, exp.ScopeName, act.ScopeName, "%s %s.ScopeName are not same", name, exp.ScopeName)
	assert.Equalf(t, exp.CreatedBy, act.CreatedBy, "%s %s.CreatedBy are not same", name, exp.CreatedBy)
	assert.ElementsMatchf(t, exp.GroupNames, act.GroupNames, "%s %s.GroupNames are not same", name, exp.GroupNames)

}

func deleteScopes(ctx context.Context, deleteScopes []string) error {
	const deleteScope = `DELETE FROM scopes where scope_code = $1 `

	for _, scopeCode := range deleteScopes {
		_, err := db.ExecContext(ctx, deleteScope, scopeCode)
		if err != nil {
			return err
		}
	}

	return nil
}

func createScopes(ctx context.Context, scope1, scope2 *v1.Scope) error {

	const insertScope = `INSERT INTO scopes (scope_code,scope_name,created_by) VALUES ($1,$2,$3), ($4,$5,$6)`

	_, err := db.ExecContext(ctx, insertScope, scope1.ScopeCode, scope1.ScopeName, scope1.CreatedBy, scope2.ScopeCode, scope2.ScopeName, scope2.CreatedBy)

	if err != nil {
		return err
	}

	return nil
}
