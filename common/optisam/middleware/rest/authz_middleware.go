package rest

import (
	"net/http"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/opa"

	"github.com/open-policy-agent/opa/rego"
)

// ValidateAuthZ for RBAC authorization
func ValidateAuthZ(p *rego.PreparedEvalQuery, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userClaims, ok := RetrieveClaims(r.Context())
		if !ok {
			logger.Log.Error("invalid claims")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Authorize
		authorized, err := opa.EvalAuthZ(r.Context(), p, opa.AuthzInput{Role: string(userClaims.Role), MethodFullName: r.RequestURI})
		if err != nil || !authorized {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r) // proceed in the middleware chain!
	})
}
