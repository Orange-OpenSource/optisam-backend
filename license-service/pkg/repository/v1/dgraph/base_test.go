// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"flag"
	"log"
	"optisam-backend/common/optisam/config"
	"optisam-backend/common/optisam/dgraph"
	"optisam-backend/common/optisam/docker"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/loader"
	"os"
	"strings"
	"testing"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var dgClient *dgo.Dgraph
var cfg *config.Config

const (
	badgerDir string = "badger"
)

func init() {
	fileName := ""
	configFileLocation := ""
	viper := viper.GetViper()
	flag.StringVar(&configFileLocation, "configFile", "", "cmd based mannaully config file passing")

	env := strings.ToLower(os.Getenv("ENV"))
	switch env {
	case "dev":
		fileName = "config-test-dev"
	case "int":
		fileName = "config-test-int"
	case "pprod":
		fileName = "config-test-pprod"
	case "prod":
		fileName = "config-test-prod"
	default:
		fileName = "config-test-local"
		env = "local"
	}
	log.Println(" Tests will be run on environment [", env, "] with [", fileName, "]")
	viper.SetConfigName(fileName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("/opt/config/")
	viper.AddConfigPath("../../../../cmd/server/")
	viper.AddConfigPath("../../../../../cmd/server/")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configFileLocation)
	viper.SetDefault("INITWAITTIME", 5)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("failed to unmarshal configuration: %v", err)
	}
	cfg.Environment = env

}

func cleanup(val []*docker.DockerInfo) {
	docker.Stop(val)
	if err := os.RemoveAll(badgerDir); err != nil {
		log.Println("Failed tp remove old badger dir...., err : ", err)
	}

}

func TestMain(m *testing.M) {
	var dockers []*docker.DockerInfo
	logger.Init(-1, "")
	var err error

	defer cleanup(dockers)
	if cfg.Environment == "local" || cfg.Environment == "" {
		// dockers, err = docker.Start(cfg.Dockers)
		// if err != nil {
		// 	log.Println("Failed to start containers, err: ", err)
		// 	return
		// }
		// time.Sleep(cfg.INITWAITTIME * time.Second)
	}
	conn, err := dgraph.NewDgraphConnection(cfg.Dgraph)
	if err != nil {
		logger.Log.Error("test main cannot connect to alpha", zap.String("reason", err.Error()))
		return
	}

	dgClient = conn

	if err := loadDgraphData(badgerDir); err != nil {
		logger.Log.Error("test main cannot load data", zap.String("reason", err.Error()))
		return
	}

	log.Println("LOADED ...")
	code := m.Run()
	cleanup(dockers)
	os.Exit(code)

}

func loadDgraphData(badgerDir string) error {
	if err := dgClient.Alter(context.Background(), &api.Operation{DropAll: true}); err != nil {
		return err
	}
	config := loader.NewDefaultConfig()
	config.BatchSize = 1000
	config.CreateSchema = true
	config.LoadMetadata = true
	config.LoadStaticData = true
	config.SchemaFiles = []string{
		"schema/application.schema",
		"schema/products.schema",
		"schema/instance.schema",
		"schema/equipment.schema",
		"schema/metadata.schema",
		"schema/acq_rights.schema",
		"schema/metric_ops.schema",
		"schema/metric_ips.schema",
		"schema/metric_sps.schema",
		"schema/metric_oracle_nup.schema",
		"schema/metric_acs.schema",
		"schema/editor.schema",
		"schema/products_aggregations.schema",
		"schema/users.schema",
	}
	config.TypeFiles = []string{
		"schema/application.types",
		"schema/products.types",
		"schema/instance.types",
		"schema/equipment.types",
		"schema/metadata.types",
		"schema/acq_rights.types",
		"schema/metric_ops.types",
		"schema/metric_ips.types",
		"schema/metric_sps.types",
		"schema/metric_oracle_nup.types",
		"schema/metric_acs.types",
		"schema/editor.types",
		"schema/products_aggregations.types",
		"schema/users.types",
	}
	config.ScopeSkeleten = "skeletonscope"
	config.MasterDir = "testdata"
	config.Scopes = []string{
		// TODO: ADD scopes directories here like
		// EX:
		"scope1",
		"scope2",
		"scope3",
		"scope4",
	}
	config.ProductFiles = []string{
		"prod.csv",
		"productsnew.csv",
	}
	config.ProductEquipmentFiles = []string{
		"products_equipments.csv",
		"products_equipments_users.csv",
	}
	config.AppFiles = []string{
		"applications.csv",
	}
	config.AppProdFiles = []string{
		"applications_products.csv",
	}
	config.InstFiles = []string{
		"applications_instances.csv",
	}
	config.InstProdFiles = []string{
		"instances_products.csv",
	}
	config.InstEquipFiles = []string{
		"instances_equipments.csv",
	}
	config.AcqRightsFiles = []string{
		"products_acquiredRights.csv",
	}
	config.UsersFiles = []string{
		"products_equipments_users.csv",
	}
	return loader.Load(config)
}

func loadEquipments(badgerDir, masterDir string, scopes []string, filenames ...string) error {
	config := loader.NewDefaultConfig()
	config.MasterDir = masterDir
	config.EquipmentFiles = filenames
	config.Scopes = scopes
	config.LoadEquipments = true
	config.IgnoreNew = true
	config.Repository = NewLicenseRepository(dgClient)
	return loader.Load(config)
}

func dropall() {
	_ = dgClient.Alter(context.Background(), &api.Operation{DropAll: true})
}
