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

var cfg *config.Config
var db *sql.DB

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
	log.Printf("Test config %+v", cfg)
}

func cleanup(val []*docker.DockerInfo) {
	docker.Stop(val)
}
func TestMain(m *testing.M) {
	var dockers []*docker.DockerInfo
	logger.Init(-1, "")
	var err error

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
	db, err = postgres.NewConnection(*cfg.Postgres)
	if err != nil {
		logger.Log.Error("failed to open connection with postgres: %v", zap.Error(err))
		return
	}

	// Verify connection.
	if err = db.Ping(); err != nil {
		logger.Log.Error("failed to verify connection to PostgreSQL: %v", zap.Error(err))
		db.Close()
		return
	}
	logger.Log.Info(" DB connection established for testing at  ", zap.String(" host ", cfg.Postgres.Host))
	defer db.Close()

	// Run Migration
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("migrations", "schema"),
	}
	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	log.Printf("Applied %d migrations!\n", n)

	log.Println("Prerequisites for testing has been initiated , starting unit testing.....")
	code := m.Run()

	os.Exit(code)
}
