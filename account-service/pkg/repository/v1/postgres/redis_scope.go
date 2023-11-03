package postgres

import (
	"context"
	"encoding/json"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/repository/v1"
)

// func (r *AccountRepository) FetchScopes(ctx context.Context,scopeCodes []string) ([]*v1.Scope, error) {
// 	r.r.Get(ctx, "")
// }

var (
	keyScope = "scope_details_"
)

func (r *AccountRepository) SetScope(ctx context.Context, scope []*v1.Scope) (err error) {
	var socpesList []interface{}
	for _, v := range scope {
		s, _ := json.Marshal(v)
		socpesList = append(socpesList, keyScope+v.ScopeCode, s)
	}
	status := r.r.MSet(ctx, socpesList...)
	if status.Err() != nil {
		return status.Err()
	}
	return nil
}

func (r *AccountRepository) GetScopes(ctx context.Context, s []string) (scope []*v1.Scope, err error) {
	var scp []string
	for _, v := range s {
		scp = append(scp, keyScope+v)
	}
	sc, err := r.r.MGet(ctx, scp...).Result()
	if err != nil {
		return scope, err
	}
	for _, v := range sc {
		s, _ := v.(string)
		scp := *&v1.Scope{}
		err = json.Unmarshal([]byte(s), &scp)
		if err != nil {
			return scope, err
		}
		scope = append(scope, &scp)
	}
	return scope, err
}

func (r *AccountRepository) DropScope(ctx context.Context, s string) (err error) {
	_, err = r.r.Del(ctx, keyScope+s).Result()
	if err != nil {
		return err
	}
	return err
}
