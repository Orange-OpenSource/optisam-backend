package loader

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"optisam-backend/common/optisam/dgraph"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"

	"go.uber.org/zap"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

func init() {
	logger.Init(-1, "")
}

var batchSize int

// Config ...
type Config struct {
	// State dir containing json
	StateConfig string

	Alpha []string

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
	MasterDir             string
	ScopeSkeleten         string
	Scopes                []string
	ProductFiles          []string
	AppFiles              []string
	AppProdFiles          []string
	InstFiles             []string
	InstProdFiles         []string
	InstEquipFiles        []string
	AcqRightsFiles        []string
	EquipmentFiles        []string
	ProductEquipmentFiles []string
	MetadataFiles         *MetadataFiles
	SchemaFiles           []string
	TypeFiles             []string
	UsersFiles            []string
	Repository            v1.Equipment
	BatchSize             int
	IgnoreNew             bool
	GenerateRDF           bool
}

// MetadataFiles ...
type MetadataFiles struct {
	EquipFiles []string
}

// NewDefaultConfig ...
func NewDefaultConfig() *Config {
	return &Config{
		Alpha: []string{"127.0.0.1:9080"}, //":9084",

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
	load   func(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, files []string, ch chan<- *api.Request, doneChan <-chan struct{})
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
			load:   loadApplicationProducts,
			files:  config.AppProdFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadInstances,
			files:  config.InstFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadInstanceProducts,
			files:  config.InstProdFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadInstanceEquipments,
			files:  config.InstEquipFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadAcquiredRights,
			files:  config.AcqRightsFiles,
			scopes: config.Scopes,
		},
		{
			load:   loadProductEquipments,
			files:  config.ProductEquipmentFiles,
			scopes: config.Scopes,
		},
		// {
		// 	load:   loadUsers,
		// 	files:  config.UsersFiles,
		// 	scopes: config.Scopes,
		// },
	}
	return &AggregateLoader{
		loaders: loaders,
		config:  config,
	}
}

var ignoreNew bool
var dgCl *dgo.Dgraph

// Load .. loads data
func (al *AggregateLoader) Load() (retErr error) {
	log.Println("loader started")
	config := al.config
	ignoreNew = config.IgnoreNew
	genRDF = config.GenerateRDF
	dg, err := dgraph.NewDgraphConnection(&dgraph.Config{
		Hosts: config.Alpha,
	})
	dgCl = dg
	if err != nil {
		logger.Log.Error("Error in creating new dg connection ", zap.String("Reason", err.Error()))
		return err
	}

	// api.NewDgraphClient(server1),

	// Drop schema and all the data present in database
	if config.DropSchema {
		if err := dropSchema(dg); err != nil {
			return err
		}
	}

	// creates schema whatever is present in database
	if config.CreateSchema {
		if err := createSchema(dg, config.SchemaFiles, config.TypeFiles); err != nil {
			return err
		}
	}

	if !(config.LoadMetadata || config.LoadEquipments || config.LoadStaticData) {
		return
	}

	batchSize = config.BatchSize
	fmt.Println(batchSize)
	wg := new(sync.WaitGroup)
	ch := make(chan *api.Request)
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
		eqTypes, err := config.Repository.EquipmentTypes(context.Background(), config.Scopes)
		if err != nil {
			return err
		}
		log.Println(config.EquipmentFiles)
		loadEquipments(ml, ch, config.MasterDir, config.Scopes, config.EquipmentFiles, eqTypes, doneChan)

		wg.Add(1)
		go func() {
			ml.Load(config.MasterDir)
			wg.Done()
		}()
	}

	// Load static data like products equipments that
	// Like Products,application,acquired rights,instances
	if config.LoadStaticData {

		func() {
			for _, l := range al.loaders {
				l.load(ml, dg, config.MasterDir, l.scopes, l.files, ch, doneChan)
			}
			wg.Add(1)
			go func() {
				ml.Load(config.MasterDir)
				wg.Done()
			}()
		}()
	}

	muRetryChan := make(chan *api.Request)
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
			close(doneChan)
			wg1.Done()
		case <-filesProcessedChan:
			for {
				if ac.Count() == 0 {
					close(doneChan)
					wg1.Done()
					break
				}
				time.Sleep(2 * time.Second)
			}
		}
		// Close done chan

		// Close done chan
	}()
	mutations := uint32(0)
	//	abortedMutations := uint32(0)
	nquads := uint64(0)
	t := time.Now()

	for i := 0; i < 1; i++ {
		var file *os.File
		if genRDF {
			f, err := os.Create("data.rdf")
			if err != nil {
				panic(err)
			}
			defer f.Close()
			file = f
		}
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			for {
				var mu *api.Request
				select {
				case mu = <-ch:
					if config.GenerateRDF {
						data := make([]string, len(mu.Mutations[0].Set))
						for i, nq := range mu.Mutations[0].Set {
							if nq.ObjectId != "" {
								data[i] = fmt.Sprintf("<%s> <%s> <%s> %s .", nq.Subject, nq.Predicate, nq.ObjectId, facetsRDF(nq.Facets))
								continue
							}

							switch t := nq.ObjectValue.Val.(type) {
							case *api.Value_StrVal:
								data[i] = fmt.Sprintf("<%s> <%s> \"%s\"^^<xs:string> %s .", nq.Subject, nq.Predicate, t.StrVal, facetsRDF(nq.Facets))
							case *api.Value_DoubleVal:
								data[i] = fmt.Sprintf("<%s> <%s> \"%f\"^^<xs:float> %s .", nq.Subject, nq.Predicate, t.DoubleVal, facetsRDF(nq.Facets))
							case *api.Value_IntVal:
								data[i] = fmt.Sprintf("<%s> <%s> \"%d\"^^<xs:int> %s .", nq.Subject, nq.Predicate, t.IntVal, facetsRDF(nq.Facets))
							case *api.Value_DefaultVal:
								data[i] = fmt.Sprintf("<%s> <%s> \"%s\"^^<xs:string> %s .", nq.Subject, nq.Predicate, t.DefaultVal, facetsRDF(nq.Facets))
							default:
								log.Printf(" nq.ObjectValue.Val unsupported type: %T\n", t)
							}
							//	data[i] = fmt.Sprintf("%s <%s> %v .", nq.Subject, nq.Predicate, nq.ObjectValue.)
						}
						file.Write([]byte(strings.Join(data, "\n") + "\n"))
						continue
					}
				case <-filesProcessedChan:
					return
				}
				mu.CommitNow = true
				_, err := dg.NewTxn().Do(context.Background(), mu)
				if err != nil {
					// TODO : do not return here directly make an error and return
					ac.IncCount()
					muRetryChan <- mu
					log.Println(err)
					continue
				}
				// fmt.Printf("%+v\n", resp)
				// atomic.Add
				atomic.AddUint32(&mutations, 1)
				atomic.AddUint64(&nquads, uint64(len(mu.Mutations[0].Set)))
				// mutations++
				//	nquads += len(mu.Set)
				fmt.Printf("time elapsed[%v],completed mutations: %v,aborted: %v edges_total:%v,edges this mutation: %v \n", time.Now().Sub(t), mutations, ac.Count(), nquads, len(mu.Mutations[0].Set))
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

// Load function load all the data in
func Load(config *Config) error {
	return NewLoader(config).Load()
}

func handleAborted(in <-chan *api.Request, out chan<- *api.Request, doneChan chan struct{}, ac *AbortCounter) {
	fmt.Println("handleAborted")
	wg := new(sync.WaitGroup)
	for {
		select {
		case abortedMutation := <-in:
			wg.Add(1)
			go func(mu *api.Request) {
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

func infiniteRetry(mu *api.Request, out chan<- *api.Request, doneChan <-chan struct{}, ac *AbortCounter) {
	count := 2
retryLoop:
	for {
		select {
		case <-doneChan:
		default:
			d := time.Duration(rand.Intn(count+9) + count)
			time.Sleep(d * time.Second)
			if _, err := dgCl.NewTxn().Do(context.Background(), mu); err != nil {
				// TODO : do not return here directly make an error and return
				log.Println("infiniteRetry", d, err)
				count++
				continue
			}
			ac.DecCount()
			break retryLoop

		}
	}
}

func readFile(filename string) (*os.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func loadFile(l Loader, ch chan<- *api.Request, masterDir, scope, version string, filename, xidColumn string, doneChan <-chan struct{}, nquadFunc staticNquadFunc) (retUpdatedOn time.Time, retErr error) {
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
		// CommitNow: true,
	}
	maxUpdated := time.Time{}
	defer func() {
		if maxUpdated.After(updatedOn) {
			updatedOn = maxUpdated
			retUpdatedOn = maxUpdated
		}
	}()
	upsertsMap := make(map[string]string)
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

		if err == nil && l.CurrentState() != LoaderStateCreated && !t.After(updatedOn) {
			continue
		}

		if t.After(maxUpdated) {
			maxUpdated = t
		}

		// shouldProceed := isRowCreated
		nqs, uids, upserts, uid, scopeNquadsNeeded := nquadFunc(columns, scope, row, index)
		for i := range uids {
			upsertsMap[uids[i]] = upserts[i]
		}
		mu.Set = append(mu.Set, nqs...)
		if scopeNquadsNeeded {
			mu.Set = append(mu.Set, scopeNquad(scope, uid)...)
		}
		if err == io.EOF {
			break
		}
		if len(mu.Set) < batchSize {
			continue
		}

		select {
		case <-doneChan:
			return updatedOn, errors.New("file processing is not complete")
		case ch <- &api.Request{
			Query:     upsertQueries(upsertsMap),
			Mutations: []*api.Mutation{mu},
		}:
		}

		mu = &api.Mutation{
			//	CommitNow: true,
		}
		upsertsMap = make(map[string]string)

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
	case ch <- &api.Request{
		Query:     upsertQueries(upsertsMap),
		Mutations: []*api.Mutation{mu},
	}:
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
		{
			Subject:     uid,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(filepath.Base(scope)),
		},
	}
}

type syncMap struct {
	strMap map[string]struct{}
	//	scopeMap map[string]map[string]struct{}
	mu sync.Mutex
}

func (sm *syncMap) Add(key string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, ok := sm.strMap[key]
	if !ok {
		sm.strMap[key] = struct{}{}
		return false
	}
	// _, ok = sm.scopeMap[scope]
	// if !ok {
	// 	sm.scopeMap[scope] = make(map[string]struct{})
	// 	sm.scopeMap[scope][key] = struct{}{}
	// }
	return true
}

var sm *syncMap

func init() {
	sm = &syncMap{
		strMap: make(map[string]struct{}),
	}
}

var genRDF bool

func uidForXIDForType(xid, objType, pkPredName, pkPredVal string, types ...dgraphType) (string, []*api.NQuad, string) {

	// xid = regexp.QuoteMeta(xid)
	var uid string
	if genRDF {
		switch pkPredName {
		case "application":
			xid = strings.TrimPrefix(xid, "app_")
		case "instance":
			xid = strings.TrimPrefix(xid, "inst_")
		}
		xid = strings.Replace(xid, " ", "_", -1)
		xid = strings.Replace(xid, "{", "_", -1)
		xid = strings.Replace(xid, "}", "_", -1)
		xid = strings.Replace(xid, "*", "Y_Y", -1)
		xid = strings.Replace(xid, "-", "X_X", -1)
		xid = strings.Replace(xid, "?", "Z_Z", -1)
		uid = "_:" + xid
		if sm.Add(xid) {
			//	log.Println("skipping nquads for xid:", xid)
			return uid, nil, ""
		}
	} else {
		// id, isNew := uidForXid(xid)
		// uid = id
		// if !ignoreNew {
		// 	if !isNew {
		// 		return uid, nil
		// 	}
		// }
		uid = `uid(` + xid + `)`
	}
	upsert := xid + ` as var(func: eq(` + pkPredName + `, "` + pkPredVal + `"))`

	nqs := []*api.NQuad{
		{
			Subject:     uid,
			Predicate:   "type_name",
			ObjectValue: stringObjectValue(objType),
		},
		{
			Subject:     uid,
			Predicate:   pkPredName,
			ObjectValue: stringObjectValue(pkPredVal),
		},
	}

	for _, t := range types {
		nqs = append(nqs, &api.NQuad{
			Subject:     uid,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue(t.String()),
		})
	}

	return uid, nqs, upsert
}
