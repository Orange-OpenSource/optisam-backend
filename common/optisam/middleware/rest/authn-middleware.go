// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"optisam-backend/common/optisam/token/claims"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

type key uint8

const (
	keyClaims key = 0
)

// AddClaims add claims to context
func AddClaims(ctx context.Context, clms *claims.Claims) context.Context {
	ctx.Value(LoggerKey{}).(*LoggerUserDetails).UserID = clms.UserID
	ctx.Value(LoggerKey{}).(*LoggerUserDetails).Role = string(clms.Role)
	return context.WithValue(ctx, keyClaims, clms)
}

// RetrieveClaims retuive claims from context
func RetrieveClaims(ctx context.Context) (*claims.Claims, bool) {
	clms, ok := ctx.Value(keyClaims).(*claims.Claims)
	return clms, ok
}

// ValidateAuth is a middleware to check for JWT authorization
// TODO
func ValidateAuth(verifyKey *rsa.PublicKey, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader != "" {
			bearerToken := strings.TrimPrefix(authorizationHeader, "Bearer")
			bearerToken = strings.TrimSpace(bearerToken)
			//tokenPart := bearerToken[1] //Grab the token part, what we are truly interested in
			customClaims := &claims.Claims{}

			token, err := jwt.ParseWithClaims(bearerToken, customClaims, func(token *jwt.Token) (interface{}, error) {
				return verifyKey, nil
			})

			if err != nil { //Malformed token, returns with http code 403 as usual
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if !token.Valid { //Token is invalid, maybe not signed on this server
				w.WriteHeader(http.StatusForbidden)

				return
			}
			ctx := r.Context()
			//Everything went well, proceed with the request and set the caller to the user retrieved from the parsed token
			r = r.WithContext(AddClaims(ctx, customClaims))
			h.ServeHTTP(w, r) //proceed in the middleware chain!

			//fmt.Println(len(bearerToken))
		} else {
			json.NewEncoder(w).Encode("Invalid Authorization Token")
		}
	})
}
