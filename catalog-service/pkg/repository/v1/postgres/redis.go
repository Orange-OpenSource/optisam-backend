package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

var (
	keyScope = "scope_details_"
)

type Scope struct {
	ScopeCode  string
	ScopeName  string
	CreatedBy  string
	ScopeType  string
	CreatedOn  time.Time
	GroupNames []string
	Expenses   sql.NullFloat64
}

func (p *ProductCatalogRepository) GetScope(ctx context.Context, scopes []string) (scope []*Scope, err error) {
	var s []string
	for _, v := range scopes {
		s = append(s, keyScope+v)
	}
	sc, err := p.r.MGet(ctx, s...).Result()
	if err != nil {
		return scope, err
	}
	for _, v := range sc {
		s, _ := v.(string)
		scp := *&Scope{}
		err = json.Unmarshal([]byte(s), &scp)
		if err != nil {
			return scope, err
		}
		scope = append(scope, &scp)
	}
	return scope, err
}

func (p *ProductCatalogRepository) GetAllScope(ctx context.Context) (scope []*Scope, err error) {
	scopesKeys, err := p.r.Keys(ctx, keyScope+"*").Result()
	if err != nil {
		return nil, err
	}
	sc, err := p.r.MGet(ctx, scopesKeys...).Result()
	if err != nil {
		return scope, err
	}
	for _, v := range sc {
		s, _ := v.(string)
		scp := *&Scope{}
		err = json.Unmarshal([]byte(s), &scp)
		if err != nil {
			return scope, err
		}
		scope = append(scope, &scp)
	}
	return scope, err
}
