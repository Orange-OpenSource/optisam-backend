// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package generator

import (
	"optisam-backend/common/optisam/token/claims"
	"os"
	"testing"
)

func Test_tokenGenerator_GenerateAccessToken(t *testing.T) {
	type args struct {
		osClaims *claims.Claims
	}
	tests := []struct {
		name    string
		t       *tokenGenerator
		args    args
		want    string
		wantErr bool
	}{
		{name: "Token",
			args: args{
				osClaims: &claims.Claims{
					UserID: "admin@test.com",
					Locale: "France",
					Role:   claims.RoleAdmin,
					Socpes: []string{"France"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg, err := NewTokenGenerator("../../../../auth-service/cmd/server/key.pem")
			if err != nil {
				t.Errorf("NewTokenGenerator - could not create tokengenerator %v", err)
				return
			}
			got, err := tg.GenerateAccessToken(tt.args.osClaims)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenGenerator.GenerateAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if got != tt.want {
			// 	t.Errorf("tokenGenerator.GenerateAccessToken() = %v, want %v", got, tt.want)
			// }
			f, err := os.Create("token.txt")
			if err != nil {
				t.Errorf("Create - could not create a file %v", err)
				return
			}
			_, err = f.WriteString(got)
			if err != nil {
				t.Errorf("WriteString - could not write in file %v", err)
				f.Close()
				return
			}
			f.Close()
		})
	}
}
