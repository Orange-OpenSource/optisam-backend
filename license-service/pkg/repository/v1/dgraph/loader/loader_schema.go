// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package loader

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"optisam-backend/common/optisam/logger"
	"os"
	"strings"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"go.uber.org/zap"
)

func dropSchema(dg *dgo.Dgraph) error {
	fmt.Println("started schema drop fmt")
	log.Println("started schema drop")
	if err := dg.Alter(context.Background(), &api.Operation{
		DropAll: true,
	}); err != nil {
		logger.Log.Error("cannot drop schema", zap.String("reason", err.Error()))
		return err
	}
	log.Println("completed schema drop")
	return nil
}

func createSchema(dg *dgo.Dgraph, files []string) error {
	log.Println("started schema creation")
	schema := ""
	for i := range files {
		f, err := os.Open(files[i])
		if err != nil {
			logger.Log.Error("drop scema cannot open file", zap.String("filename", files[i]), zap.String("reason", err.Error()))
			return err
		}
		sch, err := ioutil.ReadAll(f)
		if err != nil {
			logger.Log.Error("drop scema cannot read file", zap.String("filename", files[i]), zap.String("reason", err.Error()))
			return err
		}
		schema += string(sch) + "\n"
	}

	fmt.Println(schema)

	if err := dg.Alter(context.Background(), &api.Operation{
		Schema: strings.TrimSpace(schema),
	}); err != nil {
		logger.Log.Error("cannot create schema", zap.String("reason", err.Error()))
		return err
	}
	log.Println("completed schema creation")
	return nil
}
