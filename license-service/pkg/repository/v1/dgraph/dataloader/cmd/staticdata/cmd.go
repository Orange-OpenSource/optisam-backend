// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package staticdata

import (
	"fmt"
	"optisam-backend/common/optisam/files"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/loader"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// CmdStaticdata informs about the command
	CmdStaticdata *config.Command
)

func init() {
	CmdStaticdata = &config.Command{
		Cmd: &cobra.Command{
			Use:   "staticdata",
			Short: "load staticdata in the dgraph",
			Long:  `load all static data like products,applications,acqrights and instances`,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("loading config " + CmdStaticdata.Conf.GetString("config"))
				fmt.Println("loading destination dir from " + CmdStaticdata.Conf.GetString("data_dir"))
				fmt.Println("connecting alpha to " + strings.Join(CmdStaticdata.Conf.GetStringSlice("alpha"), ","))
				fmt.Println("connecting zero on " + CmdStaticdata.Conf.GetString("zero"))
				fmt.Println("loading state from " + CmdStaticdata.Conf.GetString("state_config"))
				if err := loadStaticData(); err != nil {
					return err
				}
				return nil
			},
		},
		EnvPrefix: "STATIC_DATA",
	}
	CmdStaticdata.Cmd.Flags().StringP("data_dir", "d", "datadir", "directory where static data files are present")
}

func loadStaticData() error {
	config := loader.NewDefaultConfig()
	config.LoadStaticData = true
	//year, month, day := time.Now().UTC().Add(-time.Hour * 24).Date()
	//date := fmt.Sprintf("%d_%s_%d", year, month.String(), day)
	config.Zero = CmdStaticdata.Conf.GetString("zero")
	config.BadgerDir = CmdStaticdata.Conf.GetString("badger_dir")
	config.Alpha = CmdStaticdata.Conf.GetStringSlice("alpha")
	config.BatchSize = CmdStaticdata.Conf.GetInt("batch_size")
	destDir := CmdStaticdata.Conf.GetString("data_dir")
	config.MasterDir = destDir
	//destDir := CmdStaticdata.Conf.GetString("data_dir") + "/" + date
	scopes, err := files.GetAllTheDirectories(destDir)
	if err != nil {
		return err
	}

	//TODO : consider scope based files in future versions
	config.Scopes = scopes
	config.StateConfig = CmdStaticdata.Conf.GetString("state_config")
	config.ProductFiles = []string{"prod.csv", "productsnew.csv", "products_equipments.csv"}
	config.AppFiles = []string{"applications.csv", "applications_products.csv"}
	config.InstFiles = []string{"applications_instances.csv", "instances_products.csv", "instances_equipments.csv"}
	config.AcqRightsFiles = []string{"products_acquiredRights.csv"}
	config.UsersFiles = []string{"products_equipments_users.csv"}
	return loader.Load(config)
}
