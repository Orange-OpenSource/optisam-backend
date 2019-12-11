// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package cmd

import (
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/cmd/addcolumn"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/cmd/equipments"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/cmd/equipmentstypes"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/cmd/metadata"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/cmd/schema"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/cmd/staticdata"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cmdRoot = &cobra.Command{
		Use:   "dataloader",
		Short: "p",
		Long:  "dataloader provides commands to load csv data in dgraph for optisam",
		Args:  cobra.NoArgs,
	}
)

var (
	alphas []string
	zero   string
)

var (
	subCommands = []*config.Command{
		schema.CmdSchema,
		metadata.CmdMetadata,
		staticdata.CmdStaticdata,
		equipments.CmdEquipments,
		equipmentstypes.CmdEquipmentsTypes,
		addcolumn.CmdAddColumn,
	}
)

var rootConf = viper.New()

func initCmds() {
	cmdRoot.PersistentFlags().StringSliceP("alpha", "a", []string{"localhost:9080"}, "dataloader schema --alpha localhost:9080 --alpha localhost:9081")
	cmdRoot.PersistentFlags().Int32("batch_size", 1000, "dataloader staticdata --batch_size 1000")
	cmdRoot.PersistentFlags().StringP("zero", "z", "localhost:5080", "dataloader metadata --zero localhost:5080")
	cmdRoot.PersistentFlags().StringP("state_config", "c", "state.json", "dataloader staticdata --alpha localhost:5080 -sc state.json")
	cmdRoot.PersistentFlags().StringP("badger_dir", "b", "badger", "dataloader staticdata --alpha localhost:5080 -sc state.json -b badger")
	cmdRoot.PersistentFlags().String("config", "",
		"Configuration file. Takes precedence over default values, but is "+
			"overridden to values set with environment variables and flags.")
	rootConf.BindPFlags(cmdRoot.Flags())
	rootConf.BindPFlags(cmdRoot.PersistentFlags())
	for _, sc := range subCommands {
		cmdRoot.AddCommand(sc.Cmd)
		sc.Conf = viper.New()
		sc.Conf.BindPFlags(sc.Cmd.Flags())
		sc.Conf.BindPFlags(cmdRoot.PersistentFlags())
		sc.Conf.AutomaticEnv()
		sc.Conf.SetEnvPrefix(sc.EnvPrefix)
		sc.Conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}
	cobra.OnInitialize(func() {
		cfg := rootConf.GetString("config")
		if cfg == "" {
			return
		}
		for _, sc := range subCommands {
			sc.Conf.SetConfigFile(cfg)
			if err := sc.Conf.ReadInConfig(); err != nil {
				panic(err)
			}
		}
	})
}

// Execute ...
func Execute() error {
	initCmds()
	return cmdRoot.Execute()
}
