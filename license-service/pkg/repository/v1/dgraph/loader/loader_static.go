// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package loader

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
)

type staticNquadFunc func(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, string)

func loadStaticTypes(object, xidColumn string, ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Mutation, doneChan <-chan struct{}, nquadFunc staticNquadFunc) {
	wg := new(sync.WaitGroup)
	fmt.Println(masterDir, scopes)
	for _, scope := range scopes {
		scopeLoader := ml.GetLoader(filepath.Base(scope))
		for _, filename := range filenames {
			fl := scopeLoader.GetLoader(masterDir, filepath.Base(filename))
			func(fileLoader *FileLoader, f, s string) {
				load := func(version string) (time.Time, error) {
					return loadFile(fileLoader, ch, masterDir, s, version, f, xidColumn, doneChan, nquadFunc)
				}
				fileLoader.SetLoaderFunc(load)
			}(fl, filename, scope)
		}

		wg.Add(1)
		go func(sl *ScopeLoader) {
			sl.Load(masterDir)
			wg.Done()
		}(scopeLoader)
	}
	wg.Wait()
}
