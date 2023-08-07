package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
)

func buildQueryUCS(metric *v1.MetricUCSComputed, scopes []string, id ...string) string {
	profileFilter := `eq(concurrent.user.profile_user,"` + metric.Profile + `" ) AND`
	if strings.ToLower(metric.Profile) == "all" || metric.Profile == "" {
		profileFilter = ``
	}
	q := `{
		var(func:uid($ID)){
			product.concurrent.users @filter(` + profileFilter + `  eq(scopes,[$Scopes])){
				uc as concurrent.user.number_of_users
	 		}
		}
		Licenses(){
		  Licenses: sum(val(uc))
				}
	  }`
	return replacer(q, map[string]string{
		"$ID":     strings.Join(id, ","),
		"$Scopes": strings.Join(scopes, ","),
	})
}

func buildQueryUCSAgg(metric *v1.MetricUCSComputed, scopes []string, id string) string {
	profileFilter := `eq(concurrent.user.profile_user,"` + metric.Profile + `" ) AND`
	if strings.ToLower(metric.Profile) == "all" || metric.Profile == "" {
		profileFilter = ``
	}
	q := `{
		var(func:uid($ID)){
			aggregation.concurrent.users @filter( ` + profileFilter + ` eq(scopes,[$Scopes])){
				uc as concurrent.user.number_of_users
	 		}
		}
		Licenses(){
		  Licenses: sum(val(uc))
		}
	  }`
	return replacer(q, map[string]string{
		"$ID":     id,
		"$Scopes": strings.Join(scopes, ","),
	})
}
