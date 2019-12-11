// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package metadata

import (
	"fmt"
	"io/ioutil"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/loader"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// CmdMetadata informs about the command
	CmdMetadata *config.Command
)

func init() {
	CmdMetadata = &config.Command{
		Cmd: &cobra.Command{
			Use:   "metadata",
			Short: "load metadata in the dgraph",
			Long:  `load metadata in the dgraph all the csv files must be present in dgraph`,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("loading config " + CmdMetadata.Conf.GetString("config"))
				fmt.Println("loading metadata from " + CmdMetadata.Conf.GetString("skeleton_scope"))
				fmt.Println("connecting alpha to " + strings.Join(CmdMetadata.Conf.GetStringSlice("alpha"), ","))
				if err := loadMetadata(); err != nil {
					return err
				}
				return nil
			},
		},
		EnvPrefix: "METADATA",
	}
	CmdMetadata.Cmd.Flags().StringP("skeleton_scope", "m", "skeletonscope", "directory where skeleton scope files are present")
}

func loadMetadata() error {
	config := loader.NewDefaultConfig()
	config.LoadMetadata = true
	config.Alpha = CmdMetadata.Conf.GetStringSlice("alpha")
	config.ScopeSkeleten = CmdMetadata.Conf.GetString("skeleton_scope")
	config.BatchSize = CmdMetadata.Conf.GetInt("batch_size")
	files, err := getAllFilesWithPrefix(config.ScopeSkeleten, "equipment_")
	if err != nil {
		return err
	}
	config.MetadataFiles = &loader.MetadataFiles{
		EquipFiles: files,
	}
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
