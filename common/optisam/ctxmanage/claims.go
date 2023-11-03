package ctxmanage

import (
	"context"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
)

type key uint8

const (
	keyClaims key = 0
)

// AddClaims add claims to context
func AddClaims(ctx context.Context, clms *claims.Claims) context.Context {
	return context.WithValue(ctx, keyClaims, clms)
}

// RetrieveClaims retuive claims from context
func RetrieveClaims(ctx context.Context) (*claims.Claims, bool) {
	clms, ok := ctx.Value(keyClaims).(*claims.Claims)
	return clms, ok
}
