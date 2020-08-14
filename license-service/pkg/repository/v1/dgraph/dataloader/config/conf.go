// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

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
