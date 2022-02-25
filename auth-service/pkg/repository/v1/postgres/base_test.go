package postgres

import (
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"optisam-backend/common/optisam/config"
	"optisam-backend/common/optisam/docker"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/postgres"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

var db *sql.DB
var cfg config.Config

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

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	var err error
	var dockers []*docker.DockerInfo

	if cfg.Environment == "local" || cfg.Environment == "" {
		dockers, err = docker.Start(cfg.Dockers)
		if err != nil {
			panic(err)
		}
		time.Sleep(cfg.INITWAITTIME * time.Second)
	}
	defer docker.Stop(dockers)

	pgDB, err := postgres.NewConnection(*cfg.Postgres)
	if err != nil {
		panic(err)
	}

	if err := pgDB.Ping(); err != nil {
		panic(err)
	}

	db = pgDB

	if err := loadData(); err != nil {
		panic(err)
	}
	code := m.Run()
	docker.Stop(dockers)
	os.Exit(code)
}

func loadData() error {
	files := []string{"scripts/1_user_login.sql"}
	for _, file := range files {
		query, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(query)); err != nil {
			return err
		}

	}
	return nil
}
