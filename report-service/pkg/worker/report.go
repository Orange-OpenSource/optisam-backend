package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1/postgres/db"
	l_v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/thirdparty/license-service/pkg/api/v1"
	p_v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/thirdparty/product-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type Worker struct {
	id            string
	licenseClient l_v1.LicenseServiceClient
	productClient p_v1.ProductServiceClient
	reportRepo    repo.Report
	dgraphRepo    repo.DgraphReport
	maxRetries    int
}

type ReportType string

const (
	AcqRightsReport             ReportType = "AcqRightsReport"
	ProductEquipmentsReport     ReportType = "ProductEquipmentsReport"
	ScopeExpensesByEditorReport ReportType = "ScopeExpensesByEditorReport"
)

type Envelope struct {
	Type     ReportType      `json:"report_type"`
	Scope    string          `json:"scope"`
	JSON     json.RawMessage `json:"json"`
	ReportID int32           `json:"report_id"`
}

type productEquipmentsReportStruct struct {
	// SwidTag   []string `json:"swidtag"`
	EquipType string `json:"equipType"`
	Editor    string `json:"editor"`
}

type AcqRightsStruct struct {
	SKU                string  `json:"sku"`
	AggregationName    string  `json:"aggregationName"`
	SwidTag            string  `json:"swidtags"`
	Editor             string  `json:"editor"`
	Product            string  `json:"product"`
	Metric             string  `json:"metric"`
	NumCptLicences     int32   `json:"computedLicenses"`
	ComputationDetails string  `json:"computationDetails"`
	NumAcqLicences     int32   `json:"acquiredLicenses"`
	DeltaNumber        int32   `json:"delta(licenses)"`
	DeltaCost          float64 `json:"delta(cost)"`
	TotalCost          float64 `json:"totalcost"`
	AvgUnitPrice       float64 `json:"avgunitprice"`
}

type AcqRightsReportStruct struct {
	Editor string `json:"editor"`
}

func NewWorker(id string, reportRepo repo.Report, grpcServers map[string]*grpc.ClientConn, dgraphRepo repo.DgraphReport, retries int) *Worker {
	return &Worker{id: id, reportRepo: reportRepo, licenseClient: l_v1.NewLicenseServiceClient(grpcServers["license"]), productClient: p_v1.NewProductServiceClient(grpcServers["product"]), dgraphRepo: dgraphRepo, maxRetries: retries}
}

func (w *Worker) ID() string {
	return w.id
}

func handleReportFailure(ctx context.Context, j *job.Job, err error, w *Worker, id int32) error {
	if err != nil && int(j.RetryCount.Int32) >= w.maxRetries {
		j.Comments = sql.NullString{String: err.Error(), Valid: true}
		logger.Log.Error("Report worker failed", zap.Int32("reporId", id), zap.Error(err))
		err = w.reportRepo.UpdateReportStatus(ctx, db.UpdateReportStatusParams{ReportStatus: db.ReportStatusFAILED, ReportID: id})
		if err != nil {
			logger.Log.Error("worker - handleReportFailure - UpdateReportStatus", zap.Error(err))
		}
	}
	return err
}

// nolint: funlen, gocyclo
func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	var e Envelope
	err := json.Unmarshal(j.Data, &e)
	if err != nil {
		logger.Log.Error("worker - Unmarshall Error in envelope", zap.Error(err))
		return handleReportFailure(ctx, j, fmt.Errorf("worker - Unmarshall Error"), w, e.ReportID)
	}
	switch e.Type {
	case AcqRightsReport:
		var r AcqRightsReportStruct
		err := json.Unmarshal(e.JSON, &r)
		if err != nil {
			logger.Log.Error("worker - AcqRightsReport - Unmarshall Error", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - AcqRightsReport - Json Marshalling failed"), w, e.ReportID)
		}
		var complianceObjects []string
		resp, err := w.licenseClient.GetOverAllCompliance(ctx, &l_v1.GetOverAllComplianceRequest{
			Scope:  e.Scope,
			Editor: r.Editor,
		})
		if err != nil {
			logger.Log.Error("worker - acqrights report - LicenseService - GetOverAllCompliance", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - acqrights report - LicenseService - GetOverAllCompliance failed"), w, e.ReportID)
		}
		// fmt.Println("resp", resp.AcqRights)
		for _, a := range resp.AcqRights {
			workerAcqRights := &AcqRightsStruct{
				SKU:                a.SKU,
				AggregationName:    a.AggregationName,
				SwidTag:            a.SwidTags,
				Editor:             r.Editor,
				Product:            a.ProductNames,
				Metric:             a.Metric,
				NumCptLicences:     a.NumCptLicences,
				ComputationDetails: a.ComputedDetails,
				NumAcqLicences:     a.NumAcqLicences,
				TotalCost:          a.TotalCost,
				DeltaNumber:        a.DeltaNumber,
				DeltaCost:          a.DeltaCost,
				AvgUnitPrice:       a.AvgUnitPrice,
			}
			if a.MetricNotDefined {
				workerAcqRights.ComputationDetails = "Metric Not Defined"
			}
			if a.NotDeployed {
				workerAcqRights.ComputationDetails = "Product Not Deployed"
			}
			var acqJSON json.RawMessage
			acqJSON, error := json.Marshal(workerAcqRights)
			if error != nil {
				logger.Log.Error("worker - AcqRightsReport -  json marshall error", zap.Error(error))
				return handleReportFailure(ctx, j, fmt.Errorf("worker - AcqRightsReport - Json Marshalling failed"), w, e.ReportID)
			}
			// fmt.Println("acqjson string", string(acqJSON))
			complianceObjects = append(complianceObjects, string(acqJSON))
		}
		complianceJSONArray := "[" + strings.Join(complianceObjects, ",") + "]"
		// fmt.Println("compliance array", complianceJSONArray)
		rawJSON := json.RawMessage(complianceJSONArray)
		bytes, err := rawJSON.MarshalJSON()
		if err != nil {
			logger.Log.Error("worker - AcqRightsReport -  json marshall error", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - AcqRightsReport - Json Marshalling failed"), w, e.ReportID)
		}
		err = w.reportRepo.InsertReportData(ctx, db.InsertReportDataParams{
			ReportDataJson: bytes,
			ReportID:       e.ReportID,
		})
		if err != nil {
			logger.Log.Error("worker - acqrights report - ReportRepo - InsertReportData", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - acqrights report - ReportRepo - InsertReportData"), w, e.ReportID)
		}
		err = w.reportRepo.UpdateReportStatus(ctx, db.UpdateReportStatusParams{ReportStatus: db.ReportStatusCOMPLETED, ReportID: e.ReportID})
		if err != nil {
			logger.Log.Error("worker - acqrights report - ReportRepo - UpdateReportStatus", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - acqrights report - ReportRepo - UpdateReportStatus"), w, e.ReportID)
		}
	case ProductEquipmentsReport:
		var r productEquipmentsReportStruct
		err := json.Unmarshal(e.JSON, &r)
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport - Unmarshall Error", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport - Json Marshalling failed"), w, e.ReportID)
		}

		// Find equipment type parents to make columns
		parents, err := w.dgraphRepo.EquipmentTypeParents(ctx, r.EquipType, e.Scope)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentTypeParents", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport - EquipmentTypeParents failed"), w, e.ReportID)
		}

		// Find equipmenttype attributes to make columns
		attrs, err := w.dgraphRepo.EquipmentTypeAttrs(ctx, r.EquipType, e.Scope)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentTypeAttrs", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport - EquipmentTypeAttrs failed"), w, e.ReportID)
		}

		var jsonProductArray []string

		// for _, swidtag := range r.SwidTag {

		// Find Equipments on which the product is installed
		productEquipments, err := w.dgraphRepo.ProductEquipments(ctx, r.Editor, e.Scope, r.EquipType)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("worker - ProductEquipmentsReport -  ProductEquipments", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport - ProductEquipments failed"), w, e.ReportID)
		}
		// If there are equipments attached with the product
		if err != repo.ErrNoData {
			for _, pro := range productEquipments {

				for _, equipment := range pro.Equipments {
					var jsonValues []string
					swidtagString := `"swidtag":"` + pro.Swidtag + `"`
					editorString := `"editor":"` + r.Editor + `"`
					productString := `"product":"` + pro.ProductName + `"`
					jsonValues = append(jsonValues, swidtagString, editorString, productString)
					directEquipmentString := `"` + equipment.EquipmentType + `":"` + equipment.EquipmentID + `"`
					jsonValues = append(jsonValues, directEquipmentString)
					// Find all attributes value if the attribute are available
					if attrs != nil {
						attributeJSON, error := w.dgraphRepo.EquipmentAttributes(ctx, equipment.EquipmentID, equipment.EquipmentType, attrs, e.Scope)
						if error != nil {
							logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentAttributes", zap.Error(error))
							return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport - EquipmentAttributes failed"), w, e.ReportID)
						}
						attributeSlice := attributeJSON[1 : len(attributeJSON)-1]
						jsonValues = append(jsonValues, string(attributeSlice))
					}
					// Find parentsIDs if there are parents exists
					if parents != nil {
						equipmentParents, error := w.dgraphRepo.EquipmentParents(ctx, equipment.EquipmentID, equipment.EquipmentType, e.Scope)
						if error != nil && error != repo.ErrNoData {
							logger.Log.Error("worker - ProductEquipmentsReport -  EquipmentParents", zap.Error(error))
							return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport - EquipmentParents failed"), w, e.ReportID)
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
		rawJSON := json.RawMessage(finalJSONRes)
		bytes, err := rawJSON.MarshalJSON()
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport -  json marshall error", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport - Json Marshalling failed"), w, e.ReportID)
		}
		err = w.reportRepo.InsertReportData(ctx, db.InsertReportDataParams{
			ReportDataJson: bytes,
			ReportID:       e.ReportID,
		})
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport report - ReportRepo - AppendReportData", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport report - ReportRepo - AppendReportData"), w, e.ReportID)
		}
		err = w.reportRepo.UpdateReportStatus(ctx, db.UpdateReportStatusParams{ReportStatus: db.ReportStatusCOMPLETED, ReportID: e.ReportID})
		if err != nil {
			logger.Log.Error("worker - ProductEquipmentsReport report - ReportRepo - UpdateReportStatus", zap.Error(err))
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ProductEquipmentsReport report - ReportRepo - UpdateReportStatus"), w, e.ReportID)
		}
	case ScopeExpensesByEditorReport:
		logger.Log.Sugar().Infow("ScopeExpensesByEditorReport Report genration started ", "scope", e.Scope)
		editorResponse, err := w.productClient.GetEditorExpensesByScope(ctx, &p_v1.EditorExpensesByScopeRequest{
			Scope: e.Scope,
		})
		if err != nil {
			logger.Log.Sugar().Errorw("Report Worker - ScopeExpensesByEditorReport - ProductService - GetEditorExpensesByScope",
				"scope", e.Scope,
				"error", err.Error(),
			)
			return handleReportFailure(ctx, j, fmt.Errorf("Report Worker - ScopeExpensesByEditorReport - ProductService - GetEditorExpensesByScope failed"), w, e.ReportID)
		}
		if len(editorResponse.EditorExpensesByScope) > 0 {
			var jsonEditorArray []string
			for _, editorData := range editorResponse.EditorExpensesByScope {

				var jsonValues []string
				editorString := `"editor":"` + editorData.EditorName + `"`
				purCostString := `"purchaseCost":"` + fmt.Sprintf("%.2f", editorData.TotalPurchaseCost) + `"`
				maintainceCostString := `"maintenanceCost":"` + fmt.Sprintf("%.2f", editorData.TotalMaintenanceCost) + `"`
				totalCostString := `"totalCost":"` + fmt.Sprintf("%.2f", editorData.TotalCost) + `"`
				jsonValues = append(jsonValues, editorString, purCostString, maintainceCostString, totalCostString)
				jsonEditorArray = append(jsonEditorArray, "{"+strings.Join(jsonValues, ",")+"}")
			}

			finalJSONRes := "[" + strings.Join(jsonEditorArray, ",") + "]"
			rawJSON := json.RawMessage(finalJSONRes)
			bytes, err := rawJSON.MarshalJSON()
			if err != nil {
				logger.Log.Sugar().Errorw("Report Worker - ScopeExpensesByEditorReport - json marshall error ",
					"scope", e.Scope,
					"ReportID", e.ReportID,
					"ReportDataJson", rawJSON,
					"error", err.Error(),
				)
				return handleReportFailure(ctx, j, fmt.Errorf("worker - ScopeExpensesByEditorReport - Json Marshalling failed"), w, e.ReportID)
			}
			err = w.reportRepo.InsertReportData(ctx, db.InsertReportDataParams{
				ReportDataJson: bytes,
				ReportID:       e.ReportID,
			})
			if err != nil {
				logger.Log.Sugar().Errorw("Report Worker - ScopeExpensesByEditorReport - reportRepo - InsertReportData",
					"scope", e.Scope,
					"ReportID", e.ReportID,
					"ReportDataJson", rawJSON,
					"error", err.Error(),
				)
				return handleReportFailure(ctx, j, fmt.Errorf("worker - ScopeExpensesByEditorReport report - ReportRepo - AppendReportData"), w, e.ReportID)
			}

		} else {
			logger.Log.Sugar().Debugw("Report Worker - ScopeExpensesByEditorReport - reportRepo - UpdateReportStatus",
				"scope", e.Scope,
				"ReportID", e.ReportID,
				"ReportStatus", db.ReportStatusCOMPLETED,
				"error", "no data found",
			)
		}
		err = w.reportRepo.UpdateReportStatus(ctx, db.UpdateReportStatusParams{ReportStatus: db.ReportStatusCOMPLETED, ReportID: e.ReportID})
		if err != nil {
			logger.Log.Sugar().Errorw("Report Worker - ScopeExpensesByEditorReport - reportRepo - UpdateReportStatus",
				"scope", e.Scope,
				"ReportID", e.ReportID,
				"ReportStatus", db.ReportStatusCOMPLETED,
				"error", err.Error(),
			)
			return handleReportFailure(ctx, j, fmt.Errorf("worker - ScopeExpensesByEditorReport report - ReportRepo - UpdateReportStatus"), w, e.ReportID)
		}
	}
	return nil
}

func findEquipmentID(typ string, equipmentParents []*repo.Equipment) string {

	for _, parent := range equipmentParents {
		if typ == parent.EquipmentType {
			return parent.EquipmentID
		}
	}

	return ""
}
