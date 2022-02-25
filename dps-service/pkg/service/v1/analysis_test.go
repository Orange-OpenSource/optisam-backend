package v1

import (
	"context"
	"fmt"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	repo "optisam-backend/dps-service/pkg/repository/v1"
	queuemock "optisam-backend/dps-service/pkg/repository/v1/queuemock"
	"testing"

	"github.com/golang/mock/gomock"
)

func Test_DataAnalysis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var rep repo.Dps
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.DataAnalysisRequest
		setup   func(*v1.DataAnalysisRequest)
		output  *v1.DataAnalysisResponse
		wantErr bool
	}{
		{
			name: "claims Not found",
			ctx:  context.Background(),
			input: &v1.DataAnalysisRequest{
				Scope: "Scope1",
				File:  "Scope1_applications.xlsx",
			},
			setup:   func(*v1.DataAnalysisRequest) {},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: true,
		},
		{
			name: "Scope Not found",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "Scope10",
				File:  "Scope0_applications.xlsx",
			},
			setup:   func(*v1.DataAnalysisRequest) {},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: true,
		},
		{
			name: "Invalid File Extension",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "Scope10",
				File:  "temo.csv",
			},
			setup:   func(*v1.DataAnalysisRequest) {},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, qObj, nil)
			_, err := obj.DataAnalysis(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.NotifyUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("Test passed ", tt.name)
		})
	}
}
