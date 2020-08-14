// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rpc

import (
	"context"
	"encoding/json"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	product "optisam-backend/product-service/pkg/api/v1"
	"google.golang.org/grpc"
	"go.uber.org/zap"
)


type Envelope struct {

	Id int32 `json:"id"`
	Name string `json:"Name`
	Swidtags []string `json:"swidtags`
	ActionType string `json:"actionType`
}

type Worker struct {
	id string
	grpcServers map[string]*grpc.ClientConn
}

//NewWorker give worker object
func NewWorker(id string, conn map[string]*grpc.ClientConn) *Worker {
	return &Worker{id: id, grpcServers: conn}
}

func (w *Worker) ID() string {
	return w.id
}

//Dowork will load products/linked applications,linked equipments data into Dgraph
func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	var e Envelope
	_ = json.Unmarshal(j.Data, &e)
	appData := product.UpsertAggregationRequest{
		AggregationId : e.Id,
		AggregationName : e.Name,
		Swidtags : e.Swidtags,
		ActionType : e.ActionType}

	resp, err := product.NewProductServiceClient(w.grpcServers["product"]).UpsertProductAggregation(ctx, &appData)
	if err != nil {
		logger.Log.Error("upsert Product aggregation failed", zap.String("error", string(err.Error())))
		return err	
	}
	if resp.Success == false {
		logger.Log.Error("upsert Product aggregation failed from rpc", zap.String("err", string(err.Error())))
		return err		
	}
	return nil
	}
	
