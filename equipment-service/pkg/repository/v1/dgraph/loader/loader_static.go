// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

type staticNquadFunc func(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool)

func loadStaticTypes(object, xidColumn string, ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}, nquadFunc staticNquadFunc) {
	fmt.Println(masterDir, scopes)
	for _, scope := range scopes {
		scopeLoader := ml.GetLoader(filepath.Base(scope))
		for _, filename := range filenames {
			fl := scopeLoader.GetLoader(masterDir, filepath.Base(filename))
			func(fileLoader *FileLoader, f, s string) {
				load := func(version string) (time.Time, error) {
					log.Println("started loading static " + f)
					defer log.Println("end loading static " + f)
					return loadFile(fileLoader, ch, masterDir, s, version, f, xidColumn, doneChan, nquadFunc)
				}
				log.Println("set file loader static " + f)
				fileLoader.SetLoaderFunc(load)
			}(fl, filename, scope)
		}
	}
}
