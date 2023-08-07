package dgraph

import (
	"context"
	"encoding/json"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/loader"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

// {name: "SUCCESS - partition to datacentre, agg-venter",
// 			l: NewLicenseRepository(dgClient),
// 			args: args{
// 				ctx: context.Background(),
// 				id:  ID,
// 				mat: &v1.MetricOPSComputed{
// 					EqTypeTree: []*v1.EquipmentType{
// 						&v1.EquipmentType{
// 							Type: "partition",
// 						},
// 						&v1.EquipmentType{
// 							Type: "server",
// 						},
// 						&v1.EquipmentType{
// 							Type: "cluster",
// 						},
// 						&v1.EquipmentType{
// 							Type: "vcenter",
// 						},
// 						// &v1.EquipmentType{
// 						// 	Type: "datacenter",
// 						// },
// 					},
// 					BaseType: &v1.EquipmentType{
// 						Type: "server",
// 					},
// 					AggregateLevel: &v1.EquipmentType{
// 						Type: "vcenter",
// 					},
// 					NumCoresAttr: &v1.Attribute{
// 						Name: "server_coresNumber",
// 					},
// 					NumCPUAttr: &v1.Attribute{
// 						Name: "server_processorsNumber",
// 					},
// 					CoreFactorAttr: &v1.Attribute{
// 						Name: "corefactor_oracle",
// 					},
// 				},
// 			},
// 			want: uint64(111),
// 		},

func TestLicenseRepository_MetricOPSComputedLicenses(t *testing.T) {

	type args struct {
		ctx    context.Context
		id     []string
		mat    *v1.MetricOPSComputed
		scopes string
	}
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()

	ID, err := getUIDForProductXID("ORAC099", []string{"scope1"})
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantErr bool
	}{
		{name: "SUCCESS - partition to datacentre , agg-cluster",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  []string{ID},
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
						{
							Type: "Vcenter",
						},
						{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(112),
		},
		{name: "SUCCESS - partition to datacentre, agg-venter",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: []string{ID},
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
						{
							Type: "Vcenter",
						},
						{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Vcenter",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(111),
		},
		{name: "SUCCESS - partition to Vcenter, agg - cluster",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: []string{ID},
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
						{
							Type: "Vcenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(88),
		},
		{name: "SUCCESS - partition to Cluster, agg-server",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: []string{ID},
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(34),
		},
		{name: "SUCCESS - server to server, agg-server",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: []string{ID},
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{

						{
							Type: "Server",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricOPSComputedLicenses(tt.args.ctx, tt.args.id, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricOPSComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricOPSComputedLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseRepository_MetricOPSComputedLicensesAgg(t *testing.T) {

	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricOPSComputed
		scopes string
	}
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()

	ID, err := getUIDForProductXID("ORAC099", []string{"scope1"})
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	metric := "abc"
	aggName := "xyz"
	aggCleanup, err := aggSetup(metric, ID, aggName, "scope1")
	if !assert.Empty(t, err, "error is not expected in agg setup") {
		return
	}

	defer func() {
		if !assert.Empty(t, aggCleanup(), "error is not expected in aggCleanup") {
			return
		}
	}()

	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantErr bool
	}{
		{name: "SUCCESS - partition to datacentre , agg-cluster",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
						{
							Type: "Vcenter",
						},
						{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(112),
		},
		{name: "SUCCESS - partition to datacentre, agg-venter",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
						{
							Type: "Vcenter",
						},
						{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Vcenter",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(111),
		},
		{name: "SUCCESS - partition to Vcenter, agg - cluster",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
						{
							Type: "Vcenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(88),
		},
		{name: "SUCCESS - partition to Cluster, agg-server",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(34),
		},
		{name: "SUCCESS - server to server, agg-server",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{

						{
							Type: "Server",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricOPSComputedLicensesAgg(tt.args.ctx, aggName, metric, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricOPSComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricOPSComputedLicenses() = %v, want %v", got, tt.want)
			}
		})
	}

	// time.Sleep(10 * time.Minute)
}

func TestLicenseRepository_MetricOPSComputedLicensesForAppProduct(t *testing.T) {
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()
	type args struct {
		ctx    context.Context
		prodID string
		appID  string
		mat    *v1.MetricOPSComputed
		scopes string
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:    context.Background(),
				prodID: "0x28",
				appID:  "A2",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						{
							Type: "Partition",
						},
						{
							Type: "Server",
						},
						{
							Type: "Cluster",
						},
						{
							Type: "Vcenter",
						},
						{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
				scopes: "scope3",
			},
			want: uint64(112),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLicenseRepository(dgClient)
			got, err := l.MetricOPSComputedLicensesForAppProduct(tt.args.ctx, tt.args.prodID, tt.args.appID, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricOPSComputedLicensesForAppProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricOPSComputedLicensesForAppProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func aggSetup(metricName, productID, aggName, scope string) (func() error, error) {
	mu := &api.Mutation{
		CommitNow: true,
		Set: []*api.NQuad{
			{
				Subject:     blankID(aggName),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("product_aggreagtion"),
			},
			{
				Subject:     blankID(aggName),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("ProductAggregation"),
			},
			{
				Subject:     blankID(aggName),
				Predicate:   "product_aggregation.name",
				ObjectValue: stringObjectValue(aggName),
			},
			{
				Subject:     blankID(aggName),
				Predicate:   "scopes",
				ObjectValue: stringObjectValue(scope),
			},
			{
				Subject:   blankID(aggName),
				Predicate: "product_aggregation.products",
				ObjectId:  productID,
			},
			{
				Subject:   productID,
				Predicate: "product.acqRights",
				ObjectId:  blankID("sku1"),
			},
			{
				Subject:     blankID("sku1"),
				Predicate:   "acqRights.metric",
				ObjectValue: stringObjectValue(metricName),
			},
			{
				Subject:     blankID("sku1"),
				Predicate:   "dgraph.name",
				ObjectValue: stringObjectValue("AcquiredRights"),
			},
		},
	}

	assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil {
		return nil, err
	}

	uids := make([]string, 0, len(assigned.Uids))
	for _, uid := range assigned.Uids {
		uids = append(uids, uid)
	}

	return func() error {
		return deleteNodes(uids...)
	}, nil
}

func deleteForXIDs(xids []string, query string) error {
	for _, xid := range xids {
		resp, err := dgClient.NewTxn().Query(context.Background(), strings.Replace(query, "$XID", xid, -1))
		if err != nil {
			return err
		}
		type data struct {
			Data []*id
		}

		var d data
		if err := json.Unmarshal(resp.Json, &d); err != nil {
			return err
		}
		if len(d.Data) == 0 {
			continue
		}
		for _, id := range d.Data {
			if err := deleteNode(id.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteEqMetadaTypes(metas ...string) error {
	q := `
	{
		Data(func:eq( metadata.source,$XID)){
		  uid
		}
	  }
	`
	return deleteForXIDs(metas, q)
}

func deleteEquipmentTypes(eqTypes ...string) error {
	for _, et := range eqTypes {
		if err := dgClient.Alter(context.Background(), &api.Operation{
			DropOp:    api.Operation_TYPE,
			DropValue: "Equipment" + et,
		}); err != nil {
			return err
		}
	}
	q := `
	{
		Data(func:eq(metadata.equipment.type,$XID)){
		  uid
		}
	  }
	`
	return deleteForXIDs(eqTypes, q)
}

func deleteProductEquipmentRelationships() error {
	req := &api.Request{
		Query: `	query{
			products as var(func:has(product.equipment))
			}`,
		Mutations: []*api.Mutation{
			{
				DeleteJson: []byte(`{
					"uid": "uid(products)",
					"product.equipment": null,
					"product.users": null
				  }`),
				// DelNquads: []byte("uid(products) <product.equipment> * .ss"),
				// Del: []*api.NQuad{
				// 	{
				// 		Subject:   "uid(products)",
				// 		Predicate: "product.equipment",
				// 		ObjectId:  "*",
				// 	},
				// },
			},
		},
		CommitNow: true,
	}

	// Update email only if exactly one matching uid is found.
	if _, err := dgClient.NewTxn().Do(context.Background(), req); err != nil {
		return err
	}
	return nil
}

func deleteEquipments(eqTypes ...string) error {
	q := `
	{
		Data(func:eq(equipment.type,$XID)){
		  uid
		}
	  }
	`
	return deleteForXIDs(eqTypes, q)
}

func setup() (func() error, error) {
	// Load Metadata
	config := loader.NewDefaultConfig()
	config.LoadMetadata = true
	config.MasterDir = "testdata"
	config.ScopeSkeleten = "skeletonscope"
	repo := NewLicenseRepository(dgClient)
	config.Repository = repo
	config.IgnoreNew = true
	equipFiles := []string{
		"equipment_cluster.csv",
		"equipment_datacenter.csv",
		"equipment_partition.csv",
		"equipment_server.csv",
		"equipment_vcenter.csv",
	}
	config.MetadataFiles.EquipFiles = equipFiles
	loader.Load(config)

	// LoadEquipments Types
	if err := loader.LoadDefaultEquipmentTypes(repo); err != nil {
		return nil, err
	}

	// Load Product linking
	config = loader.NewDefaultConfig()
	config.IgnoreNew = true
	config.MasterDir = "testdata"
	config.Scopes = []string{"scope1", "scope2"}
	config.LoadStaticData = true
	config.Repository = repo
	config.InstEquipFiles = []string{
		"instances_equipments.csv",
	}
	config.ProductEquipmentFiles = []string{
		"products_equipments.csv",
		"products_equipments_users.csv",
	}
	if err := loader.Load(config); err != nil {
		return nil, err
	}

	// Load Equipments
	config = loader.NewDefaultConfig()
	config.IgnoreNew = true
	config.MasterDir = "testdata"
	config.Scopes = []string{"scope1", "scope2"}
	config.LoadEquipments = true
	config.Repository = repo
	config.EquipmentFiles = equipFiles
	if err := loader.Load(config); err != nil {
		return nil, err
	}
	return func() error {
		if err := deleteEquipmentTypes("Server", "Partition", "Cluster", "Vcenter", "Datacenter"); err != nil {
			return err
		}
		if err := deleteEquipments("Server", "Partition", "Cluster", "Vcenter", "Datacenter"); err != nil {
			return err
		}
		// if err := deleteProductEquipmentRelationships(); err != nil {
		// 	return err
		// }
		equipFilesBasePath := make([]string, len(equipFiles))
		for i := range equipFiles {
			equipFilesBasePath[i] = filepath.Base(equipFiles[i])
		}
		if err := deleteEqMetadaTypes(equipFilesBasePath...); err != nil {
			return err
		}
		return nil
	}, nil
}
func getUIDForProductXID(xid string, scopes []string) (string, error) {
	type id struct {
		ID string
	}
	type data struct {
		IDs []*id
	}

	resp, err := dgClient.NewTxn().Query(context.Background(), `{
	        IDs(func: eq(product.swidtag,"`+xid+`")) `+agregateFilters(scopeFilters(scopes), typeFilters("type_name", "product"))+`{
				ID:uid
			}
	}`)
	if err != nil {
		return "", err
	}

	var d data
	if err := json.Unmarshal(resp.Json, &d); err != nil {
		return "", err
	}
	if len(d.IDs) == 0 {
		return "", v1.ErrNoData
	}
	return d.IDs[0].ID, nil
}
