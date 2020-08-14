// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	l_v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/report-service/pkg/repository/v1"
	"optisam-backend/report-service/pkg/repository/v1/postgres/db"
	"strings"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type Worker struct {
	id            string
	licenseClient l_v1.LicenseServiceClient
	reportRepo    repo.Report
	dgraphRepo    repo.DgraphReport
}

type ReportType string

const (
	AcqRightsReport         ReportType = "AcqRightsReport"
	ProductEquipmentsReport ReportType = "ProductEquipmentsReport"
)

type Envelope struct {
	Type     ReportType      `json:"report_type"`
	Scope    string          `json:"scope"`
	JSON     json.RawMessage `json:"json"`
	ReportID int32           `json:"report_id"`
}

type productEquipmentsReportStruct struct {
	SwidTag   []string `json:"swidtag"`
	EquipType string   `json:"equipType"`
	Editor    string   `json:"editor"`
}

type AcqRightsStruct struct {
	SKU            string  `json:"sku"`
	SwidTag        string  `json:"swidtag"`
	Editor         string  `json:"editor"`
	Metric         string  `json:"metric"`
	NumCptLicences int32   `json:"computedLicenses"`
	NumAcqLicences int32   `json:"acquiredLicenses"`
	DeltaNumber    int32   `json:"delta(number)"`
	DeltaCost      float64 `json:"delta(cost)"`
	TotalCost      float64 `json:"totalcost"`
	AvgUnitPrice   float64 `json:"avgunitprice"`
}

type AcqRightsReportStruct struct {
	SwidTag []string `json:"swidtag"`
	Editor  string   `json:"editor"`
}

func NewWorker(id string, reportRepo repo.Report, grpcServers map[string]*grpc.ClientConn, dgraphRepo repo.DgraphReport) *Worker {
	return &Worker{id: id, reportRepo: reportRepo, licenseClient: l_v1.NewLicenseServiceClient(grpcServers["license"]), dgraphRepo: dgraphRepo}
}

func (w *Worker) ID() string {
	return w.id
}

func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	var e Envelope
	err := json.Unmarshal(j.Data, &e)
	if err != nil {
		logger.Log.Error("worker - Unmarshall Error in envelope", zap.Error(err))
		return fmt.Errorf("worker - Unmarshall Error")
	}
	switch e.Type {
	case AcqRightsReport:
		var r AcqRightsReportStruct
		err := json.Unmarshal(e.JSON, &r)
		if err != nil {
			logger.Log.Error("worker - AcqRightsReport - Unmarshall Error", zap.Error(err))
			return fmt.Errorf("worker - AcqRightsReport - Json Marshalling failed")
		}
		var complianceObjects []string
		for _, swidtag := range r.SwidTag {
			resp, err := w.licenseClient.ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{SwidTag: swidtag})
			if err != nil {
				logger.Log.Error("worker - acqrights report - LicenseService - ListAcquiredRightsForProduct", zap.Error(err))
				return fmt.Errorf("worker - acqrights report - LicenseService - ListAcquiredRightsForProduct failed")
			}
			if len(resp.AcqRights) == 0 {
				continue
			}
			for _, a := range resp.AcqRights {
				workerAcqRights := &AcqRightsStruct{
					SKU:            a.SKU,
					SwidTag:        a.SwidTag,
					Editor:         r.Editor,
					Metric:         a.Metric,
					NumCptLicences: a.NumCptLicences,
					NumAcqLicences: a.NumAcqLicences,
					TotalCost:      a.TotalCost,
					DeltaNumber:    a.DeltaNumber,
					DeltaCost:      a.DeltaCost,
					AvgUnitPrice:   a.AvgUnitPrice,
				}
				var acqJson json.RawMessage
				acqJson, err := json.Marshal(workerAcqRights)
				if err != nil {
					logger.Log.Error("worker - ProductEquipmentsReport -  json marshall error", zap.Error(err))
					return fmt.Errorf("worker - ProductEquipmentsReport - Json Marshalling failed")
				}
				complianceObjects = append(complianceObjects, string(acqJson))

			}
		}
		complianceJsonArray := "[" + strings.Join(complianceObjects, ",") + "]"
		fmt.Println(complianceJsonArray)
		rawJson := json.RawMessage(complianceJsonArray)
		bytes, err := rawJson.MarshalJSON()
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport -  json marshall error", zap.Error(err))
			return fmt.Errorf("worker - ProductEquipmentsReport - Json Marshalling failed")
		}
		err = w.reportRepo.InsertReportData(ctx, db.InsertReportDataParams{
			ReportDataJson: bytes,
			ReportID:       e.ReportID,
		})
		if err != nil {
			logger.Log.Error("worker - acqrights report - ReportRepo - InsertReportData", zap.Error(err))
			return fmt.Errorf("worker - acqrights report - ReportRepo - InsertReportData")
		}
		err = w.reportRepo.UpdateReportStatus(ctx, db.UpdateReportStatusParams{ReportStatus: db.ReportStatusCOMPLETED, ReportID: e.ReportID})
		if err != nil {
			logger.Log.Error("worker - acqrights report - ReportRepo - UpdateReportStatus", zap.Error(err))
			return fmt.Errorf("worker - acqrights report - ReportRepo - UpdateReportStatus")
		}
	case ProductEquipmentsReport:
		var r productEquipmentsReportStruct
		err := json.Unmarshal(e.JSON, &r)
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport - Unmarshall Error", zap.Error(err))
			return fmt.Errorf("worker - ProductEquipmentsReport - Json Marshalling failed")
		}

		// Find equipment type parents to make columns
		parents, err := w.dgraphRepo.EquipmentTypeParents(ctx, r.EquipType)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentTypeParents", zap.Error(err))
			return fmt.Errorf("worker - ProductEquipmentsReport - EquipmentTypeParents failed")
		}

		// Find equipmenttype attributes to make columns
		attrs, err := w.dgraphRepo.EquipmentTypeAttrs(ctx, r.EquipType)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentTypeAttrs", zap.Error(err))
			return fmt.Errorf("worker - ProductEquipmentsReport - EquipmentTypeAttrs failed")
		}

		var jsonProductArray []string

		for _, swidtag := range r.SwidTag {

			// Find Equipments in which the product is installed
			productEquipments, err := w.dgraphRepo.ProductEquipments(ctx, swidtag, e.Scope, r.EquipType)
			if err != nil && err != repo.ErrNoData {
				logger.Log.Error("worker - ProductEquipmentsReport -  ProductEquipments", zap.Error(err))
				return fmt.Errorf("worker - ProductEquipmentsReport - ProductEquipments failed")
			}
			// If there are equipments attached with the product
			if err != repo.ErrNoData {
				for _, equipment := range productEquipments {
					var jsonValues []string
					swidtagString := `"swidtag":"` + swidtag + `"`
					editorString := `"editor":"` + r.Editor + `"`
					jsonValues = append(jsonValues, swidtagString, editorString)
					directEquipmentString := `"` + equipment.EquipmentType + `":"` + equipment.EquipmentID + `"`
					jsonValues = append(jsonValues, directEquipmentString)
					// Find all attributes value if the attribute are available
					if attrs != nil {
						attributeJSON, err := w.dgraphRepo.EquipmentAttributes(ctx, equipment.EquipmentID, equipment.EquipmentType, attrs)
						if err != nil {
							logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentAttributes", zap.Error(err))
							return fmt.Errorf("worker - ProductEquipmentsReport - EquipmentAttributes failed")
						}
						attributeSlice := attributeJSON[1 : len(attributeJSON)-1]
						jsonValues = append(jsonValues, string(attributeSlice))
					}
					// Find parentsIDs if there are parents exists
					if parents != nil {
						equipmentParents, err := w.dgraphRepo.EquipmentParents(ctx, equipment.EquipmentID, equipment.EquipmentType, e.Scope)
						if err != nil {
							logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentParents", zap.Error(err))
							return fmt.Errorf("worker - ProductEquipmentsReport - EquipmentParents failed")
						}
						for i := 0; i < len(parents); i++ {
							equipID := findEquipmentID(parents[i], equipmentParents)
							if equipID == "" {
								jsonValues = append(jsonValues, `"`+parents[i]+`":""`)
							} else {
								jsonValues = append(jsonValues, `"`+parents[i]+`":"`+equipID+`"`)
							}

						}
					}
					jsonProductArray = append(jsonProductArray, "{"+strings.Join(jsonValues, ",")+"}")
				}
			}

		}
		finalJSONRes := "[" + strings.Join(jsonProductArray, ",") + "]"
		rawJson := json.RawMessage(finalJSONRes)
		bytes, err := rawJson.MarshalJSON()
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport -  json marshall error", zap.Error(err))
			return fmt.Errorf("worker - ProductEquipmentsReport - Json Marshalling failed")
		}
		err = w.reportRepo.InsertReportData(ctx, db.InsertReportDataParams{
			ReportDataJson: bytes,
			ReportID:       e.ReportID,
		})
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport report - ReportRepo - AppendReportData", zap.Error(err))
			return fmt.Errorf("worker - ProductEquipmentsReport report - ReportRepo - AppendReportData")
		}
		err = w.reportRepo.UpdateReportStatus(ctx, db.UpdateReportStatusParams{ReportStatus: db.ReportStatusCOMPLETED, ReportID: e.ReportID})
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport report - ReportRepo - UpdateReportStatus", zap.Error(err))
			return fmt.Errorf("worker - ProductEquipmentsReport report - ReportRepo - UpdateReportStatus")
		}
	}
	return nil
}

func findEquipmentID(typ string, equipmentParents []*repo.ProductEquipment) string {

	for _, parent := range equipmentParents {
		if typ == parent.EquipmentType {
			return parent.EquipmentID
		}
	}

	return ""
}
