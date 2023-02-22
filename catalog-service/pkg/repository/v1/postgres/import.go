package postgres

import (
	"context"
	"database/sql"
	"fmt"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type productdata struct {
	name                string
	editor_name         string
	is_closesource      bool
	is_opensource       bool
	genearl_information string
	created_on          time.Time
	updated_on          time.Time
	location            db.LocationType
	opensource_type     db.OpensourceType
	productUid          string
}
type versiondata struct {
	versionUid    string
	productName   string
	editorName    string
	vname         string
	eol           *timestamppb.Timestamp
	eos           *timestamppb.Timestamp
	swidTagSystem string
}

func (p *ProductCatalogRepository) InsertRecordsTx(ctx context.Context, req *v1.UploadRecords) (msg string, err error) {
	if req.Data == nil {
		logger.Log.Error("v1/service - UpdateProduct - data is empty")
		return "data is empty", status.Error(codes.Internal, "data is empty")
	}
	product_editors := map[string]productdata{}
	version_product_editors := map[string]versiondata{}
	valueStrings := []string{}
	valueArgs := []interface{}{}
	editorQuery := fmt.Sprintf("INSERT INTO editor_catalog (id,name,created_on,updated_on) VALUES %s ON CONFLICT (LOWER(name)) DO NOTHING;", strings.Join(valueStrings, ","))

	var counter int
	start := time.Now()
	logger.Log.Info("v1/service - Bulk Import Started")
	for _, record := range req.GetData() {
		editorName := strings.Trim(record.Editor, " ")
		productName := strings.Trim(record.Name, " ")

		if editorName == "" || productName == "" {
			continue
		}
		if counter == 5000 {
			editorQuery = fmt.Sprintf("INSERT INTO editor_catalog (id,name,created_on,updated_on) VALUES %s ON CONFLICT (LOWER(name)) DO NOTHING;", strings.Join(valueStrings, ","))
			_, err := p.db.Exec(editorQuery, valueArgs...)
			if err != nil {
				logger.Log.Error("v1/service - InsertRecordsTx - query error" + err.Error())
				return "Internal Server Error: Unable to Process file records ", err
			}
			valueArgs = []interface{}{}
			valueStrings = []string{}

			prodvalueStrings, prodvalueArgs := generateQueryForProducts(product_editors)
			prodQuery := fmt.Sprintf("INSERT INTO product_catalog (id,name,editorID,editor_name,is_closesource,is_opensource,genearl_information,location,created_on,updated_on,opensource_type) values %s ON CONFLICT (LOWER(name),LOWER(editor_name)) Do UPDATE SET is_closesource = EXCLUDED.is_closesource, is_opensource = EXCLUDED.is_opensource, genearl_information = EXCLUDED.genearl_information, updated_on = EXCLUDED.updated_on", strings.Join(prodvalueStrings, ","))
			_, err = p.db.Exec(prodQuery, prodvalueArgs...)
			if err != nil {
				logger.Log.Error("v1/service - InsertRecordsTx - product query error" + err.Error())
				return "Internal Server Error: Unable to Process file records ", err
			}
			prodvalueStrings = []string{}
			prodvalueArgs = []interface{}{}
			vervalueStrings, vervalueArgs := generateQueryForVersions(version_product_editors)
			verQuery := fmt.Sprintf("INSERT INTO version_catalog (id,p_id,name,end_of_life,end_of_support,swid_tag_system) VALUES %s on CONFLICT (LOWER(name),p_id) Do UPDATE SET end_of_life = EXCLUDED.end_of_life ,end_of_support = EXCLUDED.end_of_support", strings.Join(vervalueStrings, ","))
			_, err = p.db.Exec(verQuery, vervalueArgs...)
			if err != nil {
				logger.Log.Error("v1/service - InsertRecordsTx - query error" + err.Error())
				return "Internal Server Error: Unable to Process file records ", err
			}
			vervalueStrings = []string{}
			vervalueArgs = []interface{}{}
			product_editors = map[string]productdata{}
			version_product_editors = map[string]versiondata{}
			counter = 0
			logger.Log.Info("v1/service - Executed Batch of Records in" + fmt.Sprint(time.Since(start)))
		}
		currentTimeStamp := time.Now()
		editorUid := uuid.New().String()
		numFields := 4
		n := counter * numFields
		valueStrings = append(valueStrings, "($"+strconv.Itoa(n+1)+",$"+strconv.Itoa(n+2)+",$"+strconv.Itoa(n+3)+",$"+strconv.Itoa(n+4)+")")
		valueArgs = append(valueArgs, editorUid)
		valueArgs = append(valueArgs, editorName)
		valueArgs = append(valueArgs, currentTimeStamp)
		valueArgs = append(valueArgs, currentTimeStamp)

		//prod logic
		productUid := uuid.New().String()
		genInfo := record.GenearlInformation
		var isOpenSource, isClosedSource bool
		if strings.ToLower(record.Licensing) == "open source" {
			isOpenSource = true
		}
		if strings.ToLower(record.Licensing) == "closed source" {
			isClosedSource = true
		}

		if strings.ToLower(record.Licensing) == "open source and closed source" || strings.ToLower(record.Licensing) == "closed source and open source" {
			isClosedSource = true
			isOpenSource = true
		}

		product_editors[strings.ToLower(editorName)+strings.ToLower(productName)] = productdata{
			name:                productName,
			editor_name:         editorName,
			is_closesource:      isClosedSource,
			is_opensource:       isOpenSource,
			genearl_information: genInfo,
			location:            db.LocationType("NONE"),
			opensource_type:     db.OpensourceType("NONE"),
			productUid:          productUid,
			created_on:          currentTimeStamp,
			updated_on:          currentTimeStamp,
		}

		//versionlogic
		versionUid := uuid.New().String()
		vname := strings.Trim(record.Version, " ")
		var swidTagSystem string
		if vname == "" {
			swidTagSystem = strings.ReplaceAll(strings.Join([]string{productName, editorName}, "_"), " ", "_")
		} else {
			swidTagSystem = strings.ReplaceAll(strings.Join([]string{productName, editorName, vname}, "_"), " ", "_")
		}
		eol := record.EndOfLife
		eos := record.EndOfSupport

		version_product_editors[strings.ToLower(editorName)+strings.ToLower(productName)+strings.ToLower(vname)] = versiondata{
			versionUid:    versionUid,
			productName:   productName,
			editorName:    editorName,
			vname:         vname,
			eol:           eol,
			eos:           eos,
			swidTagSystem: swidTagSystem,
		}
		counter++
	}

	editorQuery = fmt.Sprintf("INSERT INTO editor_catalog (id,name,created_on,updated_on) VALUES %s ON CONFLICT (LOWER(name)) DO NOTHING;", strings.Join(valueStrings, ","))
	_, err = p.db.Exec(editorQuery, valueArgs...)
	if err != nil {
		logger.Log.Error("v1/service - InsertRecordsTx - query error" + err.Error())
		return "Internal Server Error: Unable to Process file records ", err
	}

	prodvalueStrings, prodvalueArgs := generateQueryForProducts(product_editors)
	prodQuery := fmt.Sprintf("INSERT INTO product_catalog (id,name,editorID,editor_name,is_closesource,is_opensource,genearl_information,location,created_on,updated_on,opensource_type) values %s ON CONFLICT (LOWER(name),LOWER(editor_name)) Do UPDATE SET is_closesource = EXCLUDED.is_closesource, is_opensource = EXCLUDED.is_opensource, genearl_information = EXCLUDED.genearl_information, updated_on = EXCLUDED.updated_on", strings.Join(prodvalueStrings, ","))
	// fmt.Println(prodQuery)
	_, err = p.db.Exec(prodQuery, prodvalueArgs...)
	if err != nil {
		logger.Log.Error("v1/service - InsertRecordsTx - prod query error" + err.Error())
		return "Internal Server Error: Unable to Process file records ", err
	}
	vervalueStrings, vervalueArgs := generateQueryForVersions(version_product_editors)
	verQuery := fmt.Sprintf("INSERT INTO version_catalog (id,p_id,name,end_of_life,end_of_support,swid_tag_system) VALUES %s on CONFLICT (LOWER(name),p_id) Do UPDATE SET end_of_life = EXCLUDED.end_of_life ,end_of_support = EXCLUDED.end_of_support", strings.Join(vervalueStrings, ","))
	_, err = p.db.Exec(verQuery, vervalueArgs...)
	// fmt.Println(verQuery)

	if err != nil {
		logger.Log.Error("v1/service - InsertRecordsTx - version query error" + err.Error())
		return "Internal Server Error: Unable to Process file records ", err
	}
	logger.Log.Info("v1/service - Executed all Records in" + fmt.Sprint(time.Since(start)))
	if err == nil {
		logger.Log.Info("File Updated Successfully")
		return fmt.Sprintf("File Updated Successfully"), err
	} else {
		logger.Log.Error(err.Error())
		return fmt.Sprintf("File did not Updated Successfully"), err
	}
}

func generateQueryForProducts(product_editors map[string]productdata) (prodvalueStrings []string, prodvalueArgs []interface{}) {
	// fmt.Printf("\n%+v\n", product_editors)
	pnumFields := 11
	var counter int
	for _, v := range product_editors {
		pn := counter * pnumFields
		prodvalueStrings = append(prodvalueStrings, "($"+strconv.Itoa(pn+1)+",$"+strconv.Itoa(pn+2)+",(select id from editor_catalog where lower(name) = "+"$"+strconv.Itoa(pn+3)+" LIMIT 1 OFFSET 0) ,$"+strconv.Itoa(pn+4)+",$"+strconv.Itoa(pn+5)+",$"+strconv.Itoa(pn+6)+",$"+strconv.Itoa(pn+7)+",$"+strconv.Itoa(pn+8)+",$"+strconv.Itoa(pn+9)+",$"+strconv.Itoa(pn+10)+",$"+strconv.Itoa(pn+11)+")")

		prodvalueArgs = append(prodvalueArgs, v.productUid)
		prodvalueArgs = append(prodvalueArgs, v.name)
		prodvalueArgs = append(prodvalueArgs, strings.ToLower(v.editor_name))
		prodvalueArgs = append(prodvalueArgs, v.editor_name)
		prodvalueArgs = append(prodvalueArgs, v.is_closesource)
		prodvalueArgs = append(prodvalueArgs, v.is_opensource)
		prodvalueArgs = append(prodvalueArgs, string(v.genearl_information))
		prodvalueArgs = append(prodvalueArgs, v.location)
		prodvalueArgs = append(prodvalueArgs, v.created_on)
		prodvalueArgs = append(prodvalueArgs, v.updated_on)
		prodvalueArgs = append(prodvalueArgs, v.opensource_type)
		counter++
	}
	return
}

func generateQueryForVersions(version_product_editors map[string]versiondata) (vervalueStrings []string,
	vervalueArgs []interface{}) {
	vnumFields := 7
	var counter int
	for _, v := range version_product_editors {
		vn := counter * vnumFields
		vervalueStrings = append(vervalueStrings, "($"+strconv.Itoa(vn+1)+",(select id from product_catalog where lower(name) =  "+"$"+strconv.Itoa(vn+2)+" AND lower(editor_name) =  "+"$"+strconv.Itoa(vn+3)+"  LIMIT 1 OFFSET 0) ,$"+strconv.Itoa(vn+4)+",$"+strconv.Itoa(vn+5)+",$"+strconv.Itoa(vn+6)+",$"+strconv.Itoa(vn+7)+")")

		vervalueArgs = append(vervalueArgs, v.versionUid)
		vervalueArgs = append(vervalueArgs, strings.ToLower(v.productName))
		vervalueArgs = append(vervalueArgs, strings.ToLower(v.editorName))
		vervalueArgs = append(vervalueArgs, v.vname)
		vervalueArgs = append(vervalueArgs, sql.NullTime{Time: v.eol.AsTime(), Valid: true})
		vervalueArgs = append(vervalueArgs, sql.NullTime{Time: v.eos.AsTime(), Valid: true})
		vervalueArgs = append(vervalueArgs, v.swidTagSystem)
		counter++
	}
	return
}
