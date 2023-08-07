package postgres

import (
	"database/sql"
	"flag"
	"log"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/postgres"
	"optisam-backend/product-service/pkg/config"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfg *config.Config
var db *sql.DB

func createDBConnection() error {
	// Create database connection.
	var err error
	db, err = postgres.NewConnection(postgres.Config{
		Host: cfg.Database.Host,
		Port: cfg.Database.Port,
		Name: cfg.Database.User.Name,
		User: cfg.Database.User.User,
		Pass: cfg.Database.User.Pass})
	if err != nil {
		logger.Log.Error("failed to open connection with postgres: %v", zap.Error(err))
		return err
	}

	// Verify connection.
	if err = db.Ping(); err != nil {
		logger.Log.Error("failed to verify connection to PostgreSQL: %v", zap.Error(err))
		db.Close()
		return err
	}
	logger.Log.Info(" DB connection established for testing at  ", zap.String(" host ", cfg.Database.Host))
	return nil
}
func readConfig() {
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
	log.Printf("CFG %+v", cfg)
}
func TestMain(m *testing.M) {
	readConfig()
	logger.Init(-1, "")
	if createDBConnection() != nil {
		log.Println("Failed to start DB with ", cfg.Database)
		return
	}
	defer db.Close()
	log.Println("Prerequisites for testing has been initiated , starting unit testing.....")
	code := m.Run()

	os.Exit(code)
}
