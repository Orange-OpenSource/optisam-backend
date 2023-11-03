package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/files"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

// State tells about the state of loader
type State uint8

const (
	// LoaderStateCreated means that this loader was not presnt
	LoaderStateCreated State = 0
	// LoaderStateUpdated means that all the records of this loader are successfully imported.
	LoaderStateUpdated State = 1
	// LoaderStateFailed means there was some problem in uploading loaders records
	LoaderStateFailed State = 2
)

// Status tells the state of a loader
type Status struct {
	State State
}

// MasterLoader is a collection of scope loaders
type MasterLoader struct {
	Loaders map[string]*ScopeLoader
	lock    sync.Mutex
}

// Errors return all the errors faces by loaders
func (ml *MasterLoader) Error() error {
	ml.lock.Lock()
	defer ml.lock.Unlock()
	var errs []string
	for scope, ldr := range ml.Loaders {
		if ldr.Error() != nil {
			errs = append(errs, scope+":"+ldr.Error().Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, ",\n"))
}

func (ml *MasterLoader) Load(masterDir string) {
	ml.lock.Lock()
	defer ml.lock.Unlock()
	wg := new(sync.WaitGroup)
	for _, sl := range ml.Loaders {
		wg.Add(1)
		go func(s *ScopeLoader) {
			s.Load(masterDir)
			wg.Done()
		}(sl)
	}
	wg.Wait()
}

// GetLoader gets a loader if it does not exists it creates one
func (ml *MasterLoader) GetLoader(scope string) *ScopeLoader {
	fmt.Println(scope)
	ml.lock.Lock()
	defer ml.lock.Unlock()
	if ml.Loaders == nil {
		ml.Loaders = make(map[string]*ScopeLoader)
	}
	scopeLoader, ok := ml.Loaders[scope]
	if !ok {
		scopeLoader = &ScopeLoader{
			Scope:   scope,
			State:   LoaderStateCreated,
			Loaders: make(map[string]*FileLoader),
		}
		ml.Loaders[scope] = scopeLoader
	}
	return scopeLoader
}

func newMasterLoaderFromFile(file string) (*MasterLoader, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	ml := &MasterLoader{}
	if err := json.Unmarshal(data, ml); err != nil {
		return nil, err
	}
	return ml, nil
}

func saveMaterLoaderTofile(file string, ml *MasterLoader) error {
	data, err := json.Marshal(ml)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, data, 0644)
}

// ScopeLoader is a collection of file Loaders
type ScopeLoader struct {
	Scope   string
	State   State
	Loaders map[string]*FileLoader
	lock    sync.Mutex
	errors  []error
}

// Load ...
func (s *ScopeLoader) Load(masterDir string) error {
	versionDirs, err := files.GetAllTheDirectories(filepath.Join(masterDir, s.Scope))
	if err != nil {
		return err
	}
	sort.SliceStable(versionDirs, func(i, j int) bool {
		s1, err1 := strconv.Atoi(strings.TrimPrefix(versionDirs[i], "v"))
		if err1 != nil {
			logger.Log.Error("error while converting directory version to int", zap.String("reason", err1.Error()))
		}
		s2, err2 := strconv.Atoi(strings.TrimPrefix(versionDirs[j], "v"))
		if err1 != nil {
			logger.Log.Error("error while converting directory version to int", zap.String("reason", err2.Error()))
		}

		return s1 < s2
	})
	wg := new(sync.WaitGroup)
	s.lock.Lock()
	for _, fl := range s.Loaders {
		wg.Add(1)
		go func(f *FileLoader) {
			f.Load(versionDirs)
			wg.Done()
		}(fl)
	}
	s.lock.Unlock()
	wg.Wait()
	if err := s.Error(); err != nil {
		return err
	}

	s.State = LoaderStateUpdated
	return nil
}

// GetLoader gets a loader if it does not exists it creates one
func (sl *ScopeLoader) GetLoader(masterDir, file string) *FileLoader {
	sl.lock.Lock()
	defer sl.lock.Unlock()
	fileLoader, ok := sl.Loaders[file]
	if !ok {
		fileLoader = &FileLoader{
			File:  file,
			State: LoaderStateCreated,
		}
		sl.Loaders[file] = fileLoader
	}
	return fileLoader
}

// Error return all the errors faces by loaders
func (sl *ScopeLoader) Error() error {
	sl.lock.Lock()
	defer sl.lock.Unlock()
	var errs []string
	for file, ldr := range sl.Loaders {
		if ldr.Error != nil {
			errs = append(errs, file+":"+ldr.Error.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, ",\n"))
}

// FileLoader represents a file loader
type FileLoader struct {
	load        func(version string) (time.Time, error)
	File        string
	State       State
	Version     string
	versionDirs []string
	Error       error `json:"-"`
	Updated     time.Time
}

// SetLoaderFunc sets the loader function
func (l *FileLoader) SetLoaderFunc(load func(version string) (time.Time, error)) {
	l.load = load
}

// Load call the loader func
func (l *FileLoader) Load(versionDirs []string) {
	version := l.Version
	if version != "" {
		ind := findCurrentVersion(version, versionDirs)
		if ind != -1 {
			switch l.State {
			case LoaderStateCreated:

			case LoaderStateUpdated:
				versionDirs = versionDirs[ind+1:]

			case LoaderStateFailed:
				versionDirs = versionDirs[ind:]
			}
		}
	}

	fmt.Println(l.File, l.Version, versionDirs)

	for _, v := range versionDirs {
		if l.load != nil {
			l.State = LoaderStateFailed
			l.Version = v
			t, err := l.load(v)
			if err != nil {
				l.SetError(err)
				return
			}
			l.Succeeded(t)
		}
	}
}

// SetError sets the error
func (l *FileLoader) SetError(err error) {
	l.Error = err
	l.State = LoaderStateFailed
}

// Succeeded sets the updation time
func (l *FileLoader) Succeeded(t time.Time) {
	l.Updated = t
	l.State = LoaderStateUpdated
}

// UpdatedOn implements loader UpdatedOn
func (l *FileLoader) UpdatedOn() time.Time {
	return l.Updated
}

// CurrentState implements loader CurrentState
func (l *FileLoader) CurrentState() State {
	return l.State
}

// Loader gives informantion about a loadeer
type Loader interface {
	UpdatedOn() time.Time
	CurrentState() State
}

// AbortCounter keeps track of aborted mutations
type AbortCounter struct {
	count uint32
	mutex sync.Mutex
}

// IncCount increase count by one
func (ac *AbortCounter) IncCount() {
	ac.mutex.Lock()
	ac.count++
	ac.mutex.Unlock()
}

func (ac *AbortCounter) DecCount() {
	ac.mutex.Lock()
	ac.count--
	ac.mutex.Unlock()
}

// Count returns totla aborted counters
func (ac *AbortCounter) Count() uint32 {
	ac.mutex.Lock()
	c := ac.count
	ac.mutex.Unlock()
	return c
}

func findCurrentVersion(version string, versionDirs []string) int {
	for i, v := range versionDirs {
		if v == version {
			return i
		}
	}
	return -1
}
