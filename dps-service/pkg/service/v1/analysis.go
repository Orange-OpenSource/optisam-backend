package v1

import (
	"context"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/config"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	excel "github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// action
	MissingMandatoryHeader string = "MissingMandatoryHeader"
	MissingMandatoryValue  string = "MissingMandatoryValue"
	Inconsistent1          string = "inconsistent1"
	Inconsistent2          string = "inconsistent2"
	WrongTypeField         string = "wrongTypeField"
	DuplicateLine          string = "duplicateLine"
	MandatoryHeader        string = "mandatoryHeader"
	MissingField           string = "missingField"
	DuplicateHeader        string = "duplicateHeader"
	BadReference           string = "badReference"

	// errors
	BadFile              string = "BadOrCorruptFile"
	InvalidFileExtension string = "Invalid file extension, expecting file.xlsx"
	SuccessfulAnalysis   string = "Analysis has been done, please check the report"
	GlobalFileExtension  string = "xlsx"
	BadSheet             string = "BadOrCorruptSheet"
	Wished               int    = 0
	MandatoryWithBlank   int    = 1
	Mandatory            int    = 2
	BadObject            int    = 1
	MissingObject        int    = 0
	GoodObject           int    = 2
	COMPLETED            string = "COMPLETED"
	FAILED               string = "FAILED"
	PARTIAL              string = "PARTIAL"
	servers              string = "servers"
	softpartitions       string = "softpartitions"
	products             string = "products"
	acquiredRights       string = "acquiredRights"
	DEFAULT              string = "default"
)

type Info struct {
	IsMandatory int // 0 : wished, 1: mandatory but field can be blank, 2: header and value both are mandatory
	DataType    int
}

const (
	STRING int = iota
	INT
	FLOAT64
	DATE
)

var (
	dataTypes map[int]string = map[int]string{0: "string", 1: "integer", 2: "float", 3: "DD-MM-YYYY or DD/MM/YYYY"}

	sheetsAndHeaders map[string]map[string]Info = map[string]map[string]Info{
		servers:        {"server_name": Info{MandatoryWithBlank, STRING}, "server_id": Info{Mandatory, STRING}, "server_type": Info{Wished, STRING}, "server_os": Info{Wished, STRING}, "cpu_model": Info{Mandatory, STRING}, "cores_per_processor": Info{Mandatory, INT}, "hyperthreading": Info{Wished, STRING}, "cluster_name": Info{MandatoryWithBlank, STRING}, "vcenter_name": Info{MandatoryWithBlank, STRING}, "vcenter_version": Info{Wished, STRING}, "datacenter_name": Info{Wished, STRING}, "ibm_pvu": Info{MandatoryWithBlank, FLOAT64}, "sag_uvu": Info{MandatoryWithBlank, INT}, "cpu_manufacturer": Info{MandatoryWithBlank, STRING}, "server_processors_numbers": Info{Mandatory, INT}},
		acquiredRights: {"maintenance_provider": Info{MandatoryWithBlank, STRING}, "last_po": Info{MandatoryWithBlank, STRING}, "support_number": Info{MandatoryWithBlank, STRING}, "software_provider": Info{MandatoryWithBlank, STRING}, "ordering_date": Info{MandatoryWithBlank, DATE}, "csc": Info{MandatoryWithBlank, STRING}, "sku": Info{Mandatory, STRING}, "product_name": Info{Mandatory, STRING}, "product_version": Info{Mandatory, STRING}, "product_editor": Info{Mandatory, STRING}, "metric": Info{Mandatory, STRING}, "licence_type": Info{Wished, STRING}, "acquired_licenses": Info{Mandatory, INT}, "unit_price": Info{Mandatory, FLOAT64}, "maintenance_licences": Info{MandatoryWithBlank, INT}, "maintenance_unit_price": Info{MandatoryWithBlank, FLOAT64}, "maintenance_start": Info{MandatoryWithBlank, DATE}, "maintenance_end": Info{MandatoryWithBlank, DATE}},
		softpartitions: {"softpartition_name": Info{MandatoryWithBlank, STRING}, "softpartition_id": Info{Mandatory, STRING}, "server_id": Info{Mandatory, STRING}},
		products:       {"product_name": Info{Mandatory, STRING}, "product_version": Info{Mandatory, STRING}, "product_editor": Info{Mandatory, STRING}, "host_id": Info{Mandatory, STRING}, "domain": Info{MandatoryWithBlank, STRING}, "environment": Info{MandatoryWithBlank, STRING}, "application_name": Info{MandatoryWithBlank, STRING}, "application_id": Info{MandatoryWithBlank, STRING}, "application_instance_name": Info{MandatoryWithBlank, STRING}, "number_of_access": Info{MandatoryWithBlank, INT}}}

	actionAndColors map[string]string = map[string]string{Inconsistent1: "#FFC0CB", Inconsistent2: "#C0C0C0", WrongTypeField: "#FFA500", DuplicateLine: "#1986EE", MandatoryHeader: "#7FFD4", MissingField: "#008000", DuplicateHeader: "#808000", BadReference: "#F5CBA7"}

	corefactors        map[string]*Node
	isCoreFactorStored bool
	mu                 sync.Mutex
)

type Node struct {
	Key   string
	Value float64
	Edges map[string]*Node
}

type ObjectCommentInfo struct {
	Msg          string
	Action       string
	Column       string   // single column need to highlight
	IsFullRow    bool     // true for full row
	Coordinates  []int    // cordinates of full row
	ColumnRanges []string // multiple columns need to highlight
}

type BadReferenceInfo struct {
	Data      [][]string
	RowIndics []int
}

// setCorefactorInCache
func cacheCorefactor(ctx context.Context, d *dpsServiceServer) error {
	if !isCoreFactorCached() {
		dbresp, err := d.dpsRepo.GetCoreFactorList(ctx)
		if err != nil {
			logger.Log.Error("Failed to save core factor in cache", zap.Error(err))
			return err
		}
		corefactors = make(map[string]*Node)
		for _, v := range dbresp {
			mf := strings.ToLower(v.Manufacturer)
			ml := strings.ToLower(v.Model)
			cf, err := strconv.ParseFloat(v.CoreFactor, 64)
			if err != nil {
				logger.Log.Error("bad core factor value", zap.Error(err))
				return err
			}
			if corefactors[mf] == nil {
				corefactors[mf] = &Node{Key: strings.TrimSpace(mf)}
			}
			if mf == DEFAULT {
				corefactors[mf].Value = cf
				continue
			}
			root := corefactors[mf]
			list := strings.Split(ml, " ")
			for i, val := range list {
				if root.Edges == nil {
					root.Edges = make(map[string]*Node)
				}
				if root.Edges[val] == nil {
					node := &Node{
						Key: strings.TrimSpace(val),
					}
					if i == len(list)-1 || val == DEFAULT {
						node.Value = cf
					}
					root.Edges[val] = node
				}
				root = root.Edges[val]
			}
		}
		coreFactorCached()
	}
	return nil
}

func getCoreFactor(manufacturer, model string) float64 {
	model = strings.ToLower(model)
	manufacturer = strings.ToLower(manufacturer)
	var defCf float64
	logger.Log.Debug("getting core factor for ", zap.String("manufacturer", manufacturer), zap.String("model", model))
	if corefactors[DEFAULT] != nil {
		defCf = corefactors[DEFAULT].Value
	}
	if manufacturer == DEFAULT {
		return defCf
	}
	root := corefactors[manufacturer]
	if root == nil {
		logger.Log.Error("No manufaturer found using default", zap.String("manufacturer", manufacturer), zap.Any("cf", defCf))
		return defCf
	}
	list := strings.Split(model, " ")
	if len(list) == 0 {
		return defCf
	}
	for _, v := range list {
		v = strings.TrimSpace(v)
		if root.Edges[v] != nil {
			root = root.Edges[v]
		} else if root.Edges[DEFAULT] != nil {
			root = root.Edges[DEFAULT]
		} else {
			break
		}
	}

	if root.Value == 0.0 {
		return defCf
	}
	return root.Value
}

func (d *dpsServiceServer) DataAnalysis(ctx context.Context, req *v1.DataAnalysisRequest) (*v1.DataAnalysisResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	if !isFileExtensionValid(req.File) {
		logger.Log.Error("Couldn't perform analysis ,Invalid file extension", zap.Any("received", req.File))
		return &v1.DataAnalysisResponse{
			Status:      FAILED,
			Description: InvalidFileExtension,
		}, nil
	}

	file := fmt.Sprintf("%s/%s/analysis/%s", config.GetConfig().RawdataLocation, req.Scope, req.File)
	fp, err := excel.OpenFile(file)
	if err != nil {
		logger.Log.Error("Failed to open file for analysis", zap.Error(err), zap.Any("file", file))
		return &v1.DataAnalysisResponse{
			Status:      FAILED,
			Description: BadFile,
		}, nil
	}

	if err = isSheetMissing(fp); err != nil {
		logger.Log.Error("Couldn't perform analysis as sheets are missing", zap.Error(err))
		return &v1.DataAnalysisResponse{
			Status:      FAILED,
			Description: err.Error(),
		}, nil
	}

	var headersIndex map[string]map[string]int
	sheetSeq := make(map[string]int)
	headersIndex, err = handleHeaders(fp, sheetSeq)
	if err != nil {
		status, description := getErrorResponse(err)
		if status != "" {
			return &v1.DataAnalysisResponse{Description: description, Status: status}, nil
		}
	}
	analysisStatus := COMPLETED
	description := SuccessfulAnalysis

	goodObjQueue := make(chan map[string][][]string, 4)
	badObjQueue := make(chan map[string][][]string, 4)
	mainObjQueue := make(chan map[string][]ObjectCommentInfo, 4)
	badServersQueue := make(chan map[string]int, 1)
	badSoftpartitionsQueue := make(chan map[string]int, 1)

	if err := cacheCorefactor(ctx, d); err != nil {
		logger.Log.Error("Failed to create N-array cache of corefactor", zap.Error(err))
		return &v1.DataAnalysisResponse{Description: "CoreFactorCachingError", Status: FAILED}, nil
	}
	var g errgroup.Group
	g.Go(func() error {
		err := analyzeServerSheet(fp, goodObjQueue, badObjQueue, mainObjQueue, badServersQueue, headersIndex[servers])
		return err
	})

	g.Go(func() error {
		err := analyzeSoftpartitionSheet(fp, goodObjQueue, badObjQueue, mainObjQueue, badServersQueue, badSoftpartitionsQueue, headersIndex[softpartitions])
		return err
	})
	g.Go(func() error {
		err := analyzeProductSheet(fp, goodObjQueue, badObjQueue, mainObjQueue, badSoftpartitionsQueue, headersIndex[products])
		return err
	})
	g.Go(func() error {
		err := analyzeAcquiredRightSheet(fp, goodObjQueue, badObjQueue, mainObjQueue, headersIndex[acquiredRights])
		return err
	})
	g.Go(func() error {
		err := goodObjWriter(req.File, req.Scope, headersIndex, goodObjQueue)
		return err
	})
	g.Go(func() error {
		err := badObjWriter(req.File, req.Scope, headersIndex, badObjQueue)
		return err
	})
	g.Go(func() error {
		err := mainObjWriter(fp, req.Scope, req.File, mainObjQueue, sheetSeq)
		return err
	})

	if err := g.Wait(); err != nil {
		logger.Log.Error("Analysing is failed in asyn process", zap.Error(err))
		analysisStatus, description = getErrorResponse(err)
		return &v1.DataAnalysisResponse{Description: description, Status: analysisStatus}, nil
	}

	return &v1.DataAnalysisResponse{
		Report:      fmt.Sprintf("api/v1/import/download?fileName=%s&downloadType=analysis&scope=%s", req.File, req.Scope),
		TargetFile:  fmt.Sprintf("api/v1/import/download?fileName=good_%s&downloadType=analysis&scope=%s", req.File, req.Scope),
		ErrorFile:   fmt.Sprintf("api/v1/import/download?fileName=bad_%s&downloadType=error&scope=%s", req.File, req.Scope),
		Description: description, Status: analysisStatus}, nil

}

func getErrorResponse(err error) (status string, msg string) {
	if strings.Contains(err.Error(), "analysis:") { // got failed case
		status = FAILED
		msg = err.Error()
	} else {
		status = PARTIAL
		msg = "InternalError"
	}
	return
}

// isFileExtensionValid checks for global file ext.
func isFileExtensionValid(name string) bool {
	list := strings.Split(name, ".")
	if len(list) == 0 {
		return false
	}
	if list[len(list)-1] != GlobalFileExtension {
		return false
	}
	return true
}

// checkAnalysisPreCondintions checks missing sheet, empty file(not sheet), wrong sheets
func isSheetMissing(fp *excel.File) error {
	errMsg := ""
	count := 0
	expectedSheets := map[string]bool{"servers": false, "softpartitions": false, "products": false, "acquiredRights": false}
	for _, v := range fp.GetSheetList() {
		expectedSheets[v] = true
	}
	for k, v := range expectedSheets {
		if !v {
			errMsg += fmt.Sprintf("%s ,", k)
			count++
		}
	}
	if errMsg != "" {
		errMsg = strings.TrimSuffix(errMsg, ",")
		if count > 1 {
			errMsg += " sheets are missing"
		} else {
			errMsg += " sheet is missing"
		}
		return fmt.Errorf("%s", errMsg)
	}
	return nil
}

func analyzeServerSheet(fp *excel.File, goodObjQueue, badObjQueue chan map[string][][]string, mainObjQueue chan map[string][]ObjectCommentInfo, badServersQueue chan map[string]int, headersIndex map[string]int) error {
	logger.Log.Debug("Starting server=================================================")

	checkDuplicates := make(map[string]int)
	inconsitency := make(map[string]string)
	var goodObj, badObj [][]string
	badServers := make(map[string]int)
	mainObj := make(map[string][]ObjectCommentInfo)
	rows, err := fp.GetRows(servers)
	if err != nil {
		logger.Log.Error("Failed to read the sheet", zap.Error(err), zap.Any("sheet", servers))
		return errors.New("failedToReadServerSheet")
	}

	for i := 1; i < len(rows); i++ {
		serverID := getCellValue(fp, "server_id", i+1, headersIndex, servers)
		if checkDuplicates[strings.Join(rows[i], "|")] == 0 {
			checkDuplicates[strings.Join(rows[i], "|")] = i
			objects, ok, err := handleMandatoryFieldMissingAndWrongType(fp, i+1, servers, headersIndex)
			if err != nil {
				logger.Log.Error("Error in handleMandatoryFieldMissingAndWrongType", zap.Any("sheet", servers), zap.Error(err))
				return err
			}
			if ok {
				obj, ok := handleServerSheetInconsistency(serverID, rows[i], inconsitency, headersIndex, i+1)
				if ok {
					goodObj = append(goodObj, rows[i])
					badServers[serverID] = 2
				} else {
					col, err := excel.ColumnNumberToName(headersIndex["server_id"])
					if err != nil {
						logger.Log.Error("Failed to get column in server ", zap.Any("col_number", headersIndex["server_id"]), zap.Error(err))
						return err
					}
					obj.Column = fmt.Sprintf("%s%d", col, i)
					badObj = append(badObj, rows[i])
					mainObj[servers] = append(mainObj[servers], obj)
					badServers[serverID] = 1
				}
			} else {
				badObj = append(badObj, rows[i])
				mainObj[servers] = append(mainObj[servers], objects...)
				badServers[serverID] = 1
			}
		} else {
			mainObj[servers] = append(mainObj[servers], ObjectCommentInfo{
				Msg:         fmt.Sprintf("This row is duplicate with row no %d", checkDuplicates[strings.Join(rows[i], "|")]+1),
				Action:      DuplicateLine,
				IsFullRow:   true,
				Coordinates: []int{len(rows[i]), i + 1},
			})
			badObj = append(badObj, rows[i])
			badServers[serverID] = 1
		}
	}

	logger.Log.Debug("Filtered server msg objects ", zap.Any("goodObj", goodObj), zap.Any("baddObj", badObj), zap.Any("mainObj", mainObj), zap.Any("BadServers", badServers))
	queueData := make(map[string][][]string)
	queueData[servers] = goodObj
	goodObjQueue <- queueData
	queueData = make(map[string][][]string)
	queueData[servers] = badObj
	badObjQueue <- queueData
	mainObjQueue <- mainObj
	badServersQueue <- badServers
	logger.Log.Debug("end server=================================================")
	return nil
}

func analyzeSoftpartitionSheet(fp *excel.File, goodObjQueue, badObjQueue chan map[string][][]string, mainObjQueue chan map[string][]ObjectCommentInfo, badServersQueue, badSoftpartitionsQueue chan map[string]int, headersIndex map[string]int) error {
	logger.Log.Debug("Starting partition=================================================")

	checkDuplicates := make(map[string]int)
	inconsitency := make(map[string]string)
	var goodObj, badObj [][]string
	badSoftpartitions := make(map[string]int)
	softpartitionInfo := make(map[string]BadReferenceInfo)
	mainObj := make(map[string][]ObjectCommentInfo)
	rows, err := fp.GetRows(softpartitions)
	if err != nil {
		logger.Log.Error("Failed to read the sheet", zap.Error(err), zap.Any("sheet", softpartitions))
		return errors.New("failedToReadSoftpartitionSheet")
	}

	for i := 1; i < len(rows); i++ {
		serverID := getCellValue(fp, "server_id", i+1, headersIndex, softpartitions)
		partitionID := getCellValue(fp, "softpartition_id", i+1, headersIndex, softpartitions)
		if checkDuplicates[strings.Join(rows[i], "|")] == 0 {
			checkDuplicates[strings.Join(rows[i], "|")] = i
			objects, ok, err := handleMandatoryFieldMissingAndWrongType(fp, i+1, softpartitions, headersIndex)
			if err != nil {
				logger.Log.Error("Error in handleMandatoryFieldMissingAndWrongType ", zap.Any("sheet", softpartitions), zap.Error(err))
				return err
			}
			if ok {
				obj, ok := handleSoftpartitionInconsistency(fp, headersIndex, inconsitency, i+1)
				if ok {
					if _, y := softpartitionInfo[serverID]; !y {
						softpartitionInfo[serverID] = BadReferenceInfo{}
					}
					temp := softpartitionInfo[serverID]
					temp.Data = append(temp.Data, rows[i])
					temp.RowIndics = append(temp.RowIndics, i)
					softpartitionInfo[serverID] = temp
					if badSoftpartitions[partitionID] == 0 {
						badSoftpartitions[partitionID] = 2
					}
				} else {
					badObj = append(badObj, rows[i])
					mainObj[softpartitions] = append(mainObj[softpartitions], obj)
					if badSoftpartitions[partitionID] == 0 {
						badSoftpartitions[partitionID] = 1
					}
				}
			} else {
				badObj = append(badObj, rows[i])
				mainObj[softpartitions] = append(mainObj[softpartitions], objects...)
				if badSoftpartitions[partitionID] == 0 {
					badSoftpartitions[partitionID] = 1
				}
			}
		} else {
			mainObj[softpartitions] = append(mainObj[softpartitions], ObjectCommentInfo{
				Msg:         fmt.Sprintf("This row is duplicate with row no %d", checkDuplicates[strings.Join(rows[i], "|")]+1),
				Action:      DuplicateLine,
				IsFullRow:   true,
				Coordinates: []int{len(rows[i]), i + 1},
			})
			badObj = append(badObj, rows[i])
			if badSoftpartitions[partitionID] == 0 {
				badSoftpartitions[partitionID] = 1
			}
		}
	}

	for {
		if len(badServersQueue) == 1 {
			close(badServersQueue)
			serverInfo := make(map[string]int)
			for key, val := range <-badServersQueue {
				serverInfo[key] = val
				if _, ok := badSoftpartitions[key]; !ok {
					badSoftpartitions[key] = val
				}
			}
			for server, val := range softpartitionInfo {
				var obj []ObjectCommentInfo
				if serverInfo[server] == GoodObject {
					goodObj = append(goodObj, val.Data...)
				} else {
					badObj = append(badObj, val.Data...)
					obj = getBadObjectsForComment(val, server, badSoftpartitions, headersIndex, "softpartition_id")
				}
				if len(obj) > 0 {
					mainObj[softpartitions] = append(mainObj[softpartitions], obj...)
				}
			}
			logger.Log.Debug("Filtered softpartition msg objects ", zap.Any("goodObj", goodObj), zap.Any("baddObj", badObj), zap.Any("mainObj", mainObj), zap.Any("softpartitionInfo", badSoftpartitions))
			queueData := make(map[string][][]string)
			queueData[softpartitions] = badObj
			badObjQueue <- queueData
			mainObjQueue <- mainObj
			queueData = make(map[string][][]string)
			queueData[softpartitions] = goodObj
			goodObjQueue <- queueData
			badSoftpartitionsQueue <- badSoftpartitions
			break
		} else {
			time.Sleep(100 * time.Millisecond)
			logger.Log.Debug("waiting for bad servers.......")
		}
	}
	logger.Log.Debug("end partition=================================================")
	return nil
}

func getBadObjectsForComment(data BadReferenceInfo, key string, badObjectsForQueue map[string]int, headersIndex map[string]int, badObjKey string) []ObjectCommentInfo {
	obj := []ObjectCommentInfo{}
	for i, val := range data.Data {
		var colName string
		id := ""
		var err error
		if badObjKey == "softpartition_id" {
			id = "server_id"
		} else {
			id = "host_id"
		}
		colName, err = excel.ColumnNumberToName(headersIndex[id])
		if err != nil {
			logger.Log.Error("Failed to get colm name for server in softpartiton analysis", zap.Error(err))
			colName = "A"
		}
		obj = append(obj, ObjectCommentInfo{
			Msg:    fmt.Sprintf("Bad reference, Either server %s is missing or containing errors", key),
			Action: BadReference,
			Column: fmt.Sprintf("%s%d", colName, data.RowIndics[i]+1),
		})
		if badObjKey != "" {
			badObjectsForQueue[val[headersIndex[badObjKey]-1]] = 1
		}
	}
	return obj
}

func analyzeProductSheet(fp *excel.File, goodObjQueue, badObjQueue chan map[string][][]string, mainObjQueue chan map[string][]ObjectCommentInfo, badSoftpartitionQueue chan map[string]int, headersIndex map[string]int) error {
	logger.Log.Debug("Starting product=================================================")

	checkDuplicates := make(map[string]int)
	inconsitency := make(map[string]string)
	var goodObj, badObj [][]string
	productInfo := make(map[string]BadReferenceInfo)
	mainObj := make(map[string][]ObjectCommentInfo)
	rows, err := fp.GetRows(products)
	if err != nil {
		logger.Log.Error("Failed to read the sheet", zap.Error(err), zap.Any("sheet", products))
		return errors.New("failedToReadProductSheet")
	}

	for i := 1; i < len(rows); i++ {
		var temp BadReferenceInfo
		hostID := getCellValue(fp, "host_id", i+1, headersIndex, products)
		if checkDuplicates[strings.Join(rows[i], "|")] == 0 {
			checkDuplicates[strings.Join(rows[i], "|")] = i
			objects, ok, err := handleMandatoryFieldMissingAndWrongType(fp, i+1, products, headersIndex)
			if err != nil {
				return err
			}
			if ok {
				obj, ok := handleProductSheetInconsistency(fp, headersIndex, inconsitency, i+1)
				if ok {
					if _, y := productInfo[hostID]; !y {
						productInfo[hostID] = BadReferenceInfo{}
					}
					temp = productInfo[hostID]
					temp.Data = append(temp.Data, rows[i])
					temp.RowIndics = append(temp.RowIndics, i)
					productInfo[hostID] = temp
				} else {
					col, err := excel.ColumnNumberToName(headersIndex["application_id"])
					if err != nil {
						logger.Log.Error("Failed to get column in product ", zap.Any("col_number", headersIndex["application_id"]), zap.Error(err))
						return err
					}
					obj.Column = fmt.Sprintf("%s%d", col, i)
					badObj = append(badObj, rows[i])
					mainObj[products] = append(mainObj[products], obj)
				}
			} else {
				badObj = append(badObj, rows[i])
				mainObj[products] = append(mainObj[products], objects...)
			}
		} else {
			mainObj[products] = append(mainObj[products], ObjectCommentInfo{
				Msg:         fmt.Sprintf("This row is duplicate with row no %d", checkDuplicates[strings.Join(rows[i], "|")]+1),
				Action:      DuplicateLine,
				IsFullRow:   true,
				Coordinates: []int{len(rows[i]), i + 1},
			})
			badObj = append(badObj, rows[i])
		}
	}

	for {
		if len(badSoftpartitionQueue) == 1 {
			close(badSoftpartitionQueue)
			badSoftpartitions := make(map[string]int)
			for hostID, val := range <-badSoftpartitionQueue {
				badSoftpartitions[hostID] = val
			}
			for hostID, val := range productInfo {
				var obj []ObjectCommentInfo
				if badSoftpartitions[hostID] == GoodObject {
					goodObj = append(goodObj, val.Data...)
				} else {
					badObj = append(badObj, val.Data...)
					obj = getBadObjectsForComment(val, hostID, nil, headersIndex, "")
				}
				if len(obj) > 0 {
					mainObj[products] = append(mainObj[products], obj...)
				}
			}
			logger.Log.Debug("Filtered product msg objects ", zap.Any("goodObj", goodObj), zap.Any("baddObj", badObj), zap.Any("mainObj", mainObj))
			queueData := make(map[string][][]string)
			queueData[products] = goodObj
			goodObjQueue <- queueData
			queueData = make(map[string][][]string)
			queueData[products] = badObj
			badObjQueue <- queueData
			mainObjQueue <- mainObj
			break
		} else {
			time.Sleep(100 * time.Millisecond)
			logger.Log.Debug("Wainting for bad softpartitionQuue.....")
		}
	}
	logger.Log.Debug("end product =================================================")
	return nil
}

func analyzeAcquiredRightSheet(fp *excel.File, goodObjQueue, badObjQueue chan map[string][][]string, mainObjQueue chan map[string][]ObjectCommentInfo, headersIndex map[string]int) error {
	logger.Log.Debug("Starting acq=================================================")

	checkDuplicates := make(map[string]int)
	inconsitency := make(map[string]string)
	var goodObj, badObj [][]string
	mainObj := make(map[string][]ObjectCommentInfo)
	rows, err := fp.GetRows(acquiredRights)
	if err != nil {
		logger.Log.Error("Failed to read the sheet", zap.Error(err), zap.Any("sheet", acquiredRights))
		return errors.New("failedToReadAcquiredRightSheet")
	}

	for i := 1; i < len(rows); i++ {
		if checkDuplicates[strings.Join(rows[i], "|")] == 0 {
			checkDuplicates[strings.Join(rows[i], "|")] = i
			objects, ok, err := handleMandatoryFieldMissingAndWrongType(fp, i+1, acquiredRights, headersIndex)
			if err != nil {
				logger.Log.Error("handleMandatoryFieldMissingAndWrongType error", zap.Error(err))
				return err
			}
			if ok {
				obj, ok, err := handleAcquiredRightSheetInconsistency(fp, headersIndex, inconsitency, len(rows[i]), i+1)
				if err != nil {
					logger.Log.Error("handleAcquiredRightSheetInconsistency err", zap.Error(err))
					return err
				}
				if ok {
					goodObj = append(goodObj, rows[i])
				} else {
					badObj = append(badObj, rows[i])
					mainObj[acquiredRights] = append(mainObj[acquiredRights], obj...)
				}
			} else {
				badObj = append(badObj, rows[i])
				mainObj[acquiredRights] = append(mainObj[acquiredRights], objects...)
			}
		} else {
			mainObj[acquiredRights] = append(mainObj[acquiredRights], ObjectCommentInfo{
				Msg:         fmt.Sprintf("This row is duplicate with row no %d", checkDuplicates[strings.Join(rows[i], "|")]+1),
				Action:      DuplicateLine,
				Coordinates: []int{len(rows[i]), i + 1},
				IsFullRow:   true,
			})
			badObj = append(badObj, rows[i])
		}
	}
	logger.Log.Debug("Filtered acquiredRight msg objects ", zap.Any("goodObj", goodObj), zap.Any("baddObj", badObj), zap.Any("mainObj", mainObj))
	queueData := make(map[string][][]string)
	queueData[acquiredRights] = goodObj
	goodObjQueue <- queueData
	queueData = make(map[string][][]string)
	queueData[acquiredRights] = badObj
	badObjQueue <- queueData
	mainObjQueue <- mainObj
	logger.Log.Debug("end acq=================================================")
	return nil
}

func getSheetHeaderList(data map[string]int) *[]string {
	list := make([]string, len(data))
	for header, index := range data {
		list[(index-1)%len(data)] = header
	}
	return &list
}

func getFormatedRow(row []string, sheet string, list *[]string) (data []interface{}) {
	data = make([]interface{}, len(row))
	if list == nil {
		return
	}
	header := *list
	for i, v := range row {
		if i < len(header) {
			switch sheetsAndHeaders[sheet][header[i]].DataType {
			case INT:
				data[i] = 0
				data[i], _ = strconv.ParseInt(v, 10, 64)
			case FLOAT64:
				data[i] = 0.0
				data[i], _ = strconv.ParseFloat(v, 64)
			default:
				data[i] = fmt.Sprintf("%v", v)
			}
		}
	}
	return
}

func goodObjWriter(fileName, scope string, headersIndex map[string]map[string]int, goodObjQueue chan map[string][][]string) error {
	var isDataPresent bool
	gp := excel.NewFile()
	for {
		chLen := len(goodObjQueue)
		if chLen == len(sheetsAndHeaders) {
			close(goodObjQueue)
			break
		}
		time.Sleep(100 * time.Millisecond)
		logger.Log.Debug("Waiting for capture all good objects", zap.Any("goodObjectQueueLen", chLen))
	}
	sheetNum := 1
	for obj := range goodObjQueue {
		for sheet, objects := range obj {
			gp.SetActiveSheet(sheetNum)
			sheetNum++
			gp.NewSheet(sheet)
			if err := gp.SetSheetRow(sheet, "A1", getSheetHeaderList(headersIndex[sheet])); err != nil {
				logger.Log.Error("goodObjectWriter failed to add headers", zap.Any("sheet", sheet), zap.Error(err))
				return err
			}
			if sheet == servers {
				newColNum := len(headersIndex[sheet]) + 1
				colName, err := excel.ColumnNumberToName(newColNum)
				if err != nil {
					logger.Log.Error("Failed to Get new Column for server", zap.Error(err))
					return err
				}
				if err := gp.InsertCol(sheet, colName); err != nil {
					logger.Log.Error("Failed to insert new column for  corefactor  for server", zap.Error(err))
					return err
				}
				cell := fmt.Sprintf("%s1", colName)
				if err := gp.SetCellStr(servers, cell, "oracle_core_factor"); err != nil {
					logger.Log.Error("Failed to set new column value  corefactor  for server", zap.Error(err))
					return err
				}
			}
			counter := 1
			for row, val := range objects {
				counter++
				data := getFormatedRow(val, sheet, getSheetHeaderList(headersIndex[sheet]))
				if sheet == servers {
					hl := len(headersIndex[sheet])
					dl := len(data)
					if hl > dl {
						for i := 0; i < hl-dl; i++ {
							data = append(data, "")
						}
					}
					data = append(data, getCoreFactor(val[headersIndex[servers]["cpu_manufacturer"]-1], val[headersIndex[servers]["cpu_model"]-1]))

				}
				if err := gp.SetSheetRow(sheet, fmt.Sprintf("A%d", counter), &data); err != nil {
					logger.Log.Error("GoodObjectWriter failed to add row", zap.Any("sheet", sheet), zap.Any("row", row), zap.Any("value", val), zap.Error(err))
					return err
				}
				isDataPresent = true
			}
		}
	}
	gp.DeleteSheet("sheet1")
	file := fmt.Sprintf("%s/%s/analysis/good_%s", config.GetConfig().RawdataLocation, scope, fileName)
	if err := gp.SaveAs(file); err != nil {
		logger.Log.Error("goodObjectWriter failed to create sheet ", zap.Any("objectType", "good"), zap.Any("file", fileName), zap.Error(err))
		return err
	}
	if !isDataPresent {
		os.Remove(file)
	}
	logger.Log.Info("good Object file has been filtered", zap.Any("goodObjectFile", file))
	return nil
}

func badObjWriter(fileName, scope string, headersIndex map[string]map[string]int, badObjQueue chan map[string][][]string) error {
	var isDataPresent bool
	for {
		chLen := len(badObjQueue)
		if chLen == len(sheetsAndHeaders) {
			close(badObjQueue)
			break
		}
		time.Sleep(100 * time.Millisecond)
		logger.Log.Debug("Waiting for capture all bad objects", zap.Any("badObjectQueueLen", chLen))
	}
	bp := excel.NewFile()
	for obj := range badObjQueue {
		for sheet, objects := range obj {
			bp.NewSheet(sheet)
			if err := bp.SetSheetRow(sheet, "A1", getSheetHeaderList(headersIndex[sheet])); err != nil {
				logger.Log.Error("BadObjectWriter failed to add headers ", zap.Any("sheet", sheet), zap.Error(err))
				return err
			}
			counter := 1
			for row, val := range objects {
				counter++
				data := getFormatedRow(val, sheet, getSheetHeaderList(headersIndex[sheet]))
				if err := bp.SetSheetRow(sheet, fmt.Sprintf("A%d", counter), &data); err != nil {
					logger.Log.Error("BadObjectWriter failed to add row", zap.Any("sheet", sheet), zap.Any("row", row), zap.Any("value", val), zap.Error(err))
					return err
				}
				isDataPresent = true
			}
		}
	}

	bp.DeleteSheet("sheet1")
	if err := os.MkdirAll(fmt.Sprintf("%s/%s/errors", config.GetConfig().RawdataLocation, scope), os.ModePerm); err != nil {
		logger.Log.Error("Failed to create errors dir ", zap.String("scope", scope), zap.Error(err))
		return err
	}

	file := fmt.Sprintf("%s/%s/errors/bad_%s", config.GetConfig().RawdataLocation, scope, fileName)
	if err := bp.SaveAs(file); err != nil {
		logger.Log.Error("badObjectWriter failed to create sheet ", zap.Any("file", fileName), zap.Error(err))
		return err
	}
	if !isDataPresent {
		os.Remove(file)
	}
	logger.Log.Info("bad Object file has been filtered", zap.Any("badObjectFile", file))
	return nil
}

func mainObjWriter(fp *excel.File, scope, fileName string, mainObjQueue chan map[string][]ObjectCommentInfo, sheetSeq map[string]int) error {
	for {
		chLen := len(mainObjQueue)
		if chLen == len(sheetsAndHeaders) {
			close(mainObjQueue)
			break
		}
		time.Sleep(100 * time.Millisecond)
		logger.Log.Debug("Waiting for capture all onjects tp be commented", zap.Any("mainObjectQueueLen", chLen))
	}

	for obj := range mainObjQueue {
		for sheet, objects := range obj {
			fp.SetActiveSheet(sheetSeq[sheet])
			for _, val := range objects {
				var err error
				if val.IsFullRow {
					if len(val.Coordinates) != 2 {
						continue
					}
					err = addCellAnalysisByCordinates(fp, val.Coordinates[0], val.Coordinates[1], sheet, val.Msg, val.Action)
				} else if val.ColumnRanges != nil {
					for _, v := range val.ColumnRanges {
						err = addCellAnalysisByColumn(fp, v, sheet, val.Msg, val.Action)
					}
				} else {
					err = addCellAnalysisByColumn(fp, val.Column, sheet, val.Msg, val.Action)
				}
				if err != nil {
					logger.Log.Error("MainObjectWriter failed to add comment", zap.Any("object", val), zap.Error(err))
					return err
				}
				logger.Log.Debug("MainWriterObject", zap.Any("sheet", sheet), zap.Any("value", val))
			}
			file := fmt.Sprintf("%s/%s/analysis/%s", config.GetConfig().RawdataLocation, scope, fileName)
			if err := fp.SaveAs(file); err != nil {
				logger.Log.Error("MainObjectWriter failed to create sheet ", zap.Any("file", fileName), zap.Error(err))
				return err
			}
		}
	}
	return nil
}

func handleHeaders(fp *excel.File, sheetSeq map[string]int) (map[string]map[string]int, error) {
	incomingHeadersIndex := make(map[string]map[string]int)
	var isMissingHeader bool
	sheets := ""
	seq := 0
	sheetNum := 1
	for _, sheetName := range fp.GetSheetList() {
		fp.SetActiveSheet(sheetNum)
		sheetNum++
		sheetSeq[sheetName] = seq
		seq++
		if sheetsAndHeaders[sheetName] == nil { // dont process the other sheets
			continue
		}
		if incomingHeadersIndex[sheetName] == nil {
			incomingHeadersIndex[sheetName] = make(map[string]int)
		}
		rows, err := fp.GetRows(sheetName)
		if err != nil {
			logger.Log.Error("Failed to read the sheet", zap.Error(err), zap.Any("sheet", sheetName))
			return nil, fmt.Errorf("analysis:%s is %s", sheetName, BadSheet)
		}
		if len(rows) <= 1 {
			return nil, fmt.Errorf("analysis:%s sheet is empty", sheetName)
		}
		for i, val := range rows[0] {
			val = strings.ToLower(val)
			if incomingHeadersIndex[sheetName][val] == 0 {
				incomingHeadersIndex[sheetName][val] = i + 1
			} else if err := handleDuplicateHeader(fp, sheetName, val, i+1); err != nil {
				return nil, err
			}
		}
		for k, v := range sheetsAndHeaders[sheetName] {
			if (v.IsMandatory == Mandatory || v.IsMandatory == MandatoryWithBlank) && incomingHeadersIndex[sheetName][k] == 0 {
				logger.Log.Error("Mandatory header missing ", zap.Any("sheet", sheetName), zap.Any("header", k))
				isMissingHeader = true
				sheets += fmt.Sprintf("%s,", sheetName)
				break
			}
		}

	}
	if isMissingHeader {
		sheets = strings.TrimSuffix(sheets, ",")
		return nil, fmt.Errorf("analysis:Manadtory headers are missing in %s please check global temlplate file for more information", sheets)
	}
	return incomingHeadersIndex, nil
}

func handleDuplicateHeader(fp *excel.File, sheetName, headerName string, column int) error {
	col, err := excel.ColumnNumberToName(column)
	if err != nil {
		logger.Log.Error("failed to get col-name for duplicate header", zap.Error(err))
		return err
	}
	col = fmt.Sprintf("%s1", col)
	comment := fmt.Sprintf("This header %s is repeated", headerName)
	if err = addCellAnalysisByColumn(fp, col, sheetName, comment, DuplicateHeader); err != nil {
		return err
	}
	return nil
}

func handleSoftpartitionInconsistency(fp *excel.File, headersIndex map[string]int, inconsistency map[string]string, y int) (ObjectCommentInfo, bool) {
	key := getCellValue(fp, "softpartition_id", y, headersIndex, softpartitions)
	val := getCellValue(fp, "server_id", y, headersIndex, softpartitions)
	if inconsistency[key] != "" && val != "" && inconsistency[key] != val {
		colName, err := excel.ColumnNumberToName(headersIndex["softpartition_id"])
		if err != nil {
			logger.Log.Error("Failed to get server_id column namne", zap.Error(err))
			colName = "A"
		}
		return ObjectCommentInfo{
			Msg:    "same softpartition id cannot have multiple server_id",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", colName, y),
		}, false
	} else if key != "" && val != "" {
		inconsistency[key] = val
	}
	return ObjectCommentInfo{}, true
}

func handleServerSheetInconsistency(serverID string, data []string, inconsistency map[string]string, headersIndex map[string]int, y int) (ObjectCommentInfo, bool) {
	value := strings.Join(data, "|")
	if inconsistency[serverID] != "" {
		colName, err := excel.ColumnNumberToName(headersIndex["server_id"])
		if err != nil {
			logger.Log.Error("Failed to get server_id column namne", zap.Error(err))
			colName = "A"
		}
		return ObjectCommentInfo{
			Msg:    "ServerId is repeated with different configuration",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", colName, y+1),
		}, false
	} else if serverID != "" {
		inconsistency[serverID] = value
	}
	return ObjectCommentInfo{}, true
}

func handleProductSheetInconsistency(fp *excel.File, headersIndex map[string]int, inconsistency map[string]string, y int) (ObjectCommentInfo, bool) {
	val := getCellValue(fp, "application_name", y, headersIndex, products)
	key := getCellValue(fp, "application_id", y, headersIndex, products)
	if inconsistency[key] != "" && val != "" && inconsistency[key] != val {
		colName, err := excel.ColumnNumberToName(headersIndex["server_id"])
		if err != nil {
			logger.Log.Error("Failed to get server_id column namne", zap.Error(err))
			colName = "A"
		}
		return ObjectCommentInfo{
			Msg:    "Inconsistency, same application_id cannot have different name",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", colName, y+1),
		}, false

	}
	return ObjectCommentInfo{}, true
}

func getCellValue(fp *excel.File, key string, rowNo int, headersIndex map[string]int, sheetName string) (val string) {
	c1, err := excel.ColumnNumberToName(headersIndex[key])
	if err != nil {
		logger.Log.Error("Failed to get column-name", zap.Error(err), zap.Any("sheet", sheetName), zap.Any("key", key))
		return
	}
	c1 = fmt.Sprintf("%s%d", c1, rowNo)
	val, err = fp.GetCellValue(sheetName, c1)
	if err != nil {
		logger.Log.Error("Failed to get cell value", zap.Error(err), zap.Any("sheet", sheetName), zap.Any("cell", c1))
		return
	}
	return
}

func handleAcquiredRightSheetInconsistency(fp *excel.File, headersIndex map[string]int, inconsistency map[string]string, x, y int) ([]ObjectCommentInfo, bool, error) { // nolint
	sku := getCellValue(fp, "sku", y, headersIndex, acquiredRights)
	mp := getCellValue(fp, "maintenance_provider", y, headersIndex, acquiredRights)
	sp := getCellValue(fp, "software_provider", y, headersIndex, acquiredRights)
	csc := getCellValue(fp, "csc", y, headersIndex, acquiredRights)
	sn := getCellValue(fp, "support_number", y, headersIndex, acquiredRights)
	lpo := getCellValue(fp, "last_po", y, headersIndex, acquiredRights)
	var obj []ObjectCommentInfo
	if len(mp) > 16 {
		col, err := excel.ColumnNumberToName(headersIndex["maintenance_provider"])
		if err != nil {
			logger.Log.Error("Failed to get maintenance_provider colname in acqRights Sheet", zap.Any("colNumber", headersIndex["maintenance_provider"]), zap.Error(err))
			return nil, false, err
		}
		obj = append(obj, ObjectCommentInfo{
			Msg:    "maintenance_provider characters max limit is 16",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", col, y),
		})
	}
	if len(sn) > 16 {
		col, err := excel.ColumnNumberToName(headersIndex["support_number"])
		if err != nil {
			logger.Log.Error("Failed to get support_number colname in acqRights Sheet", zap.Any("colNumber", headersIndex["support_number"]), zap.Error(err))
			return nil, false, err
		}
		obj = append(obj, ObjectCommentInfo{
			Msg:    "support_number characters max limit is 16",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", col, y),
		})
	}
	if len(sp) > 16 {
		col, err := excel.ColumnNumberToName(headersIndex["software_provider"])
		if err != nil {
			logger.Log.Error("Failed to get software_provider colname in acqRights Sheet", zap.Any("colNumber", headersIndex["software_provider"]), zap.Error(err))
			return nil, false, err
		}
		obj = append(obj, ObjectCommentInfo{
			Msg:    "software_provider characters max limit is 16",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", col, y),
		})
	}
	if len(csc) > 16 {
		col, err := excel.ColumnNumberToName(headersIndex["csc"])
		if err != nil {
			logger.Log.Error("Failed to get csc colname in acqRights Sheet", zap.Any("colNumber", headersIndex["csc"]), zap.Error(err))
			return nil, false, err
		}
		obj = append(obj, ObjectCommentInfo{
			Msg:    "csc characters max limit is 16",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", col, y),
		})
	}
	if len(lpo) > 16 {
		col, err := excel.ColumnNumberToName(headersIndex["last_po"])
		if err != nil {
			logger.Log.Error("Failed to get last_po colname in acqRights Sheet", zap.Any("colNumber", headersIndex["last_po"]), zap.Error(err))
			return nil, false, err
		}
		obj = append(obj, ObjectCommentInfo{
			Msg:    "last_purchase_order characters max limit is 16",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", col, y),
		})
	}
	if strings.Contains(sku, "+") {
		col, err := excel.ColumnNumberToName(headersIndex["sku"])
		if err != nil {
			logger.Log.Error("Failed to get colname in acqRights Sheet", zap.Any("colNumber", headersIndex["sku"]), zap.Error(err))
			return nil, false, err
		}
		obj = append(obj, ObjectCommentInfo{
			Msg:    "Inconsistency,+  is not allowed in sku",
			Action: Inconsistent1,
			Column: fmt.Sprintf("%s%d", col, y),
		})
	}

	st := getCellValue(fp, "maintenance_start", y, headersIndex, acquiredRights)
	et := getCellValue(fp, "maintenance_end", y, headersIndex, acquiredRights)
	maintenanceLic := getCellValue(fp, "maintenance_licences", y, headersIndex, acquiredRights)
	msCol, err := excel.ColumnNumberToName(headersIndex["maintenance_start"])
	if err != nil {
		logger.Log.Error("Failed to get colname in acqRights Sheet", zap.Any("colNumber", headersIndex["maintenance_start"]), zap.Error(err))
		return nil, false, err
	}
	meCol, err := excel.ColumnNumberToName(headersIndex["maintenance_end"])
	if err != nil {
		logger.Log.Error("Failed to get colname in acqRights Sheet", zap.Any("colNumber", headersIndex["maintenance_end"]), zap.Error(err))
		return nil, false, err
	}
	if getMaintainenceLicNum(maintenanceLic) > 0 {
		if st != "" && et != "" {
			if !isMaintenanceDateOk(st, et) {
				obj = append(obj, ObjectCommentInfo{
					Msg:    "end of maintenance date must be greater than start date",
					Action: Inconsistent2,
					Column: fmt.Sprintf("%s%d", meCol, y),
				})
			}
		} else if st != "" {
			obj = append(obj, ObjectCommentInfo{
				Msg:    "End of maintenance date is mandatory with maintenance licenses",
				Action: Inconsistent2,
				Column: fmt.Sprintf("%s%d", meCol, y),
			})
		} else if et != "" {
			obj = append(obj, ObjectCommentInfo{
				Msg:    "Start of maintenance date is mandatory with maintenance licenses",
				Action: Inconsistent2,
				Column: fmt.Sprintf("%s%d", msCol, y),
			})
		} else {
			obj = append(obj, ObjectCommentInfo{
				Msg:          "start and end of maintenance date is mandatory with maintenance licenses",
				Action:       Inconsistent2,
				IsFullRow:    false,
				ColumnRanges: []string{fmt.Sprintf("%s%d", msCol, y), fmt.Sprintf("%s%d", meCol, y)},
			})
		}
	} else if st != "" && et != "" {
		obj = append(obj, ObjectCommentInfo{
			Msg:          "start and end of maintenance date is not considered as maintenance licences no is zero",
			Action:       Inconsistent2,
			IsFullRow:    false,
			ColumnRanges: []string{fmt.Sprintf("%s%d", msCol, y), fmt.Sprintf("%s%d", meCol, y)},
		})
	} else if st != "" {
		obj = append(obj, ObjectCommentInfo{
			Msg:    "End of maintenance date is not considered as maintenance licences no is zero",
			Action: Inconsistent2,
			Column: fmt.Sprintf("%s%d", msCol, y),
		})
	} else if et != "" {
		obj = append(obj, ObjectCommentInfo{
			Msg:    "start of maintenance date is not considered as maintenance licences no is zero",
			Action: Inconsistent2,
			Column: fmt.Sprintf("%s%d", meCol, y),
		})
	}
	if len(obj) > 0 {
		return obj, false, nil
	}
	return nil, true, nil
}

func isMaintenanceDateOk(st, et string) bool {
	st = strings.ReplaceAll(st, "/", "-")
	startTime, err := time.Parse("02-01-2006", st)
	if err != nil {
		logger.Log.Error("Failed to get start  time", zap.Error(err))
		return false
	}
	et = strings.ReplaceAll(et, "/", "-")
	endTime, err := time.Parse("02-01-2006", et)
	if err != nil {
		logger.Log.Error("Failed to get endt time", zap.Error(err))
		return false
	}
	return endTime.After(startTime) && endTime.After(time.Now())
}

func getMaintainenceLicNum(data string) int {
	if data != "" {
		x, y := strconv.ParseInt(data, 10, 64)
		if y != nil {
			logger.Log.Error("Failed to get maintenance licence numbers", zap.Error(y))
			return -1
		}
		return int(x)
	}
	return 0
}

func addCellAnalysisByColumn(fp *excel.File, col, sheetName, comment, action string) error {
	colour := actionAndColors[action]
	logger.Log.Debug("cellStyleInfo", zap.Any("col", col), zap.Any("sheet", sheetName), zap.Any("cmt", comment), zap.Any("action", action), zap.Any("color", colour))
	style, err := fp.NewStyle(`{"fill":{"type":"pattern","color":["` + colour + `"],"pattern":1}}`)
	if err != nil {
		logger.Log.Error("Failed to create style ", zap.Error(err), zap.Any("action", action))
		return err
	}
	if err = fp.AddComment(sheetName, col, `{"author":"OPTISAM: ","text":"`+comment+`"}`); err != nil {
		logger.Log.Error("Failed to add comment", zap.Error(err), zap.Any("action", action))
		return err
	}
	if err = fp.SetCellStyle(sheetName, col, col, style); err != nil {
		logger.Log.Error("Failed to add cell style", zap.Error(err), zap.Any("sheet", sheetName), zap.Any("action", action))
		return err
	}
	logger.Log.Debug("", zap.Any("cell", col), zap.Any("sheet", sheetName), zap.Any("action", action))
	return nil
}

func addCellAnalysisByCordinates(fp *excel.File, x, y int, sheetName, comment, action string) error {
	colour := actionAndColors[action]
	logger.Log.Debug("cellStyleInfo", zap.Any("x", x), zap.Any("y", y), zap.Any("sheet", sheetName), zap.Any("cmt", comment), zap.Any("action", action), zap.Any("color", colour))
	cell, err := excel.CoordinatesToCellName(x, y)
	if err != nil {
		logger.Log.Error("Failed to get column name", zap.Error(err), zap.Any("sheet", sheetName), zap.Any("action", action))
		return err
	}
	style, err := fp.NewStyle(`{"fill":{"type":"pattern","color":["` + colour + `"],"pattern":1}}`)
	if err != nil {
		logger.Log.Error("Failed to create style ", zap.Error(err), zap.Any("action", action))
		return err
	}
	if err = fp.AddComment(sheetName, cell, `{"author":"OPTISAM: ","text":"`+comment+`"}`); err != nil {
		logger.Log.Error("Failed to add comment", zap.Error(err), zap.Any("action", action))
		return err
	}
	col1, col2 := cell, cell
	if action == DuplicateLine || action == Inconsistent1 || action == Inconsistent2 || action == BadReference {
		col1 = fmt.Sprintf("A%d", y)
		col2, _ = excel.ColumnNumberToName(x)
		col2 = fmt.Sprintf("%s%d", col2, y)
	}
	if err = fp.SetCellStyle(sheetName, col1, col2, style); err != nil {
		logger.Log.Error("Failed to add cell style", zap.Error(err), zap.Any("sheet", sheetName), zap.Any("action", action))
		return err
	}
	logger.Log.Debug("", zap.Any("cell", cell), zap.Any("sheet", sheetName), zap.Any("action", action))
	return nil
}

func isTypeMatched(data string, expectedType int) bool {
	if data != "" {
		switch expectedType {
		case INT:
			if _, err := strconv.ParseInt(data, 10, 64); err != nil {
				return false
			}
		case FLOAT64:
			if _, err := strconv.ParseFloat(data, 64); err != nil {
				return false
			}
		case DATE:
			if strings.Contains(data, "/") {
				data = strings.ReplaceAll(data, "/", "-")
			}
			_, err := time.Parse("02-01-2006", data)
			if err != nil {
				return false
			}
		}
	}
	return true
}

func handleMandatoryFieldMissingAndWrongType(fp *excel.File, yLen int, sheetName string, header map[string]int) ([]ObjectCommentInfo, bool, error) {
	var Objects []ObjectCommentInfo
	isGoodObject := true
	for k, v := range header {
		var msg, action, column string
		colName, err := excel.ColumnNumberToName(v)
		if err != nil {
			logger.Log.Error("Failed to get column-name", zap.Error(err), zap.Any("sheet", sheetName), zap.Any("action", "MissingFieldAndWrongType"))
			return nil, isGoodObject, err
		}
		cell := fmt.Sprintf("%s%d", colName, yLen)
		celVal, err := fp.GetCellValue(sheetName, cell)
		if err != nil {
			logger.Log.Error("Failed to get cell value", zap.Error(err), zap.Any("sheet", sheetName), zap.Any("cell", cell))
			return nil, isGoodObject, err
		}
		if sheetsAndHeaders[sheetName][k].IsMandatory == Mandatory && celVal == "" { // missing mandatory field
			msg = "This mandatory value is missing."
			action = MissingField
			column = cell
		} else if !isTypeMatched(celVal, sheetsAndHeaders[sheetName][k].DataType) { // wrong type
			msg = `This value is wrongType, expected :` + dataTypes[sheetsAndHeaders[sheetName][k].DataType] + `.`
			action = WrongTypeField
			column = cell
			if sheetName == acquiredRights && (k == "maintenance_start" || k == "maintenance_end") {
				if err := fp.SetCellValue(sheetName, cell, celVal); err != nil {
					return nil, isGoodObject, err
				}
			}
		}
		if msg != "" {
			Objects = append(Objects, ObjectCommentInfo{Msg: msg, Action: action, Column: column})
			isGoodObject = false
		}
	}
	return Objects, isGoodObject, nil
}
