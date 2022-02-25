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
