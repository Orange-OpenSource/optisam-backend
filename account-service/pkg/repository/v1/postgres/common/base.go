package common

import (
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/docker"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/postgres"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var sqldb *sql.DB
var cfg config.Config

func addConfig(path string) {
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
	viper.AddConfigPath(path)
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
	if err = cfg.Postgres.Validate(); err != nil {
		log.Fatalf("failed to read config, missing :- %v ", err)
	}
	cfg.Environment = env

}

func Testdata(path string, files []string) (*sql.DB, []*docker.DockerInfo, error) {
	if err := logger.Init(-1, ""); err != nil {
		panic(err)
	}
	addConfig(path)
	var err error
	var dockers []*docker.DockerInfo
	if cfg.Environment == "local" || cfg.Environment == "" {
		dockers, err = docker.Start(cfg.Dockers)
		if err != nil {
			return nil, nil, err
		}
		time.Sleep(cfg.INITWAITTIME * time.Second)
	}
	pgDB, err := postgres.NewConnection(*cfg.Postgres)
	if err != nil {
		logger.Log.Error("Failed to connect postgres", zap.Error(err))
		return nil, nil, err
	}
	if err := pgDB.Ping(); err != nil {
		logger.Log.Error("Failed to ping postgres", zap.Error(err))
		return nil, nil, err
	}
	sqldb = pgDB
	if err := loadData(files); err != nil {
		logger.Log.Error("Failed to load data into postgres", zap.Error(err))
		return nil, nil, err
	}
	return sqldb, dockers, nil
}

func loadData(files []string) error {
	for _, file := range files {
		query, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := sqldb.Exec(string(query)); err != nil {
			return err
		}

	}
	return nil
}
