package dgraph

import (
	"strings"
)

func buildQueryINM(id ...string) string {
	q := `{
		Licenses(func:uid($ID)){
			Licenses: count(product.equipment)
		}
	  }`
	return replacer(q, map[string]string{
		"$ID": strings.Join(id, ","),
	})
}
