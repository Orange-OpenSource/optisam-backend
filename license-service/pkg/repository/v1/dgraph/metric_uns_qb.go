package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
)

func buildQueryUNS(metric *v1.MetricUNSComputed, scopes []string, id ...string) string {
	profileFilter := `eq(nominative.user.profile,"` + metric.Profile + `" ) AND`
	if strings.ToLower(metric.Profile) == "all" || metric.Profile == "" {
		profileFilter = ``
	}
	q := `{
		var(func:uid($ID)){
			product.nominative.users @filter(` + profileFilter + ` eq(scopes,[$Scopes])){
				un as count(uid)
	 		}
		}
		Licenses(){
		  Licenses: sum(val(un))
		}
	  }`
	return replacer(q, map[string]string{
		"$ID":     strings.Join(id, ","),
		"$Scopes": strings.Join(scopes, ","),
	})
}

func buildQueryUNSAgg(metric *v1.MetricUNSComputed, scopes []string, id string) string {
	profileFilter := `eq(nominative.user.profile,"` + metric.Profile + `" ) AND`
	if strings.ToLower(metric.Profile) == "all" || metric.Profile == "" {
		profileFilter = ``
	}
	q := `{
		var(func:uid($ID)){
			aggregation.nominative.users @filter(` + profileFilter + ` eq(scopes,[$Scopes])){
				un as count(uid)
	 		}
		}
		Licenses(){
		  Licenses: sum(val(un))
		}
	  }`
	return replacer(q, map[string]string{
		"$ID":     id,
		"$Scopes": strings.Join(scopes, ","),
	})
}
