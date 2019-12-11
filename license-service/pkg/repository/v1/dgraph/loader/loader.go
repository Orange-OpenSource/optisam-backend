// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package loader

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"optisam-backend/common/optisam/dgraph"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/dgraph-io/dgraph/xidmap"
)

func init() {
	logger.Init(-1, "")
}

var batchSize int

// Config ...
type Config struct {
	BadgerDir string
	// State dir containing json
	StateConfig string
	Zero        string
	Alpha       []string

	// OPERATIONS
	// Drop Schema and all the data with it
	DropSchema bool
	// Creates schema
	CreateSchema bool
	// Load metadata from skeleton scope data
	LoadMetadata              bool
	LoadDefaultEquipmentTypes bool
	// load equipments using equiments types
	// Preconditions: equipment types must have been created
	LoadEquipments bool
	// Load static data like products equipments that
	// Like Products,application,acquired rights,instances
	LoadStaticData bool

	// Scope with csv files just with headers
	MasterDir      string
	ScopeSkeleten  string
	Scopes         []string
	ProductFiles   []string
	AppFiles       []string
	InstFiles      []string
	AcqRightsFiles []string
	EquipmentFiles []string
	MetadataFiles  *MetadataFiles
	SchemaFiles    []string
	UsersFiles     []string
	Repository     v1.License
	BatchSize      int
	IgnoreNew      bool
}

// MetadataFiles ...
type MetadataFiles struct {
	EquipFiles []string
}

// NewDefaultConfig ...
func NewDefaultConfig() *Config {
	return &Config{
		BadgerDir: "badger",
		Zero:      "localhost:5080",
		Alpha:     []string{"localhost:9080"}, //":9084",

		MetadataFiles: new(MetadataFiles),
		BatchSize:     1000,
		//	SchemaFiles:   []string{"../schema/application.schema", "../schema/products.schema", "../schema/instance.schema", "../schema/equipment.schema"},
	}
}

type intConverter struct{}

func (intConverter) convert(val string) (*api.Value, error) {
	val = strings.Replace(val, ",", "", -1)
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("data: %s, error: %v", val, err)
	}
	return &api.Value{
		Val: &api.Value_IntVal{
			IntVal: intVal,
		},
	}, nil
}

type floatConverter struct{}

func (floatConverter) convert(val string) (*api.Value, error) {
	val = strings.Replace(val, ",", "", -1)
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	return &api.Value{
		Val: &api.Value_DoubleVal{
			DoubleVal: floatVal,
		},
	}, nil
}

type stringConverter struct {
}

func (stringConverter) convert(val string) (*api.Value, error) {
	return stringObjectValue(val), nil
}

type defaultConverter struct {
}

func (defaultConverter) convert(val string) (*api.Value, error) {
	return defaultObjectValue(val), nil
}

type converter interface {
	convert(val string) (*api.Value, error)
}

type schemaNode struct {
	predName string
	dataConv converter
}

type loader struct {
	load   func(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, files []string, ch chan<- *api.Mutation, doneChan <-chan struct{})
	files  []string
	scopes []string
	errors []error
}

// AggregateLoader loads dgraph data dgraph
type AggregateLoader struct {
	loaders []loader
	config  *Config
}

// NewLoader loads aggregate loader
func NewLoader(config *Config) *AggregateLoader {
	loaders := []loader{
		{
			load:   loadProducts,
			files:  config.ProductFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadApplications,
			files:  config.AppFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadInstances,
			files:  config.InstFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadAcquiredRights,
			files:  config.AcqRightsFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadUsers,
			files:  config.UsersFiles,
			scopes: config.Scopes,
		},
	}
	return &AggregateLoader{
		loaders: loaders,
		config:  config,
	}
}

var ignoreNew bool

// Load .. loads data
func (al *AggregateLoader) Load() (retErr error) {
	log.Println("loader started")
	config := al.config
	ignoreNew = config.IgnoreNew
	dg, err := dgraph.NewDgraphConnection(&dgraph.Config{
		Hosts: config.Alpha,
	})
	if err != nil {
		return err
	}

	//api.NewDgraphClient(server1),

	// Drop schema and all the data present in database
	if config.DropSchema {
		if err := dropSchema(dg); err != nil {
			return err
		}
	}

	// creates schema whatever is present in database
	if config.CreateSchema {
		if err := createSchema(dg, config.SchemaFiles); err != nil {
			return err
		}
	}

	if !(config.LoadMetadata || config.LoadEquipments || config.LoadStaticData) {
		return
	}

	batchSize = config.BatchSize
	fmt.Println(batchSize)
	wg := new(sync.WaitGroup)
	ch := make(chan *api.Mutation)
	doneChan := make(chan struct{})
	// Load metadata from equipments files
	if config.LoadMetadata {
		for _, filename := range config.MetadataFiles.EquipFiles {
			wg.Add(1)
			go func(fn string) {
				loadEquipmentMetadata(ch, doneChan, fn)
				wg.Done()
			}(config.ScopeSkeleten + "/" + filename)
		}
	}

	ml := &MasterLoader{}

	if config.LoadStaticData || config.LoadEquipments {
		opts := badger.DefaultOptions(config.BadgerDir)
		opts.Truncate = true
		db, err := badger.Open(opts)
		if err != nil {
			return err
		}

		defer db.Close()

		zero, err := grpc.Dial(config.Zero, grpc.WithInsecure())
		if err != nil {
			return err
		}

		defer zero.Close()

		xidMap = xidmap.New(db, zero, xidmap.Options{
			NumShards: 32,
			LRUSize:   4096,
		})

		defer xidMap.EvictAll()

		file := config.StateConfig

		m, err := newMasterLoaderFromFile(file)
		if err != nil {
			logger.Log.Error("state file not found all data will be processod", zap.String("state_file", file), zap.Error(err))
			ml = &MasterLoader{
				Loaders: make(map[string]*ScopeLoader),
			}
		} else {
			ml = m
		}

		defer func() {
			if err := saveMaterLoaderTofile(file, ml); err != nil {
				logger.Log.Error("cannot save state", zap.Error(err))
			}
		}()
	}

	// load equipments using equiments types
	// Preconditions: equipment types must have been created
	if config.LoadEquipments {
		eqTypes, err := config.Repository.EquipmentTypes(context.Background(), []string{})
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			log.Println(config.EquipmentFiles)
			loadEquipments(ml, ch, config.MasterDir, config.Scopes, config.EquipmentFiles, eqTypes, doneChan)
			wg.Done()
		}()
	}

	// Load static data like products equipments that
	// Like Products,application,acquired rights,instances
	if config.LoadStaticData {
		for _, ldr := range al.loaders {
			wg.Add(1)
			go func(l loader) {
				l.load(ml, dg, config.MasterDir, l.scopes, l.files, ch, doneChan)
				wg.Done()
			}(ldr)
		}
	}

	muRetryChan := make(chan *api.Mutation)
	ac := new(AbortCounter)

	filesProcessedChan := make(chan struct{})

	go func() {
		wg.Wait()
		// wg wait above will make sure that all the gouroutines possibly writing on this
		// channel are returened so that there are no panic if someone write on this channel.
		close(filesProcessedChan)
		fmt.Println("file processing done")
	}()

	wg1 := new(sync.WaitGroup)
	wg1.Add(1)
	go func() {
		handleAborted(muRetryChan, ch, doneChan, ac)
		wg1.Done()
	}()
	wg1.Add(1)
	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan)
		select {
		case sig := <-sigChan:
			log.Println(sig.String())
		case <-filesProcessedChan:
		}
		// Close done chan
		close(doneChan)
		wg1.Done()
		// Close done chan
	}()
	mutations := 0
	nquads := 0
	t := time.Now()

	for i := 0; i < 1; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			for {
				var mu *api.Mutation
				select {
				case mu = <-ch:
				case <-filesProcessedChan:
					return
				}
				if _, err := dg.NewTxn().Mutate(context.Background(), mu); err != nil {
					// TODO : do not return here directly make an error and return
					muRetryChan <- mu
					log.Println(err)
				}
				mutations++
				nquads += len(mu.Set)
				fmt.Printf("time elapsed[%v], completed mutations: %v, edges_total:%v,edges this mutation: %v \n", time.Now().Sub(t), mutations, nquads, len(mu.Set))
				// log.Println(ass.GetUids())
			}
		}()
	}
	wg1.Wait()
	if ac.Count() != 0 {
		return fmt.Errorf("cannot compete :%d number of aborted mutations\n %v", ac.Count(), ml.Error())
	}
	return ml.Error()
}

var xidMap *xidmap.XidMap

// Load function load all the data in
func Load(config *Config) error {
	return NewLoader(config).Load()
}

func handleAborted(in <-chan *api.Mutation, out chan<- *api.Mutation, doneChan chan struct{}, ac *AbortCounter) {
	fmt.Println("handleAborted")
	wg := new(sync.WaitGroup)
	for {
		select {
		case abortedMutation := <-in:
			wg.Add(1)
			go func(mu *api.Mutation) {
				infiniteRetry(mu, out, doneChan, ac)
				wg.Done()
			}(abortedMutation)
		case <-doneChan:
			fmt.Println("handleAborted - revcevied done")
			wg.Wait()
			fmt.Println("handleAborted - all infiniteRetry done")
			return
		}
	}
}

func infiniteRetry(mu *api.Mutation, out chan<- *api.Mutation, doneChan <-chan struct{}, ac *AbortCounter) {
	select {
	case <-doneChan:
		ac.IncCount()
	case out <- mu:
	}
}

func readFile(filename string) (*os.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func loadFile(l Loader, ch chan<- *api.Mutation, masterDir, scope, version string, filename, xidColumn string, doneChan <-chan struct{}, nquadsGen func(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, string)) (time.Time, error) {
	updatedOn := l.UpdatedOn()
	filename = filepath.Join(masterDir, scope, version, filename)
	log.Println("started loading " + filename)
	defer log.Println("end loading " + filename)
	f, err := readFile(filename)
	if err != nil {
		logger.Log.Error("error opening file", zap.String("filename:", filename), zap.String("reason", err.Error()))
		return updatedOn, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = ';'
	columns, err := r.Read()
	if err == io.EOF {
		return updatedOn, err
	} else if err != nil {
		logger.Log.Error("error reading header ", zap.String("filename:", filename), zap.String("reason", err.Error()))
		return updatedOn, err
	}
	// find swidtag index
	index := findColoumnIdx(xidColumn, columns)
	updatedIdx := findColoumnIdx(updatedColumnName, columns)
	createdIdx := findColoumnIdx(createdColumnName, columns)
	//	log.Println(coloumns)

	//	log.Println(applicationSchema)

	if index < 0 {
		logger.Log.Error("cannot find xid", zap.String("filename:", filename), zap.String("XID_col", xidColumn))
		return updatedOn, err
	}
	mu := &api.Mutation{
		CommitNow: true,
	}
	maxUpdated := time.Time{}
	defer func() {
		if maxUpdated.After(updatedOn) {
			updatedOn = maxUpdated
		}
	}()
	for {
		row, err := r.Read()
		if err == io.EOF {
			if len(row) == 0 {
				break
			}
		}
		if err != nil {
			logger.Log.Error("error reading header ", zap.String("filename:", filename), zap.String("reason", err.Error()))
			return updatedOn, err
		}
		for i := range row {
			row[i] = strings.TrimSpace(row[i])
		}

		t, err := isRowDirty(row, updatedIdx, createdIdx)
		if err != nil {
			// We got an error while checking the state we must consider this as dirty state
			logger.Log.Error("error while checking state", zap.Error(err), zap.String("file", filename))
		}

		if l.CurrentState() != LoaderStateCreated && !t.After(updatedOn) {
			continue
		}

		maxUpdated = t

		//shouldProceed := isRowCreated
		nqs, uid := nquadsGen(columns, scope, row, index)
		mu.Set = append(mu.Set, append(nqs, scopeNquad(scope, uid)...)...)
		if err == io.EOF {
			break
		}
		if len(mu.Set) < batchSize {
			continue
		}

		select {
		case <-doneChan:
			return updatedOn, errors.New("file processing is not complete")
		case ch <- mu:
		}

		mu = &api.Mutation{
			CommitNow: true,
		}

	}

	if len(mu.Set) == 0 {
		return updatedOn, nil
	}
	select {
	case <-doneChan:
		if len(mu.Set) != 0 {
			return updatedOn, errors.New("file processing is not complete after eof")
		}
		return updatedOn, nil
	case ch <- mu:
	}

	return updatedOn, nil
}

func stringObjectValue(val string) *api.Value {
	return &api.Value{
		Val: &api.Value_StrVal{
			StrVal: val,
		},
	}
}

func defaultObjectValue(val string) *api.Value {
	return &api.Value{
		Val: &api.Value_DefaultVal{
			DefaultVal: val,
		},
	}
}

func scopeNquad(scope, uid string) []*api.NQuad {
	return []*api.NQuad{
		&api.NQuad{
			Subject:     uid,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(filepath.Base(scope)),
		},
	}
}

func uidForXid(xid string) (string, bool) {
	uid, isNew := xidMap.AssignUid(xid)
	return fmt.Sprintf("%#x", uint64(uid)), isNew
}

func uidForXIDForType(xid, objType, pkPredName, pkPredVal string) (string, []*api.NQuad) {
	uid, isNew := uidForXid(xid)
	if !ignoreNew {
		if !isNew {
			return uid, nil
		}
	}
	return uid, []*api.NQuad{
		&api.NQuad{
			Subject:     uid,
			Predicate:   "type",
			ObjectValue: stringObjectValue(objType),
		}, &api.NQuad{
			Subject:     uid,
			Predicate:   pkPredName,
			ObjectValue: stringObjectValue(pkPredVal),
		},
	}
}
