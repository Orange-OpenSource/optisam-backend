// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
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
	"time"

	dgo "github.com/dgraph-io/dgo/v2"
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
	if err = os.RemoveAll(badgerDir); err != nil {
		log.Println("Failed tp remove old badger dir...., err : ", err)
	}

	defer func() {
		cleanup(dockers)
	}()
	if cfg.Environment == "local" || cfg.Environment == "" {
		dockers, err = docker.Start(cfg.Dockers)
		if err != nil {
			log.Println("Failed to start containers, err: ", err)
			return
		}
		time.Sleep(10 * time.Second)

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
	path := "../../../../../license-service/pkg/repository/v1/dgraph/"
	config := loader.NewDefaultConfig()
	//hosts := strings.Split(cfg.Dgraph.Hosts[0], ":")
	//zero := fmt.Sprintf("%s:5080", hosts[0])
	//config.Zero = zero
	//config.Alpha = cfg.Dgraph.Hosts
	config.BatchSize = 1000
	config.CreateSchema = true
	config.SchemaFiles = []string{
		path + "schema/all/all.schema",
	}
	config.TypeFiles = []string{
		path + "schema/all/all.types",
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
	log.Printf("ddddd %+v", config)
	return loader.Load(config)
}

// func loadEquipments(badgerDir, masterDir string, scopes []string, filenames ...string) error {
// 	config := loader.NewDefaultConfig()
// 	//hosts := strings.Split(cfg.Dgraph.Hosts[0], ":")
// 	//zero := fmt.Sprintf("%s:5080", hosts[0])
// 	//config.Zero = zero
// 	//config.Alpha = cfg.Dgraph.Hosts
// 	config.MasterDir = masterDir
// 	config.EquipmentFiles = filenames
// 	config.Scopes = scopes
// 	config.LoadEquipments = true
// 	config.IgnoreNew = true
// 	dg, err := dgraph.NewDgraphConnection(&dgraph.Config{
// 		Hosts: config.Alpha,
// 	})
// 	if err != nil {
// 		log.Println("Failed to get dgclient err", err)
// 		return err
// 	}
// 	config.Repository = NewMetricRepository(dg)

// 	return loader.Load(config)
// }

// func dropall() {
// 	_ = dgClient.Alter(context.Background(), &api.Operation{DropAll: true})
// }
