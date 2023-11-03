package loader

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
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

func createSchema(dg *dgo.Dgraph, files, typeFiles []string) error {
	log.Println("started schema creation")
	schema, err := readFiles(files, "\n")
	if err != nil {
		return err
	}

	// fmt.Println(schema)

	if err := alterSchema(dg, schema); err != nil {
		return err
	}

	types, err := readFiles(typeFiles, "\n")
	if err != nil {
		return err
	}

	// fmt.Println(types)

	if err := alterSchema(dg, types); err != nil {
		return err
	}

	// fmt.Println(schema)

	log.Println("completed schema creation")
	return nil
}

func alterSchema(dg *dgo.Dgraph, schema string) error {
	// fmt.Println(schema)
	if err := dg.Alter(context.Background(), &api.Operation{
		Schema: strings.TrimSpace(schema),
	}); err != nil {
		fmt.Println(schema)
		logger.Log.Error("cannot create schema", zap.String("reasons", err.Error()))
		return err
	}
	log.Println("completed schema creation")
	return nil
}

func readFiles(files []string, delim string) (string, error) {
	var schema string
	for i := range files {
		f, err := os.Open(files[i])
		if err != nil {
			logger.Log.Error("drop scema cannot open file", zap.String("filename", files[i]), zap.String("reason", err.Error()))
			return "", err
		}
		sch, err := ioutil.ReadAll(f)
		if err != nil {
			logger.Log.Error("drop scema cannot read file", zap.String("filename", files[i]), zap.String("reason", err.Error()))
			return "", err
		}
		schema += string(sch) + "\n"
	}
	return schema, nil
}
