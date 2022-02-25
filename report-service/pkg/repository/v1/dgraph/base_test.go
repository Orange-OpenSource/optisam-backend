package dgraph

import (
	"context"
	"flag"
	"log"
	"optisam-backend/common/optisam/config"
	"optisam-backend/common/optisam/dgraph"
	"optisam-backend/common/optisam/docker"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1/dgraph"
	"optisam-backend/equipment-service/pkg/repository/v1/dgraph/loader"
	"os"
	"strings"
	"testing"
	"time"

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
	if err = os.RemoveAll(badgerDir); err != nil {
		log.Println("Failed tp remove old badger dir...., err : ", err)
	}

	defer cleanup(dockers)
	if cfg.Environment == "local" || cfg.Environment == "" {
		log.Println("Starting container")
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
	config := loader.NewDefaultConfig()
	// hosts := strings.Split(cfg.Dgraph.Hosts[0], ":")
	// zero := fmt.Sprintf("%s:5080", hosts[0])
	// config.Zero = zero
	// config.Alpha = cfg.Dgraph.Hosts
	config.BatchSize = 1000
	config.CreateSchema = true
	config.LoadMetadata = false
	config.LoadStaticData = false
	config.SchemaFiles = []string{
		"../../../../../license-service/pkg/repository/v1/dgraph/schema/all/all.schema",
		// "schema/application.schema",
		// "schema/products.schema",
		// "schema/instance.schema",
		// "schema/equipment.schema",
		// "schema/metadata.schema",
		// "schema/acq_rights.schema",
		// "schema/metric_ops.schema",
		// "schema/metric_ips.schema",
		// "schema/metric_sps.schema",
		// "schema/metric_oracle_nup.schema",
		// "schema/editor.schema",
		// "schema/products_aggregations.schema",
	}
	config.TypeFiles = []string{
		"../../../../../license-service/pkg/repository/v1/dgraph/schema/all/all.types",
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
	// config.ProductFiles = []string{
	// 	"prod.csv",
	// 	"productsnew.csv",
	// }
	// config.ProductEquipmentFiles = []string{
	// 	"products_equipments.csv",
	// }
	// config.AppFiles = []string{
	// 	"applications.csv",
	// }
	// config.AppProdFiles = []string{
	// 	"applications_products.csv",
	// }
	// config.InstFiles = []string{
	// 	"applications_instances.csv",
	// }
	// config.InstProdFiles = []string{
	// 	"instances_products.csv",
	// }
	// config.InstEquipFiles = []string{
	// 	"instances_equipments.csv",
	// }
	// config.AcqRightsFiles = []string{
	// 	"products_acquiredRights.csv",
	// }
	// config.UsersFiles = []string{
	// 	"products_equipments_users.csv",
	// }
	return loader.Load(config)
}

func loadEquipments(badgerDir, masterDir string, scopes []string, filenames ...string) error {
	config := loader.NewDefaultConfig()
	// hosts := strings.Split(cfg.Dgraph.Hosts[0], ":")
	// zero := fmt.Sprintf("%s:5080", hosts[0])
	// config.Zero = zero
	// config.Alpha = cfg.Dgraph.Hosts
	config.MasterDir = masterDir
	config.EquipmentFiles = filenames
	config.Scopes = scopes
	config.LoadEquipments = true
	config.IgnoreNew = true
	dg, err := dgraph.NewDgraphConnection(&dgraph.Config{
		Hosts: config.Alpha,
	})
	if err != nil {
		log.Println("Failed to get dgclient err", err)
		return err
	}
	config.Repository = v1.NewEquipmentRepository(dg)

	return loader.Load(config)
}

func dropall() {
	_ = dgClient.Alter(context.Background(), &api.Operation{DropAll: true})
}
