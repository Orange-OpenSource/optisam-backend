// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package logger

// Config holds details necessary for logging.
type Config struct {

	// Level is the minimum log level that should appear on the output.
	LogLevel int

	// LogTimeFormat Time Format.
	LogTimeFormat string
}
