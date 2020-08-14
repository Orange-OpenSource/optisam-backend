// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"path/filepath"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

const (
	// Server ...
	Server    = "Server"
	Cluster   = "Cluster"
	Vcenter   = "Vcenter"
	Partition = "Partition"
	//DataCenter="Datacenter"
)

var (
	// EqTypeServer is equipment type server with all attributes
	EqTypeServer = &v1.EquipmentType{
		Type:       Server,
		SourceName: "equipment_server.csv",
		Attributes: []*v1.Attribute{
			&v1.Attribute{
				Name:               "HostName",
				Type:               v1.DataTypeString,
				IsIdentifier:       true,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_hostname",
			},
			&v1.Attribute{
				Name:               "ServerCode",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_code",
			},
			&v1.Attribute{
				Name:               "ServerManufacturer",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_manufacturer",
			},
			&v1.Attribute{
				Name:               "ServerModel",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_model",
			},
			&v1.Attribute{
				Name:               "ServerSerialNumber",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_serialNumber",
			},
			&v1.Attribute{
				Name:               "ServerDateInstallation",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_DateInstallation",
			},
			&v1.Attribute{
				Name:               "ServerProprietaryEntity",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_proprietaryEntity",
			},
			&v1.Attribute{
				Name:               "ServerHostingEntity",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_hostingEntity",
			},
			&v1.Attribute{
				Name:               "ServerUserEntity",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_userEntity",
			},
			&v1.Attribute{
				Name:               "ServerSite",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       false,
				IsParentIdentifier: false,
				MappedTo:           "server_Site",
			},
			&v1.Attribute{
				Name:               "ServerCPU",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_cpu",
			},
			&v1.Attribute{
				Name:               "ServerProcessorsNumber",
				Type:               v1.DataTypeInt,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_processorsNumber",
			},
			&v1.Attribute{
				Name:               "ServerCoresNumber",
				Type:               v1.DataTypeInt,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "server_coresNumber",
			},
			&v1.Attribute{
				Name:               "Parent",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsParentIdentifier: true,
				MappedTo:           "parent_id",
			},
			&v1.Attribute{
				Name:         "OracleCoreFactor",
				Type:         v1.DataTypeFloat,
				IsDisplayed:  true,
				IsSearchable: true,
				MappedTo:     "corefactor_oracle",
			},
			&v1.Attribute{
				Name:         "SAG",
				Type:         v1.DataTypeFloat,
				IsDisplayed:  true,
				IsSearchable: true,
				MappedTo:     "sag",
			},
			&v1.Attribute{
				Name:         "PVU",
				Type:         v1.DataTypeInt,
				IsDisplayed:  true,
				IsSearchable: true,
				MappedTo:     "pvu",
			},
		},
	}
	// EqTypeCluster ...
	EqTypeCluster = &v1.EquipmentType{
		Type:       Cluster,
		SourceName: "equipment_cluster.csv",
		Attributes: []*v1.Attribute{
			&v1.Attribute{
				Name:               "ClusterName",
				Type:               v1.DataTypeString,
				IsIdentifier:       true,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "cluster_name",
			},
			&v1.Attribute{
				Name:               "Parent",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsParentIdentifier: true,
				MappedTo:           "parent_id",
			},
		},
	}
	// EqTypeVcenter ...
	EqTypeVcenter = &v1.EquipmentType{
		Type:       Vcenter,
		SourceName: "equipment_vcenter.csv",
		Attributes: []*v1.Attribute{
			&v1.Attribute{
				Name:               "VcenterName",
				Type:               v1.DataTypeString,
				IsIdentifier:       true,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "vcenter_name",
			},
			&v1.Attribute{
				Name:               "Parent",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsParentIdentifier: true,
				MappedTo:           "parent_id",
			},
		},
	}
	// EqTypePartition ...
	EqTypePartition = &v1.EquipmentType{
		Type:       Partition,
		SourceName: "equipment_partition.csv",
		Attributes: []*v1.Attribute{
			&v1.Attribute{
				Name:               "HostName",
				Type:               v1.DataTypeString,
				IsIdentifier:       true,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_hostname",
			},
			&v1.Attribute{
				Name:               "PartitionCode",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_code",
			},
			&v1.Attribute{
				Name:               "PartitionRole",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_role",
			},
			&v1.Attribute{
				Name:               "Environment",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_environment",
			},
			&v1.Attribute{
				Name:               "PartitionShortOs",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_shortOS",
			},
			&v1.Attribute{
				Name:               "PartitionNormalizedOs",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       false,
				IsParentIdentifier: false,
				MappedTo:           "partition_normalizedOS",
			},
			&v1.Attribute{
				Name:               "CPU",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_cpu",
			},
			&v1.Attribute{
				Name:               "ProcessorNumber",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_processorsNumber",
			},
			&v1.Attribute{
				Name:               "CoresNumber",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "partition_coresNumber",
			},
			&v1.Attribute{
				Name:               "Parent",
				Type:               v1.DataTypeString,
				IsDisplayed:        true,
				IsParentIdentifier: true,
				MappedTo:           "parent_id",
			},
		},
	}
	// EqTypeDataCenter ...
	EqTypeDataCenter = &v1.EquipmentType{
		Type:       "Datacenter",
		SourceName: "equipment_datacenter.csv",
		Attributes: []*v1.Attribute{
			&v1.Attribute{
				Name:               "Name",
				Type:               v1.DataTypeString,
				IsIdentifier:       true,
				IsDisplayed:        true,
				IsSearchable:       true,
				IsParentIdentifier: false,
				MappedTo:           "datacenter_name",
			},
		},
	}
)

// LoadDefaultEquipmentTypes ...
func LoadDefaultEquipmentTypes(repo v1.Equipment) error {
	eqTypes := []*v1.EquipmentType{
		EqTypeDataCenter,
		EqTypeVcenter,
		EqTypeCluster,
		EqTypeServer,
		EqTypePartition,
	}
	metas, err := repo.MetadataAllWithType(context.Background(), v1.MetadataTypeEquipment, []string{})
	if err != nil {
		return err
	}

	for i, eqType := range eqTypes {
		for _, m := range metas {
			log.Println(m.Source)
			if m.Source == eqType.SourceName {
				eqType.SourceID = m.ID
			}
		}

		if eqType.SourceID == "" {
			logger.Log.Error("LoadDefaultEquipmentTypes - cannot find metadata for file", zap.String("file_name", eqType.SourceName))
		}

		if i < len(eqTypes) && i > 0 {
			eqType.ParentID = eqTypes[i-1].ID
		}

		if err := LoadEquipmentsType(eqType, repo); err != nil {
			return err
		}

		log.Println(eqType.ID)
	}
	return nil
}

// LoadEquipmentsType ...
func LoadEquipmentsType(eqType *v1.EquipmentType, repo v1.Equipment) error {
	if _, err := repo.CreateEquipmentType(context.Background(), eqType, []string{}); err != nil {
		return err
	}
	return nil
}

func loadEquipmentMetadata(ch chan<- *api.Request, doneChan <-chan struct{}, filename string) {
	log.Println("started metadata loading " + filename)
	defer log.Println("end metadata loading " + filename)
	f, err := readFile(filename)
	if err != nil {
		logger.Log.Error("error opening file", zap.String("filename:", filename), zap.String("reason", err.Error()))
		return
	}

	r := csv.NewReader(f)
	r.Comma = ';'
	columns, err := r.Read()
	if err == io.EOF {
		return
	} else if err != nil {
		logger.Log.Error("error reading header ", zap.String("filename:", filename), zap.String("reason", err.Error()))
		return
	}
	log.Println(columns)
	// uid := uidForXid("equip_metadata")
	uid := "_:equip_metadata"

	nquadsForAttributes := func(attrs []string, id string) []*api.NQuad {
		nqs := make([]*api.NQuad, len(attrs))
		for i := range attrs {
			nqs[i] = &api.NQuad{
				Subject:     uid,
				Predicate:   "metadata.attributes",
				ObjectValue: stringObjectValue(attrs[i]),
			}
		}
		return nqs
	}
	nqs := []*api.NQuad{
		&api.NQuad{
			Subject:     uid,
			Predicate:   "type",
			ObjectValue: stringObjectValue("metadata"),
		},
		&api.NQuad{
			Subject:     uid,
			Predicate:   "metadata.type",
			ObjectValue: stringObjectValue("equipment"),
		},
		&api.NQuad{
			Subject:     uid,
			Predicate:   "metadata.source",
			ObjectValue: stringObjectValue(filepath.Base(filename)),
		},
	}
	nqs = append(nqs, nquadsForAttributes(columns, uid)...)
	select {
	case <-doneChan:
		return
	case ch <- &api.Request{
		Mutations: []*api.Mutation{
			&api.Mutation{
				CommitNow: true,
				Set:       nqs,
			},
		},
	}:
	}
}
