package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func queryBuilderMSS(metric *v1.MetricMSSComputed, scopes []string, id ...string) string {
	var queries, ids []string
	var baseType string
	qDirectEquipments := `
		product.equipment @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes])){
			$CurrentTypeIDs as uid
	   }`
	for _, val := range metric.BaseType {
		vars := map[string]string{
			"$CurrentType": val,
			"$Scopes":      strings.Join(scopes, ","),
		}
		ids = append(ids, val+"IDs")
		baseType = val
		queries = append(queries, replacer(qDirectEquipments, vars))
	}
	childDirectEquipment := strings.Join(queries, "\n\t")
	baseTypeIDs := strings.Join(ids, ",")
	q := `{
		var(func: uid($ID)){
			$ChildDirectEquipment
		}
		var (func: uid($BaseTypeIDs)){
			equipment.parent @filter(eq(equipment.type,$ReferenceType) AND eq(scopes,[$Scopes])){
				$ReferenceTypeIDs_c as uid
			}
		}
		var(func: uid($ID)){
			product.equipment @filter(eq(equipment.type,$ReferenceType) AND eq(scopes,[$Scopes])){
				$ReferenceTypeIDs as uid
			}
		}
        var (func: uid($ReferenceTypeIDs_c)){
			~equipment.parent @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes]) AND uid($BaseTypeIDs)){
				vcpu_vm_tmp as equipment.virtualmachine.vcpu
				server_ts as math(max(vcpu_vm_tmp,4))
			}
        }
        
        var(func: uid($ReferenceTypeIDs)){
                cpu_server as equipment.$ReferenceType.$NumCPU
                cores_server as equipment.$ReferenceType.$NumCores
                server_t as math(cpu_server*(max(cores_server,4)))
        }

        Licenses()@normalize{
                l_server as sum(val(server_t))
				l_servers as sum(val(server_ts))
                Licenses:math(l_server+l_servers)
        }

	}
	`
	return replacer(q, map[string]string{
		"$ID":                   strings.Join(id, ","),
		"$BaseType":             baseType,
		"$BaseTypeIDs":          baseTypeIDs,
		"$ChildDirectEquipment": childDirectEquipment,
		"$ReferenceType":        metric.ReferenceType,
		"$NumCores":             metric.NumCoresAttr,
		"$NumCPU":               metric.NumCPUAttr,
		"$Scopes":               strings.Join(scopes, ","),
	})
}
