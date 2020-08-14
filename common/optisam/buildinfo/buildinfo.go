// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package buildinfo

import (
	"runtime"
)

// BuildInfo represents all available build information.
type BuildInfo struct {
	Version    string `json:"version"`
	CommitHash string `json:"commit_hash"`
	BuildDate  string `json:"build_date"`
	GoVersion  string `json:"go_version"`
	Os         string `json:"os"`
	Arch       string `json:"arch"`
	Compiler   string `json:"compiler"`
}

// New returns all available build information.
func New(version string, commitHash string, buildDate string) BuildInfo {
	return BuildInfo{
		Version:    version,
		CommitHash: commitHash,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Os:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Compiler:   runtime.Compiler,
	}
}

// Fields returns the build information in a log context format.
func (bi BuildInfo) Fields() map[string]interface{} {
	return map[string]interface{}{
		"version":     bi.Version,
		"commit_hash": bi.CommitHash,
		"build_date":  bi.BuildDate,
		"go_version":  bi.GoVersion,
		"os":          bi.Os,
		"arch":        bi.Arch,
		"compiler":    bi.Compiler,
	}
}
