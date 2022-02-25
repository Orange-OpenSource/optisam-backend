package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Command is a combination of config and command line
type Command struct {
	Conf      *viper.Viper
	Cmd       *cobra.Command
	EnvPrefix string
}
