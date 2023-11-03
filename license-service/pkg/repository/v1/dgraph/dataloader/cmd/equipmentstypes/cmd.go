package equipmentstypes

import (
	"fmt"
	"strings"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph/loader"

	optisam_dg "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/dgraph"

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
