// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package equipmentstypes

import (
	"fmt"
	optisam_dg "optisam-backend/common/optisam/dgraph"
	"optisam-backend/license-service/pkg/repository/v1/dgraph"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/loader"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// CmdEquipmentsTypes informs about the command
	CmdEquipmentsTypes *config.Command
)

func init() {
	CmdEquipmentsTypes = &config.Command{
		Cmd: &cobra.Command{
			Use:   "equipmentstypes",
			Short: "load equipments types  in the dgraph not to be used in production",
			Long:  `load equipments types`,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("loading config " + CmdEquipmentsTypes.Conf.GetString("config"))
				fmt.Println("connecting alpha to " + strings.Join(CmdEquipmentsTypes.Conf.GetStringSlice("alpha"), ","))
				if err := loadEquipemntsTypes(); err != nil {
					return err
				}
				return nil
			},
		},
		EnvPrefix: "EQUIPMENTTYPES",
	}
}

func loadEquipemntsTypes() error {
	dgClient, err := optisam_dg.NewDgraphConnection(&optisam_dg.Config{
		Hosts: CmdEquipmentsTypes.Conf.GetStringSlice("alpha"),
	})
	if err != nil {
		return err
	}

	return loader.LoadDefaultEquipmentTypes(dgraph.NewLicenseRepository(dgClient))
}
