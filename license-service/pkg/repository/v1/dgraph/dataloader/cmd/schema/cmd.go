package schema

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph/loader"

	"github.com/spf13/cobra"
)

var (
	// CmdSchema informs about the command
	CmdSchema *config.Command
)

func init() {
	CmdSchema = &config.Command{
		Cmd: &cobra.Command{
			Use:   "schema",
			Short: "load schema in the dgraph",
			Long:  `load schema into dgraph`,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("loading config " + CmdSchema.Conf.GetString("config"))
				fmt.Println("loading schema from " + CmdSchema.Conf.GetString("schema_dir"))
				fmt.Println("connecting alpha to " + strings.Join(CmdSchema.Conf.GetStringSlice("alpha"), ","))
				if err := loadSchema(); err != nil {
					return err
				}
				return nil
			},
		},
		EnvPrefix: "SCHEMA",
	}
	CmdSchema.Cmd.Flags().StringP("schema_dir", "s", "schema", "directory where schema files are present")
}

func loadSchema() error {
	config := loader.NewDefaultConfig()
	config.CreateSchema = true
	config.Alpha = CmdSchema.Conf.GetStringSlice("alpha")
	files, err := getAllFilesWithSuffixFullPath(CmdSchema.Conf.GetString("schema_dir"), ".schema")
	if err != nil {
		return err
	}
	typeFiles, err := getAllFilesWithSuffixFullPath(CmdSchema.Conf.GetString("schema_dir"), ".types")
	if err != nil {
		return err
	}
	fmt.Println(files)
	config.SchemaFiles = files
	config.TypeFiles = typeFiles
	return loader.Load(config)
}

func getAllFilesWithSuffixFullPath(dir, suffix string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, f := range files {
		name := filepath.Base(f.Name())
		fmt.Println(name, f.Name())
		if !f.IsDir() && strings.HasSuffix(name, suffix) {
			fileNames = append(fileNames, dir+"/"+f.Name())
		}
	}
	return fileNames, nil
}
