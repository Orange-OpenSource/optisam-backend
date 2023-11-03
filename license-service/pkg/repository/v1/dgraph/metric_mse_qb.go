package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func queryBuilderMSE(metric *v1.MetricMSEComputed, scopes []string, vmCount, serverCount int32, sa bool, id ...string) string {

	//logger.Log.Sugar().Infow("dgraph/queryBuilderMSE - licensesForMSE", "vmCount", vmCount, "serverCount", serverCount, "sa", sa)
	if (vmCount > 0 && serverCount > 0) || (vmCount > 0 && serverCount == 0 && !sa) {
		q := `
	{
		var(func: uid($ID)){
			product.equipment @filter(eq(equipment.type,virtualmachine)){
			    softpartitionIDs as uid
			}
		} 

		var(func: uid(softpartitionIDs)){
			equipment.parent @filter(eq(equipment.type,$Reference)){
			    serverIDs_c as uid
			}
		} 

		var(func: uid($ID)){
			product.equipment @filter(eq(equipment.type,$Reference)){
			    serverIDs as uid
			}
		}
		
	    var(func:uid(serverIDs)) {
			cpu as equipment.$Reference.$CPU
            cores as equipment.$Reference.$Core
            server_ts as  math(cpu*(max(cores,4)))
		}

		var(func:uid(serverIDs_c)){
			v_cpu as equipment.$Reference.$CPU
            v_cores as equipment.$Reference.$Core
            p_comp as  math(v_cpu*(max(v_cores,4)))
			~equipment.parent @filter(uid(softpartitionIDs)){
				vcpu as equipment.virtualmachine.vcpu
				result as math(max(vcpu,4))
			}
			sumVcpu as sum(val(vcpu))
			vm_comp as sum(val(result))
			serverVMMin as math(min(p_comp,vm_comp))
			serverVMSumUp as math(p_comp+vm_comp)
			server_t as math(cond(sumVcpu<=v_cores,serverVMMin,serverVMSumUp))
		}

	    Licenses() @normalize{
			l_server as sum(val(server_t))
			l_servers as sum(val(server_ts))
			Licenses:math(l_server+l_servers)
		}
	}
	`
		return replacer(q, map[string]string{
			"$ID":        strings.Join(id, ","),
			"$Reference": metric.Reference,
			"$Core":      metric.Core,
			"$CPU":       metric.CPU,
			"$Scopes":    strings.Join(scopes, ","),
		})
	} else if vmCount == 0 && serverCount > 0 && !sa {
		q := `
		{
			var(func: uid($ID)){
				product.equipment @filter(eq(equipment.type,$Reference)){
					serverIDs as uid
				}
			}
			var(func:uid(serverIDs)) {
				cpu as equipment.$Reference.$CPU
				cores as equipment.$Reference.$Core
				server_ts as  math(cpu*(max(cores,4)))
			}
			Licenses() @normalize{
				l_servers as sum(val(server_ts))
				Licenses:math(l_servers)
			}
		}
		`
		return replacer(q, map[string]string{
			"$ID":        strings.Join(id, ","),
			"$Reference": metric.Reference,
			"$Core":      metric.Core,
			"$CPU":       metric.CPU,
			"$Scopes":    strings.Join(scopes, ","),
		})
	} else if vmCount > 0 && serverCount == 0 && sa {
		q := `
		{
			var(func: uid($ID)){
				product.equipment @filter(eq(equipment.type,virtualmachine)){
					softpartitionIDs as uid
				}
			} 
	
			var(func: uid(softpartitionIDs)){
				equipment.parent @filter(eq(equipment.type,$Reference)){
					serverIDs_c as uid
				}
			} 
	
			var(func: uid($ID)){
				product.equipment @filter(eq(equipment.type,$Reference)){
					serverIDs as uid
				}
			}
			var(func:uid(serverIDs)) {
				cpu as equipment.$Reference.$CPU
				cores as equipment.$Reference.$Core
				server_ts as  math(cpu*(max(cores,4)))
			}
			var(func:uid(serverIDs_c)){
				v_cpu as equipment.$Reference.$CPU
				v_cores as equipment.$Reference.$Core
				p_comp as  math(v_cpu*(max(v_cores,4)))
				~equipment.parent @filter(uid(softpartitionIDs)){
					vcpu as equipment.virtualmachine.vcpu
					result as math(max(vcpu,4))
				}
				vm_comp as sum(val(result))
				server_t as math(min(p_comp,vm_comp))
			}
	
			Licenses() @normalize{
				l_server as sum(val(server_t))
				l_servers as sum(val(server_ts))
				Licenses:math(l_server+l_servers)
			}
		}
		`
		return replacer(q, map[string]string{
			"$ID":        strings.Join(id, ","),
			"$Reference": metric.Reference,
			"$Core":      metric.Core,
			"$CPU":       metric.CPU,
			"$Scopes":    strings.Join(scopes, ","),
		})
	}
	return ""
}
