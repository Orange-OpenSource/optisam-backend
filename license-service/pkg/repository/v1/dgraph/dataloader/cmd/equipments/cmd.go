package equipments

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph/loader"

	optisam_dg "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/dgraph"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/files"

	"github.com/spf13/cobra"
)

var (
	// CmdEquipments informs about the command
	CmdEquipments *config.Command
)

func init() {
	CmdEquipments = &config.Command{
		Cmd: &cobra.Command{
			Use:   "equipments",
			Short: "load equipments in the dgraph",
			Long:  `load all equipments to dgraph`,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("loading config " + CmdEquipments.Conf.GetString("config"))
				fmt.Println("loading metadata from " + CmdEquipments.Conf.GetString("skeleton_scope"))
				fmt.Println("loading static data from " + CmdEquipments.Conf.GetString("data_dir"))
				fmt.Println("connecting alpha to " + strings.Join(CmdEquipments.Conf.GetStringSlice("alpha"), ","))
				fmt.Println("connecting zero on " + CmdEquipments.Conf.GetString("zero"))
				fmt.Println("loading state from " + CmdEquipments.Conf.GetString("state_config"))
				if err := loadEquipemnts(); err != nil {
					return err
				}
				return nil
			},
		},
		EnvPrefix: "EQUIPMENT",
	}
	CmdEquipments.Cmd.Flags().StringP("skeleton_scope", "s", "skeletonscope", "directory where skeleton scope files are present")
	CmdEquipments.Cmd.Flags().StringP("data_dir", "e", "data_dir", "directory where data files are present")
}

func loadEquipemnts() error {
	config := loader.NewDefaultConfig()
	config.LoadEquipments = true
	config.Alpha = CmdEquipments.Conf.GetStringSlice("alpha")
	config.ScopeSkeleten = CmdEquipments.Conf.GetString("skeleton_scope")
	config.BatchSize = CmdEquipments.Conf.GetInt("batch_size")
	config.GenerateRDF = CmdEquipments.Conf.GetBool("gen_rdf")
	fls, err := getAllFilesWithPrefix(config.ScopeSkeleten, "equipment_")
	if err != nil {
		return err
	}
	config.EquipmentFiles = fls
	// year, month, day := time.Now().UTC().Add(-time.Hour * 24).Date()
	// date := fmt.Sprintf("%d_%s_%d", year, month.String(), day)
	destDir := CmdEquipments.Conf.GetString("data_dir")
	config.MasterDir = destDir
	scopes, err := files.GetAllTheDirectories(destDir)
	if err != nil {
		return err
	}
	config.Scopes = scopes
	config.StateConfig = CmdEquipments.Conf.GetString("state_config")
	dgClient, err := optisam_dg.NewDgraphConnection(&optisam_dg.Config{
		Hosts: config.Alpha,
	})
	if err != nil {
		return err
	}

	config.Repository = dgraph.NewLicenseRepository(dgClient)
	return loader.Load(config)
}

func getAllFilesWithPrefix(dir, prefix string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, f := range files {
		name := filepath.Base(f.Name())
		if !f.IsDir() && strings.HasPrefix(name, prefix) {
			fileNames = append(fileNames, name)
		}
	}
	return fileNames, nil
}
