package postgres

import (
	"database/sql"
	"flag"
	"log"
	"optisam-backend/common/optisam/config"
	"optisam-backend/common/optisam/docker"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/postgres"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/packr/v2"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// nolint: gochecknoglobals
var sqldb *sql.DB
var cfg *config.Config

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
}

func TestMain(m *testing.M) {
	var dockers []*docker.DockerInfo
	logger.Init(-1, "")
	var err error
	defer cleanup(dockers)
	if cfg.Environment == "local" || cfg.Environment == "" {
		dockers, err = docker.Start(cfg.Dockers)
		if err != nil {
			log.Println("Failed to start containers, err: ", err)
			return
		}
		time.Sleep(cfg.INITWAITTIME * time.Second)
	}

	sqldb, err = postgres.NewConnection(*cfg.Postgres)
	if err != nil {
		logger.Log.DPanic("Cannot connect to Postgres DB")
	}

	// Verify connection.
	if err = sqldb.Ping(); err != nil {
		logger.Log.Error("failed to verify connection to PostgreSQL: %v", zap.Error(err))
		sqldb.Close()
		return
	}
	logger.Log.Info(" DB connection established for testing at  ", zap.String(" host ", cfg.Postgres.Host))
	defer sqldb.Close()

	// Run Migration
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("migrations", "schema"),
	}
	n, err := migrate.Exec(sqldb, "postgres", migrations, migrate.Up)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	log.Printf("Applied %d migrations!\n", n)
	code := m.Run()
	cleanup(dockers)
	os.Exit(code)
}
