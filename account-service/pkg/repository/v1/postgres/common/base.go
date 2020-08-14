// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package common

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
	"time"

	"github.com/spf13/viper"
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
	logger.Init(-1, "")
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
		return nil, nil, err
	}
	if err := pgDB.Ping(); err != nil {
		return nil, nil, err
	}
	sqldb = pgDB
	if err := loadData(files); err != nil {
		return nil, nil, err
	}
	return sqldb, dockers, nil
}

func loadData(files []string) error {
	//files := []string{"../scripts/1_user_login.sql", "../schema/2_add_users_audit_table.sql"}
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
