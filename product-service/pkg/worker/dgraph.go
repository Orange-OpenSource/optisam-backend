// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"strconv"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type worker struct {
	id string
	dg *dgo.Dgraph
}

type MessageType string

const (
	UpsertProductRequest MessageType = "UpsertProduct"
)

type Envelope struct {
	Type MessageType `json:"message_type"`
	Json json.RawMessage
}

func NewWorker(id string, dg *dgo.Dgraph) *worker {
	return &worker{id: id, dg: dg}
}

func (w *worker) ID() string {
	return w.id
}

//Dowork will load products/linked applications,linked equipments data into Dgraph
func (w *worker) DoWork(ctx context.Context, j *job.Job) error {
	var e Envelope
	var updatePartialFlag bool
	_ = json.Unmarshal(j.Data, &e)
	logger.Log.Info("Operation Type", zap.String("message_type", string(e.Type)))
	switch e.Type {
	case UpsertProductRequest:
		logger.Log.Info("Processing UpsertProductRequest")
		var upr v1.UpsertProductRequest
		_ = json.Unmarshal(e.Json, &upr)
		var addProductApplication, addProductEquipment string
		var muUpsertProduct, muAddProductEquipment, muAddProductApplication *api.Mutation
		query := `query {
			var(func: eq(product.swidtag,"` + upr.GetSwidTag() + `")) @filter(eq(type_name,"product")){
				product as uid
			}
			`

		addProduct := `
			uid(product) <product.swidtag> "` + upr.GetSwidTag() + `" .
			uid(product) <type_name> "product" .
			uid(product) <dgraph.type> "Product" .
			uid(product) <scopes> "` + upr.GetScope() + `" .
			`
		if upr.GetOptionOf() != "" {
			query += `
				var(func: eq(product.swidtag,"` + upr.GetOptionOf() + `")) @filter(eq(type_name,"product")){
					child as uid
				}`
			addProduct += `
				uid(child) <product.child> uid(product) .
				`
		}
		// Application Upsert
		if len(upr.GetApplications().GetApplicationId()) > 0 {
			updatePartialFlag = true
			addProductApplication = `
			uid(product) <product.swidtag> "` + upr.GetSwidTag() + `" .
			uid(product) <type_name> "product" .
			uid(product) <dgraph.type> "Product" .
			`
			if upr.GetApplications().GetOperation() == "add" {

				for i, app := range upr.GetApplications().GetApplicationId() {
					query += `
					var(func: eq(application.id,"` + app + `")) @filter(eq(type_name,"application")){
						application` + strconv.Itoa(i) + ` as uid
					}
					`
					addProductApplication += `
					uid(application` + strconv.Itoa(i) + `) <application.id> "` + app + `" .
					uid(application` + strconv.Itoa(i) + `) <type_name> "application" .
					uid(application` + strconv.Itoa(i) + `) <dgraph.type> "Application" .
					uid(application` + strconv.Itoa(i) + `) <application.product> uid(product) .
					`
				}

			}
		}
		// Equipments Upsert
		if len(upr.GetEquipments().GetEquipmentusers()) > 0 {
			updatePartialFlag = true
			addProductEquipment = `
			uid(product) <product.swidtag> "` + upr.GetSwidTag() + `" .
			uid(product) <type_name> "product" .
			uid(product) <dgraph.type> "Product" .
			`
			if upr.GetEquipments().GetOperation() == "add" {
				for i, equipUser := range upr.GetEquipments().GetEquipmentusers() {
					query += `
					var(func: eq(equipment.id,"` + equipUser.GetEquipmentId() + `")) @filter(eq(type_name,"equipment")){
						equipment` + strconv.Itoa(i) + ` as uid
					}
					`
					addProductEquipment += `
					uid(equipment` + strconv.Itoa(i) + `) <equipment.id> "` + equipUser.GetEquipmentId() + `" .
					uid(equipment` + strconv.Itoa(i) + `) <type_name> "equipment" .
					uid(equipment` + strconv.Itoa(i) + `) <dgraph.type> "Equipment" .
					uid(product) <product.equipment> uid(equipment` + strconv.Itoa(i) + `) .
					`
					if equipUser.GetNumUser() > 0 {
						query += `
						var(func: eq(users.id,user_` + upr.GetSwidTag() + equipUser.GetEquipmentId() + `)) @filter(eq(type_name,"users")){
							users` + strconv.Itoa(i) + ` as uid
						}
						`
						addProductEquipment += `
						uid(users` + strconv.Itoa(i) + `) <users.id> "user_` + upr.GetSwidTag() + equipUser.GetEquipmentId() + `" .
						uid(users` + strconv.Itoa(i) + `) <type_name> "instance_users" .
						uid(users` + strconv.Itoa(i) + `) <dgraph.type> "User" .
						uid(product) <product.users> uid(users` + strconv.Itoa(i) + `) .
						uid(equipment` + strconv.Itoa(i) + `) <equipment.users>  uid(users` + strconv.Itoa(i) + `) .
						uid(users` + strconv.Itoa(i) + `) <users.count> "` + strconv.Itoa(int(equipUser.GetNumUser())) + `" .
						`
					}
				}

			}
		}

		if !updatePartialFlag {
			query += `
			var(func: eq(editor.name,"` + upr.GetEditor() + `")) @filter(eq(type_name,"editor")){
				editor as uid
			}
			`
			addProduct += `
			uid(product) <product.name> "` + upr.GetName() + `" .
			uid(product) <product.version> "` + upr.GetVersion() + `" .
			uid(product) <product.category> "` + upr.GetCategory() + `" .
			uid(product) <product.editor> "` + upr.GetEditor() + `" .
			uid(editor) <editor.product> uid(product) .
			uid(editor) <type_name> "editor" .
			uid(editor) <dgraph.type> "Editor" .
			uid(editor) <editor.name> "` + upr.GetEditor() + `" .
			`
		}
		query += `
		}`
		logger.Log.Info("", zap.String("query", query))
		logger.Log.Info("", zap.String("muUpsertProduct", addProduct))
		logger.Log.Info("", zap.String("muAddProductApplication", addProductApplication))
		logger.Log.Info("", zap.String("muAddProductEquipment", addProductEquipment))
		muUpsertProduct = &api.Mutation{SetNquads: []byte(addProduct)}
		muAddProductApplication = &api.Mutation{SetNquads: []byte(addProductApplication)}
		muAddProductEquipment = &api.Mutation{SetNquads: []byte(addProductEquipment)}
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muUpsertProduct, muAddProductApplication, muAddProductEquipment},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	default:
		fmt.Println(e.Json)
	}

	//Everything's fine, we're done here
	return nil
}
