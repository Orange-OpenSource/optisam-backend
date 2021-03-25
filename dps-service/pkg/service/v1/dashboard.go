// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type developmentRate struct {
	Entity string
	Points []float32
	Err    error
}

func (d *dpsServiceServer) DashboardQualityOverview(ctx context.Context, req *v1.DashboardQualityOverviewRequest) (*v1.DashboardQualityOverviewResponse, error) {
	var resp v1.DashboardQualityOverviewResponse

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	currYear, currMonth, _ := time.Now().Date()
	endMonth := (((int(currMonth) - int(req.NoOfDataPoints)) % 12) + 12) % 12
	endYear := currYear
	if int(currMonth)-int(req.NoOfDataPoints) <= 0 {
		endYear--
	}

	res, err := d.dpsRepo.GetEntityMonthWise(ctx, db.GetEntityMonthWiseParams{
		Year:          int32(currYear),
		Month:         int32(currMonth),
		Year_2:        int32(endYear),
		Month_2:       int32(endMonth),
		SimilarEscape: fmt.Sprintf("%s_(applications.csv|products.csv|products_acquiredRights.csv|equipment%%)", req.Scope),
		Scope:         req.Scope})

	if err != nil {
		logger.Log.Error("Failed to fetch failed record from DB ", zap.Error(err))
		return &v1.DashboardQualityOverviewResponse{}, status.Error(codes.Internal, "DBError")
	}
	temp := make(map[string]map[int]int) //map[filename]map[month]count
	temp["applications"] = make(map[int]int)
	temp["products"] = make(map[int]int)
	temp["acqRights"] = make(map[int]int)
	temp["equipments"] = make(map[int]int)
	totalApp, totalProd, totalAcq, totalEquip := 0, 0, 0, 0
	for _, val := range res {
		if val.Filename == strings.ToLower(fmt.Sprintf("%s_applications.csv", req.Scope)) {
			temp["applications"][int(val.Month)] = int(val.Sum)
			totalApp += int(val.Sum)
			continue
		} else if val.Filename == strings.ToLower(fmt.Sprintf("%s_products.csv", req.Scope)) {
			temp["products"][int(val.Month)] = int(val.Sum)
			totalProd += int(val.Sum)
			continue
		} else if val.Filename == strings.ToLower(fmt.Sprintf("%s_products_acquiredrights.csv", req.Scope)) {
			temp["acqRights"][int(val.Month)] = int(val.Sum)
			totalAcq += int(val.Sum)
			continue
		} else {
			temp["equipments"][int(val.Month)] += int(val.Sum)
			totalEquip += int(val.Sum)
		}

	}
	resp.Applications = make([]float32, int(req.NoOfDataPoints))
	resp.Products = make([]float32, int(req.NoOfDataPoints))
	resp.Acqrights = make([]float32, int(req.NoOfDataPoints))
	resp.Equipments = make([]float32, int(req.NoOfDataPoints))

	for i := 0; i < int(req.NoOfDataPoints); i++ {
		index := (((int(currMonth) - 1 - i) % 12) + 12) % 12
		if totalApp > 0 {
			resp.Applications[i] = float32(temp["applications"][index]) * float32(100) / float32(totalApp)
			resp.Applications[i] = float32(math.Round(float64(resp.Applications[i]*100)) / 100)
		}
		if totalProd > 0 {
			resp.Products[i] = float32(temp["products"][index]) * float32(100) / float32(totalProd)
			resp.Products[i] = float32(math.Round(float64(resp.Products[i]*100)) / 100)
		}
		if totalAcq > 0 {
			resp.Acqrights[i] = float32(temp["acqRights"][index]) * float32(100) / float32(totalAcq)
			resp.Acqrights[i] = float32(math.Round(float64(resp.Acqrights[i]*100)) / 100)
		}
		if totalEquip > 0 {
			resp.Equipments[i] = float32(temp["equipments"][index]) * float32(100) / float32(totalEquip)
			resp.Equipments[i] = float32(math.Round(float64(resp.Equipments[i]*100)) / 100)
		}
	}
	return &resp, nil
}

func (d *dpsServiceServer) DashboardDataFailureRate(ctx context.Context, req *v1.DataFailureRateRequest) (*v1.DataFailureRateResponse, error) {
	resp := &v1.DataFailureRateResponse{}
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	prevYear, PrevMon, prevDay := time.Now().Add(time.Hour * 24 * -(30)).Date()
	dbresp, err := d.dpsRepo.GetDataFileRecords(ctx, db.GetDataFileRecordsParams{
		Year:          int32(prevYear),
		Month:         int32(PrevMon),
		Day:           int32(prevDay),
		Scope:         req.Scope,
		SimilarEscape: fmt.Sprintf("%s_(applications|products|instance|products_acquiredRights|equipment%%)%%.csv", req.Scope),
	})
	if err != nil {
		logger.Log.Error("Failed to fetch data file records from DB ", zap.Error(err))
		return resp, status.Error(codes.Internal, "DBError")
	}

	if dbresp.TotalRecords > 0 && dbresp.FailedRecords > 0 {
		resp.FailureRate = (float32(dbresp.FailedRecords) * float32(100)) / float32(dbresp.TotalRecords)
		resp.FailureRate = float32(math.Round(float64(resp.FailureRate*100)) / 100)
	} else {
		return resp, status.Error(codes.Internal, "NoContent")
	}

	return resp, nil
}

func (d *dpsServiceServer) ListFailureReasonsRatio(ctx context.Context, req *v1.ListFailureReasonRequest) (*v1.ListFailureReasonResponse, error) {
	var resp v1.ListFailureReasonResponse
	userErrors := map[string]bool{"InvalidFileName": true,
		"FileNotSupported": true,
		"BadFile":          true,
		"NoDataInFile":     true,
		"HeadersMissing":   true,
		"InsufficentData":  true,
	}
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	qYear, qMon, qDay := time.Now().Add(time.Hour * 24 * (-30)).Date()
	dbresp, err := d.dpsRepo.GetFailureReasons(ctx, db.GetFailureReasonsParams{
		Year:  int32(qYear),
		Month: int32(qMon),
		Day:   int32(qDay),
		Data:  json.RawMessage(fmt.Sprintf("%s", req.GetScope()))})
	if err != nil {
		logger.Log.Error("Failed to fetch failed reaons from DB ", zap.Error(err))
		return &resp, status.Error(codes.Internal, "DBError")
	}
	var totalFailure int64
	resp.FailureReasons = make(map[string]float32)
	if len(dbresp) > 0 {
		for _, val := range dbresp {
			totalFailure += val.FailedRecords
			if userErrors[val.Comments.String] {
				resp.FailureReasons[val.Comments.String] = float32(val.FailedRecords)
			} else {
				resp.FailureReasons["InternalError"] += float32(val.FailedRecords)
			}
		}

		for key, val := range resp.FailureReasons {
			resp.FailureReasons[key] = float32(math.Round(float64((val*float32(100))/float32(totalFailure))*float64(100))) / float32(100)
		}
	} else {
		return &resp, status.Error(codes.Internal, "NoContent")
	}
	return &resp, nil
}
