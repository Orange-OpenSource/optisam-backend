package fileworker

import (
	"bufio"
	"reflect"
	"strings"
	"testing"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/worker/models"
)

func Test_dpsServiceServer_getProducts(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in products.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 6}
				hdrs.IndexesOfHeaders = map[string]int{
					"name":       0,
					"version":    1,
					"category":   2,
					"editor":     3,
					"swidtag":    4,
					"isoptionof": 5,
					"flag":       6}
				data := "n1;v1;c1;e1;swid1;o1;1\nn3;v3;c3;e3;swid3;o3;1\nn1;v2;c1;e2;swid1;o1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				DuplicateRecords: []interface{}{
					models.ProductInfo{
						Name:    "n1",
						Version: "v1",
						Editor:  "e1",
						SwidTag: "swid1",
						Action:  "UPSERT"},
				},
				TotalCount: 3,
				Products: map[string]models.ProductInfo{
					"swid1": {
						Name:    "n1",
						Version: "v2",
						Editor:  "e2",
						SwidTag: "swid1",
						Action:  "UPSERT",
					},
					"swid3": {
						Name:    "n3",
						Version: "v3",
						Editor:  "e3",
						SwidTag: "swid3",
						Action:  "UPSERT",
					},
				},
			},
		},
		{
			name:    "No Duplicate Records in products.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 6}
				hdrs.IndexesOfHeaders = map[string]int{
					"name":       0,
					"version":    1,
					"category":   2,
					"editor":     3,
					"swidtag":    4,
					"isoptionof": 5,
					"flag":       6}
				data := "n1;v1;c1;e1;swid1;o1;1\nn3;v3;c3;e3;swid3;o3;1\nn2;v2;c2;e2;swid2;o2;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				TotalCount: 3,
				Products: map[string]models.ProductInfo{
					"swid1": {
						Name:    "n1",
						Version: "v1",
						Editor:  "e1",
						SwidTag: "swid1",
						Action:  "UPSERT",
					},
					"swid3": {
						Name:    "n3",
						Version: "v3",
						Editor:  "e3",
						SwidTag: "swid3",
						Action:  "UPSERT",
					},
					"swid2": {
						Name:    "n2",
						Version: "v2",
						Editor:  "e2",
						SwidTag: "swid2",
						Action:  "UPSERT",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getProducts(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getProducts expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getProducts output mismatch  got = %v, want %v", got, tt.out)
			}
		})
	}
}

func Test_dpsServiceServer_getApplications(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in applications.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 5}
				hdrs.IndexesOfHeaders = map[string]int{
					"application_id": 0,
					"name":           1,
					"version":        2,
					"owner":          3,
					"domain":         4,
					"flag":           5}
				data := "a1;n1;v1;o1;d1;1\na2;n2;v2;o2;d2;1\na1;n1;v2;o2;d1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				DuplicateRecords: []interface{}{
					models.ApplicationInfo{
						Name:    "n1",
						Version: "v1",
						ID:      "a1",
						Owner:   "o1",
						Domain:  "d1",
						Action:  "UPSERT"},
				},
				TotalCount: 3,
				Applications: map[string]models.ApplicationInfo{
					"a1": {
						Name:    "n1",
						Version: "v2",
						Owner:   "o2",
						Domain:  "d1",
						ID:      "a1",
						Action:  "UPSERT",
					},
					"a2": {
						Name:    "n2",
						Version: "v2",
						ID:      "a2",
						Domain:  "d2",
						Owner:   "o2",
						Action:  "UPSERT",
					},
				},
			},
		},
		{
			name:    "No Duplicate Records in products.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 5}
				hdrs.IndexesOfHeaders = map[string]int{
					"application_id": 0,
					"name":           1,
					"version":        2,
					"owner":          3,
					"domain":         4,
					"flag":           5}
				data := "a1;n1;v1;o1;d1;1\na2;n2;v2;o2;d2;1\na3;n3;v3;o3;d3;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				TotalCount: 3,
				Applications: map[string]models.ApplicationInfo{
					"a1": {
						Name:    "n1",
						Version: "v1",
						ID:      "a1",
						Domain:  "d1",
						Owner:   "o1",
						Action:  "UPSERT",
					},
					"a2": {
						Name:    "n2",
						Version: "v2",
						ID:      "a2",
						Domain:  "d2",
						Owner:   "o2",
						Action:  "UPSERT",
					},
					"a3": {
						Name:    "n3",
						Version: "v3",
						ID:      "a3",
						Domain:  "d3",
						Owner:   "o3",
						Action:  "UPSERT",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getApplications(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplications expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getApplications output mismatch  got = %v, want %v", got, tt.out)
			}
		})
	}
}

func Test_dpsServiceServer_getAcqRightsOfProducts(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in acqrights.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 21}
				hdrs.IndexesOfHeaders = map[string]int{
					"product_version":             0,
					"sku":                         1,
					"swidtag":                     2,
					"product_name":                3,
					"editor":                      4,
					"metric":                      5,
					"acquired_licenses":           6,
					"maintenance_licenses":        7,
					"unit_price":                  8,
					"maintenance_unit_price":      9,
					"total_license_cost":          10,
					"total_maintenance_cost":      11,
					"total_cost":                  12,
					"maintenance_start":           13,
					"maintenance_end":             14,
					"corporate_sourcing_contract": 15,
					"ordering_date":               16,
					"software_provider":           17,
					"maintenance_provider":        18,
					"last_purchased_order":        19,
					"support_number":              20,
					"flag":                        21}
				data := ""
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				AcqRights: map[string]models.AcqRightsInfo{},
			},
		},
		{
			name:    "No Duplicate Records in acqrights.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 21}
				hdrs.IndexesOfHeaders = map[string]int{
					"product_version":             0,
					"sku":                         1,
					"swidtag":                     2,
					"product_name":                3,
					"editor":                      4,
					"metric":                      5,
					"acquired_licenses":           6,
					"maintenance_licenses":        7,
					"unit_price":                  8,
					"maintenance_unit_price":      9,
					"total_license_cost":          10,
					"total_maintenance_cost":      11,
					"total_cost":                  12,
					"maintenance_start":           13,
					"maintenance_end":             14,
					"corporate_sourcing_contract": 15,
					"ordering_date":               16,
					"software_provider":           17,
					"maintenance_provider":        18,
					"last_purchased_order":        19,
					"support_number":              20,
					"flag":                        21}
				data := ""
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				AcqRights: map[string]models.AcqRightsInfo{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAcqRightsOfProducts(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getAcqRightsOfProducts expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getAcqRightsOfProducts output mismatch  got = %+v, want %+v", got, tt.out)
			}
		})
	}
}

func Test_dpsServiceServer_getApplicationsAndProducts(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in application_products.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"application_id": 0,
					"swidtag":        1,
					"flag":           2}
				data := "a1;p1;1\na1;p1;1\na1;p1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				DuplicateRecords: []interface{}{
					models.ProdApplink{
						ProdID: "p1",
						AppID:  "a1",
						Action: "UPSERT",
					},
					models.ProdApplink{
						ProdID: "p1",
						AppID:  "a1",
						Action: "UPSERT",
					},
				},
				TotalCount: 3,
				AppProducts: map[string]map[string][]string{
					"UPSERT": {
						"p1": {"a1"},
					},
					"DELETE": {},
				},
			},
		},
		{
			name:    "No Duplicate Records in application_products.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"application_id": 0,
					"swidtag":        1,
					"flag":           2}
				data := "a1;p1;1\na2;p1;1\na3;p1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				TotalCount: 3,
				AppProducts: map[string]map[string][]string{
					"UPSERT": {
						"p1": {"a1", "a2", "a3"},
					},
					"DELETE": {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getApplicationsAndProducts(tt.setup())
			(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationsAndProducts expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getApplicationsAndProducts output mismatch  got = %+v, want %+v", got, tt.out)
			}
		})
	}
}

func Test_dpsServiceServer_getInstancesOfProducts(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in instance_products.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"instance_id": 0,
					"swidtag":     1,
					"flag":        2}
				data := "a1;p1;1\na1;p1;1\na1;p1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				DuplicateRecords: []interface{}{
					models.ProdInstancelink{
						ProdID:     "p1",
						InstanceID: "a1",
						Action:     "UPSERT",
					},
					models.ProdInstancelink{
						ProdID:     "p1",
						InstanceID: "a1",
						Action:     "UPSERT",
					},
				},
				TotalCount: 3,
				ProdInstances: map[string]map[string][]string{
					"UPSERT": {
						"a1": {"p1"},
					},
					"DELETE": {},
				},
			},
		},
		{
			name:    "No Duplicate Records in instance_products.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"instance_id": 0,
					"swidtag":     1,
					"flag":        2}
				data := "a1;p1;1\na2;p1;1\na3;p1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				TotalCount: 3,
				ProdInstances: map[string]map[string][]string{
					"UPSERT": {
						"a1": {"p1"},
						"a2": {"p1"},
						"a3": {"p1"},
					},
					"DELETE": {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInstancesOfProducts(tt.setup())
			(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationsAndProducts expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getApplicationsAndProducts output mismatch  got = %+v, want %+v", got, tt.out)
			}
		})
	}
}

func Test_dpsServiceServer_getInstanceOfApplications(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in instance_application.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"instance_id":    0,
					"application_id": 1,
					"environment":    2,
					"flag":           3}
				data := "i1;a1;e1;1\ni1;a1;e1;1\ni2;a1;e1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				DuplicateRecords: []interface{}{
					models.AppInstanceLink{
						AppID:      "a1",
						InstanceID: "i1",
						Env:        "e1",
						Action:     "UPSERT",
					},
				},
				TotalCount: 3,
				AppInstances: map[string][]models.AppInstance{
					"a1": {
						{
							ID:     "i1",
							Env:    "e1",
							Action: "UPSERT",
						},
						{
							ID:     "i2",
							Env:    "e1",
							Action: "UPSERT",
						},
					},
				},
			},
		},
		{
			name:    "No Duplicate Records in instance_application.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"instance_id":    0,
					"application_id": 1,
					"environment":    2,
					"flag":           3}
				data := "i1;a1;e1;1\ni2;a1;e1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				TotalCount: 2,
				AppInstances: map[string][]models.AppInstance{
					"a1": {
						{
							ID:     "i1",
							Env:    "e1",
							Action: "UPSERT",
						},
						{
							ID:     "i2",
							Env:    "e1",
							Action: "UPSERT",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInstanceOfApplications(tt.setup())
			(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationsAndProducts expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getApplicationsAndProducts output mismatch  got = %+v, want %+v", got, tt.out)
			}
		})
	}
}

func Test_dpsServiceServer_getEquipmentsOfProducts(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in products_equipment.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 3}
				hdrs.IndexesOfHeaders = map[string]int{
					"swidtag":         0,
					"equipment_id":    1,
					"allocatedmetric": 2,
					"allocatedusers":  3,
					"flag":            4}
				data := "p1;e1;met;1;1\np1;e1;met;1;1\np1;e1;met;1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				DuplicateRecords: []interface{}{
					models.ProductEquipmentLink{
						ProdID:          "p1",
						EquipID:         "e1",
						AllocatedMetric: "met",
						AllocatedUsers:  "1",
						Action:          "UPSERT",
					},
				},
				TotalCount: 3,
				ProdEquipments: map[string]map[string][]models.ProdEquipemtInfo{
					"UPSERT": {
						"p1": {{EquipID: "e1", SwidTag: "p1", AllocatedMetric: "met", AllocatedUsers: "1", Action: "UPSERT"}},
						"p2": {{EquipID: "e1", SwidTag: "p1", AllocatedMetric: "met", AllocatedUsers: "1", Action: "UPSERT"}},
					},
					"DELETE": {},
				},
			},
		},
		{
			name:    "No Duplicate Records in products_equipment.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 3}
				hdrs.IndexesOfHeaders = map[string]int{
					"swidtag":        0,
					"equipment_id":   1,
					"allocatedusers": 2,
					"flag":           3}
				data := "p1;e1;1;1\np3;e1;1;1\np2;e1;1;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				TotalCount: 3,
				ProdEquipments: map[string]map[string][]models.ProdEquipemtInfo{
					"UPSERT": {
						"p1": {{EquipID: "e1", SwidTag: "", AllocatedMetric: "", AllocatedUsers: "1", Action: ""}},
						"p2": {{EquipID: "e1", SwidTag: "", AllocatedMetric: "", AllocatedUsers: "1", Action: ""}},
						"p3": {{EquipID: "e1", SwidTag: "", AllocatedMetric: "", AllocatedUsers: "1", Action: ""}},
					},
					"DELETE": {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEquipmentsOfProducts(tt.setup())
			(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationsAndProducts expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getApplicationsAndProducts output mismatch  got = %+v, want %+v", got, tt.out)
			}
		})
	}
}

func Test_dpsServiceServer_getEquipmentsOnInstances(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*bufio.Scanner, models.HeadersInfo)
		out     models.FileData
		wantErr bool
	}{
		{
			name:    "Duplicate Records in equipment.instance.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"equipment_id": 0,
					"instance_id":  1,
					"flag":         2}
				data := "e1;i1;1\ne1;i1;1\ne2;i2;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				DuplicateRecords: []interface{}{
					models.EquipmentApplicationLink{
						AppID:   "i1",
						EquipID: "e1",
						Action:  "UPSERT",
					},
				},
				TotalCount: 3,
				EquipApplications: map[string]map[string][]string{
					"UPSERT": {
						"i1": {"e1"},
						"i2": {"e2"},
					},
					"DELETE": {},
				},
			},
		},
		{
			name:    "No Duplicate Records in equipment.instance.csv",
			wantErr: false,
			setup: func() (*bufio.Scanner, models.HeadersInfo) {
				hdrs := models.HeadersInfo{MaxIndexVal: 2}
				hdrs.IndexesOfHeaders = map[string]int{
					"equipment_id": 0,
					"instance_id":  1,
					"flag":         2}
				data := "e1;i1;1\ne1;i2;1\ne2;i3;1"
				scanner := bufio.NewScanner(strings.NewReader(data))
				return scanner, hdrs
			},
			out: models.FileData{
				TotalCount: 3,
				EquipApplications: map[string]map[string][]string{
					"UPSERT": {
						"i1": {"e1"},
						"i2": {"e1"},
						"i3": {"e2"},
					},
					"DELETE": {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEquipmentsOnApplication(tt.setup())
			(tt.setup())
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationsAndProducts expected error mismatch  = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("getApplicationsAndProducts output mismatch  got = %+v, want %+v", got, tt.out)
			}
		})
	}
}
