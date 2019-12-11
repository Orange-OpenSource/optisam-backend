// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package querybuilders

import (
	//"github.com/dgraph-io/dgraph/gql"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"
)

func Test_query_builder(t *testing.T) {
	type args struct {
		ops *v1.MetricOPSComputed
		id  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// {name: "test1",
		// 	args: args{
		// 		ops: opsMatrix{
		// 			EqTypeTree: []*v1.EquipmentType{
		// 				&v1.EquipmentType{
		// 					Type: "PartitionChild",
		// 				},
		// 				&v1.EquipmentType{
		// 					Type: "Partition",
		// 				},
		// 				&v1.EquipmentType{
		// 					Type: "Server",
		// 				},
		// 				&v1.EquipmentType{
		// 					Type: "Cluster",
		// 				},
		// 				&v1.EquipmentType{
		// 					Type: "Vcenter",
		// 				},
		// 				&v1.EquipmentType{
		// 					Type: "Datacenter",
		// 				},
		// 			},
		// 			BaseType: &v1.EquipmentType{
		// 				Type: "Server",
		// 			},
		// 			AggregateLevel: &v1.EquipmentType{
		// 				Type: "Cluster",
		// 			},
		// 			Cores: &v1.Attribute{
		// 				Name: "CoresNumber",
		// 			},
		// 			CPU: &v1.Attribute{
		// 				Name: "ProcessorNumber",
		// 			},
		// 			CoreFactor: &v1.Attribute{
		// 				Name: "CoreFactor",
		// 			},
		// 		},
		// 		id: "0x4567",
		// 	},
		// },
		{name: "test2",
			args: args{
				ops: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{

						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "CoreFactor",
					},
				},
				id: "0x2718",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queryBuilder(tt.args.ops, tt.args.id); got != tt.want {
				fmt.Println(got)
				//t.Errorf("query_builder() = %v, want %v", got, tt.want)
			}
		})
	}
}
