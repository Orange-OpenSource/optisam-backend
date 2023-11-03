package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func queryBuilderWSS(metric *v1.MetricWSSComputed, scopes []string, id ...string) string {
	var queries, queriesIndirect, ids []string
	qDirectEquipments := `
		product.equipment @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes])){
			$CurrentTypeIDs as uid
	   }`
	qInDirectEquipments := `
	~equipment.parent @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes]) AND uid($CurrentTypeIDs)){
		vcpucount as count(equipment.id)
	
	}`
	for _, val := range metric.BaseType {
		vars := map[string]string{
			"$CurrentType": val,
			"$Scopes":      strings.Join(scopes, ","),
		}
		ids = append(ids, val+"IDs")
		queries = append(queries, replacer(qDirectEquipments, vars))
		queriesIndirect = append(queriesIndirect, replacer(qInDirectEquipments, vars))
	}
	childDirectEquipment := strings.Join(queries, "\n\t")
	childInDirectEquipment := strings.Join(queriesIndirect, "\n\t")
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
        var(func: uid($ReferenceTypeIDs)){
                cpu_server as equipment.$ReferenceType.$NumCPU
                cores_server as equipment.$ReferenceType.$NumCores
                server_t as math(max (max(cores_server,8)*cpu_server,16))
        }
		var(func: uid($ReferenceTypeIDs_c)){
			cpu_server_vm as equipment.$ReferenceType.$NumCPU
			cores_server_vm as equipment.$ReferenceType.$NumCores
			server_rqlc as math(max(max(cores_server_vm,8)*cpu_server_vm,16))
			
			$ChildInDirectEquipment

			totalVCPUCount as sum(val(vcpucount))
			vcpuMaxValue as math(max(totalVCPUCount,2))
			totalReqLic as math(vcpuMaxValue/2)
			server_vm as math(totalReqLic*server_rqlc)
        }
        Licenses()@normalize{
                l_server as sum(val(server_t))
				l_vm as sum(val(server_vm))
                Licenses:math(l_server+l_vm)
        }

	}
	`
	return replacer(q, map[string]string{
		"$ID":                     strings.Join(id, ","),
		"$BaseTypeIDs":            baseTypeIDs,
		"$ChildDirectEquipment":   childDirectEquipment,
		"$ChildInDirectEquipment": childInDirectEquipment,
		"$ReferenceType":          metric.ReferenceType,
		"$NumCores":               metric.NumCoresAttr,
		"$NumCPU":                 metric.NumCPUAttr,
		"$Scopes":                 strings.Join(scopes, ","),
	})
}

func queryBuilderWithSAWSS(metric *v1.MetricWSSComputed, scopes []string, id ...string) string {
	var queries, queriesIndirect, queriesVM, ids []string
	qDirectEquipments := `
		product.equipment @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes])){
			$CurrentTypeIDs as uid
	   }`
	qInDirectEquipments := `
	~equipment.parent @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes]) AND uid($CurrentTypeIDs)){
		vcpucount as count(equipment.id)
	
	}`

	qInDirectEquipmentVM := `
	~equipment.parent @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes]) AND uid($CurrentTypeIDs)){
		vcpuVM as equipment.$CurrentType.vcpu
	}`
	for _, val := range metric.BaseType {
		vars := map[string]string{
			"$CurrentType": val,
			"$Scopes":      strings.Join(scopes, ","),
		}
		ids = append(ids, val+"IDs")
		queries = append(queries, replacer(qDirectEquipments, vars))
		queriesIndirect = append(queriesIndirect, replacer(qInDirectEquipments, vars))
		queriesVM = append(queriesVM, replacer(qInDirectEquipmentVM, vars))
	}
	childDirectEquipment := strings.Join(queries, "\n\t")
	childInDirectEquipment := strings.Join(queriesIndirect, "\n\t")
	childInDirectSAEquipmentVM := strings.Join(queriesVM, "\n\t")
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
        var(func: uid($ReferenceTypeIDs)){
                cpu_server as equipment.$ReferenceType.$NumCPU
                cores_server as equipment.$ReferenceType.$NumCores
                server_t as math(max (max(cores_server,8)*cpu_server,16))
        }
		var(func: uid($ReferenceTypeIDs_c)) @filter(not eq(len($ReferenceTypeIDs),0 )){
			cpu_server_vm as equipment.$ReferenceType.$NumCPU
			cores_server_vm as equipment.$ReferenceType.$NumCores
			server_rqlc as math(max(max(cores_server_vm,8)*cpu_server_vm,16))
			
			$ChildInDirectEquipment

        	totalVCPUCount as sum(val(vcpucount))
			vcpuMaxValue as math(max(totalVCPUCount,2))
			totalReqLic as math(vcpuMaxValue/2)
			server_vm as math(totalReqLic*server_rqlc)
        }
		var(func: uid($ReferenceTypeIDs_c)) @filter(eq(len($ReferenceTypeIDs),0)){
			cpu_servervm as equipment.$ReferenceType.$NumCPU
			cores_servervm as equipment.$ReferenceType.$NumCores
			server_rqlcvm as math(max(max(cores_servervm,8)*cpu_servervm,16))
			
			$ChildInDirectSAEquipmentVM
			
        	totalVCPUVm as sum(val(vcpuVM))
			servervm as math(min(totalVCPUVm,server_rqlcvm))
        }
        Licenses()@normalize{
			l_server as sum(val(server_t))
			l_vm as sum(val(server_vm))
			l_servervm as sum(val(servervm))
			Licenses:math(l_server+l_vm+l_servervm)
        }

	}
	`
	return replacer(q, map[string]string{
		"$ID":                         strings.Join(id, ","),
		"$BaseTypeIDs":                baseTypeIDs,
		"$ChildDirectEquipment":       childDirectEquipment,
		"$ChildInDirectEquipment":     childInDirectEquipment,
		"$ChildInDirectSAEquipmentVM": childInDirectSAEquipmentVM,
		"$ReferenceType":              metric.ReferenceType,
		"$NumCores":                   metric.NumCoresAttr,
		"$NumCPU":                     metric.NumCPUAttr,
		"$Scopes":                     strings.Join(scopes, ","),
	})
}
