package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func queryBuilderWSD(metric *v1.MetricWSDComputed, scopes []string, id ...string) string {
	var queries, ids []string
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
        var(func: uid($ReferenceTypeIDs,$ReferenceTypeIDs_c)){
                cpu_server as equipment.$ReferenceType.$NumCPU
                cores_server as equipment.$ReferenceType.$NumCores
                server_t as math(max (max(cores_server,8)*cpu_server,16))
        }

        Licenses()@normalize{
                l_server as sum(val(server_t))
                Licenses:math(l_server)
        }

	}
	`
	return replacer(q, map[string]string{
		"$ID":                   strings.Join(id, ","),
		"$BaseTypeIDs":          baseTypeIDs,
		"$ChildDirectEquipment": childDirectEquipment,
		"$ReferenceType":        metric.ReferenceType,
		"$NumCores":             metric.NumCoresAttr,
		"$NumCPU":               metric.NumCPUAttr,
		"$Scopes":               strings.Join(scopes, ","),
	})
}

func queryBuilderWithSAWSD(metric *v1.MetricWSDComputed, scopes []string, id ...string) string {
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
		var(func: uid($ReferenceTypeIDs,$ReferenceTypeIDs_c)) @filter( not eq(len($ReferenceTypeIDs),0)){
				cpu_server as equipment.$ReferenceType.$NumCPU
				cores_server as equipment.$ReferenceType.$NumCores
				server_t as math(max (max(cores_server,8)*cpu_server,16))
		}
		var(func: uid($ReferenceTypeIDs_c)) @filter( eq(len($ReferenceTypeIDs),0)){
				
			~equipment.parent @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes]) AND uid($BaseTypeIDs)){
				vcpu_vm_tmp as equipment.$BaseType.vcpu
				vcpu_vm as math(min(vcpu_vm_tmp,8))
			}
			sum_vcpu_min as sum(val(vcpu_vm))
			cpu_servers as equipment.$ReferenceType.$NumCPU
			cores_servers as equipment.$ReferenceType.$NumCores
			server_t_server as math(max (max(cores_servers,8)*cpu_servers,16))
			server_ts as math(min(sum_vcpu_min,server_t_server))
		
		}
		Licenses()@normalize{
				l_server as sum(val(server_t) )
				l_servers as sum(val(server_ts))
				Licenses:math(l_server+l_servers)
		}
	}`
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

	/*
		var(func: uid($BaseTypeIDs)) @filter( eq(len($ReferenceTypeIDs),0)){
					vcpu_vm_tmp as equipment.$BaseType.vcpu
					equipment.parent @filter(eq(equipment.type,$ReferenceType) AND eq(scopes,[$Scopes])){
						cpu_servers as equipment.$ReferenceType.$NumCPU
						cores_servers as equipment.$ReferenceType.$NumCores
						server_t_server as math(max (max(cores_servers,8)*cpu_servers,16))
						vcpu_vm as math(min(vcpu_vm_tmp,8))
						server_vm as math(min(vcpu_vm,server_t_server))
					}
					server_vm_t as sum(val(server_vm))
					server_ts as math(ceil server_vm_t)

			}
	*/
}
