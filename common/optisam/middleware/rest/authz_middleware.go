// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"net/http"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/opa"

	"github.com/open-policy-agent/opa/rego"
)

func ValidateAuthZ(p *rego.PreparedEvalQuery, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userClaims, ok := ctxmanage.RetrieveClaims(r.Context())
		if !ok {
			logger.Log.Error("invalid claims")
			return
		}
		// Authorize
		authorized, err := opa.EvalAuthZ(r.Context(), p, opa.AuthzInput{Role: string(userClaims.Role), MethodFullName: r.RequestURI})
		if err != nil || !authorized {
			logger.Log.Sugar().Errorf("User Authorized to access %s with Role %s", r.RequestURI, string(userClaims.Role))
			return
		}
		logger.Log.Sugar().Infof("User Authorized to access %s with Role %s", r.RequestURI, string(userClaims.Role))
		h.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
