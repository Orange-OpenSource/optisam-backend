package dgraph

import (
	"strings"
)

func buildQueryINM(id ...string) string {
	q := `{
		var(func:uid($ID)){
			instanceCount as count(product.equipment)
		}
		Licenses(){
		  Licenses: sum(val(instanceCount))
		}
	  }`
	return replacer(q, map[string]string{
		"$ID": strings.Join(id, ","),
	})
}
