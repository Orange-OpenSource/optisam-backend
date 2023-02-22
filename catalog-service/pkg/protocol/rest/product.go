package rest

import (
	"database/sql"
	"encoding/json"
	"net/http"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/middleware/grpc"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const getProduct = `-- name: GetProduct :one
SELECT id, name, editorid, genearl_information, contract_tips, support_vendors, metrics, is_opensource, licences_opensource, is_closesource, licenses_closesource, location, created_on, updated_on, recommendation, useful_links,swid_tag_product,	opensource_type,editor_name,(Select json_agg(t.scope) as a from (
    (Select scope from products where product_editor = product_catalog.editor_name AND product_name = product_catalog.name )
    UNION
    (select scope from acqrights where product_editor = product_catalog.editor_name AND product_name = product_catalog.name)
) t) 
 from product_catalog 
WHERE id = $1`

const getVersionByPrductID = `-- name: GetVersionByPrductID :many
SELECT id, p_id, name, end_of_life, end_of_support, recommendation,swid_tag_version from version_catalog 
WHERE p_id = $1 AND name != ''
`

func (handler *handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var productResponse Product
	if r.Method != http.MethodGet {
		logger.Log.Error("rest - geteditor - Meathod Not Allowed", zap.String("Reason: ", "Meathod Not Allowed"))
		sendErrorResponse(http.StatusMethodNotAllowed, "Meathod Not Allowed", w)
		return
	}

	id := (r.URL.Query().Get("id"))
	if id == "" {
		logger.Log.Error("rest - geteditor - QueryParams", zap.String("Reason: ", "Unable to fetch Query Params"))
		sendErrorResponse(http.StatusBadRequest, "Unable to fetch Query Params", w)
		return
	}
	var scopes []byte
	var support_vendors, metrics, closeLicences, usefulLinks []byte
	var swidTagProduct, openLicences sql.NullString
	row := handler.Db.QueryRowContext(r.Context(), getProduct, id)
	err := row.Scan(
		&productResponse.ID,
		&productResponse.Name,
		&productResponse.Editorid,
		&productResponse.GenearlInformation,
		&productResponse.ContractTips,
		&support_vendors,
		&metrics,
		&productResponse.IsOpensource,
		&openLicences,
		&productResponse.IsClosesource,
		&closeLicences,
		&productResponse.Location,
		&productResponse.CreatedOn,
		&productResponse.UpdatedOn,
		&productResponse.Recommendation,
		&usefulLinks,
		&swidTagProduct,
		&productResponse.OpensourceType,
		&productResponse.EditorName,
		&scopes,
	)
	if err != nil {
		logger.Log.Error("rest- GetProduct - GetProduct", zap.String("reason", err.Error()))
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}
	if len(support_vendors) != 0 {
		productResponse.SupportVendors = support_vendors
	} else {
		str := make([]string, 1)
		s, _ := json.Marshal(str)
		productResponse.SupportVendors = s
	}
	if len(metrics) != 0 {
		productResponse.Metrics = metrics
	} else {
		str := make([]string, 1)
		s, _ := json.Marshal(str)
		productResponse.Metrics = s
	}
	if len(closeLicences) != 0 {
		productResponse.LicensesClosesource = closeLicences
	} else {
		str := make([]string, 1)
		s, _ := json.Marshal(str)
		productResponse.LicensesClosesource = s
	}
	if len(usefulLinks) != 0 {
		productResponse.UsefulLinks = usefulLinks
	} else {
		str := make([]string, 1)
		s, _ := json.Marshal(str)
		productResponse.UsefulLinks = s
	}
	productResponse.SwidTagProduct = swidTagProduct
	productResponse.LicencesOpensource = openLicences

	//for response
	var responseObject ProductResponse
	responseObject.Version = make([]*Version, 0)
	responseObject.CloseSource = new(CloseSource)
	responseObject.OpenSource = new(OpenSource)

	responseObject.Id = productResponse.ID
	responseObject.EditorID = productResponse.Editorid
	responseObject.EditorName = productResponse.EditorName
	responseObject.Name = productResponse.Name
	responseObject.CloseSource.IsCloseSource = productResponse.IsClosesource.Bool
	jsonStr, err := json.Marshal(productResponse.LicensesClosesource)
	if err != nil {
		logger.Log.Error("service/v1 - GetProduct - Marshal", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
	}
	json.Unmarshal(jsonStr, &responseObject.CloseSource.CloseLicences)

	responseObject.OpenSource.IsOpenSource = productResponse.IsOpensource.Bool
	responseObject.OpenSource.OpenLicences = productResponse.LicencesOpensource.String
	responseObject.OpenSource.OpensourceType = string(productResponse.OpensourceType)
	responseObject.ContracttTips = productResponse.ContractTips.String
	responseObject.GenearlInformation = productResponse.GenearlInformation.String
	responseObject.Recommendation = productResponse.Recommendation.String
	responseObject.LocationType = string(productResponse.Location)
	responseObject.ProductSwidTag = productResponse.SwidTagProduct.String
	if scopes != nil {
		var scopeCode []string
		err = json.Unmarshal(scopes, &scopeCode)
		if err != nil {
			logger.Log.Error("rest - product - Marshal", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
			return
		}
		cronCtx, err := CreateSharedContext(handler.AuthAPI, handler.Application)
		if err != nil {
			logger.Log.Error("couldnt fetch token, will try next time when API will execute", zap.Any("error", err))
		}
		if cronCtx != nil {
			cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, handler.VerifyKey, handler.APIKey)
			if err != nil {
				logger.Log.Error("Cron AddClaims Failed", zap.Error(err))
			}

			listScopesNames, err := handler.account.GetScopeLists(cronAPIKeyCtx, &accv1.GetScopeListRequest{Scopes: scopeCode})
			if err != nil {
				logger.Log.Error("couldnt fetch scope name, will try next time when API will execute", zap.Any("error", err))
			}
			responseObject.Scopes = listScopesNames.ScopeNames
		}
	}
	jsonStr, err = json.Marshal(productResponse.Metrics)
	if err != nil {
		logger.Log.Error("service/v1 - GetProduct - Marshal", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
	}
	json.Unmarshal(jsonStr, &responseObject.Metrics)

	jsonStr, err = json.Marshal(productResponse.SupportVendors)
	if err != nil {
		logger.Log.Error("service/v1 - GetProduct - Marshal", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
	}
	json.Unmarshal(jsonStr, &responseObject.SupportVendors)

	jsonStr, err = json.Marshal(productResponse.UsefulLinks)
	if err != nil {
		logger.Log.Error("service/v1 - GetProduct - Marshal", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
	}
	json.Unmarshal(jsonStr, &responseObject.UsefulLinks)

	// createdOnObject, _ := ptypes.TimestampProto(productResponse.CreatedOn)
	responseObject.CreatedOn = productResponse.CreatedOn

	// updatedOnObject, _ := ptypes.TimestampProto(productResponse.UpdatedOn)
	responseObject.UpdatedOn = productResponse.UpdatedOn

	rows, _ := handler.Db.QueryContext(r.Context(), getVersionByPrductID, id)
	defer rows.Close()
	var items []db.VersionCatalog
	for rows.Next() {
		var i db.VersionCatalog
		if err := rows.Scan(
			&i.ID,
			&i.PID,
			&i.Name,
			&i.EndOfLife,
			&i.EndOfSupport,
			&i.Recommendation,
			&i.SwidTagVersion,
		); err != nil {
			sendErrorResponse(http.StatusBadRequest, err.Error(), w)

		}
		items = append(items, i)
	}

	for _, version := range items {
		var ver Version
		eos := version.EndOfSupport.Time
		eol := version.EndOfLife.Time
		if version.EndOfLife.Time.String() == "1970-01-01 00:00:00 +0000 +0000" || version.EndOfLife.Time.String() == "0001-01-01 00:00:00 +0000 +0000" {
			ver.EndOfLife = nil
		} else {
			// pbtime, _ := ptypes.TimestampProto(ver.EndOfLife.Time)
			ver.EndOfLife = &eol
		}
		if version.EndOfSupport.Time.String() == "1970-01-01 00:00:00 +0000 +0000" || version.EndOfSupport.Time.String() == "0001-01-01 00:00:00 +0000 +0000" {
			ver.EndOfSupport = nil
		} else {
			ver.EndOfSupport = &eos
		}

		ver.SwidTagVersion = version.SwidTagVersion.String
		ver.Id = version.ID
		ver.Name = version.Name
		ver.Recommendation = version.Recommendation.String
		responseObject.Version = append(responseObject.Version, &ver)
	}
	e, err := json.Marshal(responseObject)
	if err != nil {
		logger.Log.Error("rest - GetProduct - Marshal", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}
	w.Write([]byte(e))
}

// const listProducts = `
// SELECT count(*) OVER() AS totalRecords,pc.id, pc.name, pc.editorid, pc.genearl_information, pc.contract_tips, pc.support_vendors, pc.metrics, pc.is_opensource, pc.licences_opensource, pc.is_closesource, pc.licenses_closesource, pc.location, pc.created_on, pc.updated_on, pc.recommendation, pc.useful_links ,pc.swid_tag_product,pc.editor_name,pc.opensource_type,json_agg(v.*)  as ve,(Select json_agg(t.scope) as a from (
//     (Select scope from products where product_editor = pc.editor_name AND product_name = pc.name )
//     UNION
//     (select scope from acqrights where product_editor = pc.editor_name AND product_name = pc.name)
// ) t)
// from product_catalog pc
// left outer join version_catalog v on v.p_id=pc.id and v.name !=''
// WHERE
//  (CASE WHEN $1::BOOL THEN lower(pc.name) LIKE '%' || lower($2::TEXT) || '%' ELSE TRUE END)
//  AND
//  (CASE WHEN $3::BOOL THEN lower(editor_name) LIKE '%' || lower($4::TEXT) || '%' ELSE TRUE END)
//  AND
//  (CASE WHEN $23::BOOL THEN editorid = $24::TEXT ELSE TRUE END)
//  AND
//  (CASE WHEN $5::BOOL THEN location::TEXT = $6  ELSE TRUE END)
//  AND
//  (CASE WHEN ($7::BOOL AND $8::BOOL) THEN is_opensource = TRUE AND is_closesource = FALSE ELSE TRUE END)
//  AND
//  (CASE WHEN ($7::BOOL AND $9::BOOL) THEN is_closesource = TRUE AND is_opensource = FALSE ELSE TRUE END)
//  AND
//  (CASE WHEN ($7::BOOL AND $10::BOOL) THEN is_closesource = TRUE AND is_opensource = TRUE ELSE TRUE END)
// GROUP BY pc.id
// ORDER BY
// CASE WHEN $11::BOOL THEN created_on END asc,
// CASE WHEN $12::BOOL THEN created_on END desc,
// CASE WHEN $13::BOOL THEN pc.name END asc,
// CASE WHEN $14::BOOL THEN pc.name END desc,
// CASE WHEN $15::BOOL THEN is_closesource END asc,
// CASE WHEN $16::BOOL THEN is_closesource END desc,
// CASE WHEN $17::BOOL THEN location END asc,
// CASE WHEN $18::BOOL THEN location END desc,
// CASE WHEN $19::BOOL THEN editor_name END asc,
// CASE WHEN $20::BOOL THEN editor_name END desc
// OFFSET $21 LIMIT $22;
// `

const listProducts = `
select * ,(Select json_agg(t.scope) as a from (
    (Select scope from products where product_editor = product.editor_name AND product_name = product.name )
    UNION
    (select scope from acqrights where product_editor = product.editor_name AND product_name = product.name)
) t) from (
SELECT count(*) OVER() AS totalRecords,pc.id, pc.name, pc.editorid, 
pc.genearl_information, pc.contract_tips, pc.support_vendors, pc.metrics,
 pc.is_opensource, pc.licences_opensource, pc.is_closesource, pc.licenses_closesource,
pc.location, pc.created_on, pc.updated_on, pc.recommendation, pc.useful_links ,
pc.swid_tag_product,pc.editor_name,pc.opensource_type,json_agg(v.*)  as ve
from product_catalog pc
left outer join version_catalog v on v.p_id=pc.id and v.name !=''
WHERE
 (CASE WHEN $1::BOOL THEN lower(pc.name) LIKE '%' || lower($2::TEXT) || '%' ELSE TRUE END)
 AND
 (CASE WHEN $3::BOOL THEN lower(editor_name) LIKE '%' || lower($4::TEXT) || '%' ELSE TRUE END)
 AND
 (CASE WHEN $23::BOOL THEN editorid = $24::TEXT ELSE TRUE END)
 AND
 (CASE WHEN $5::BOOL THEN location::TEXT = $6  ELSE TRUE END)
 AND
 (CASE WHEN ($7::BOOL AND $8::BOOL) THEN is_opensource = TRUE AND is_closesource = FALSE ELSE TRUE END)
 AND
 (CASE WHEN ($7::BOOL AND $9::BOOL) THEN is_closesource = TRUE AND is_opensource = FALSE ELSE TRUE END)
 AND
 (CASE WHEN ($7::BOOL AND $10::BOOL) THEN is_closesource = TRUE AND is_opensource = TRUE ELSE TRUE END)
GROUP BY pc.id 
ORDER BY
CASE WHEN $11::BOOL THEN created_on END asc,
CASE WHEN $12::BOOL THEN created_on END desc,
CASE WHEN $13::BOOL THEN pc.name END asc,
CASE WHEN $14::BOOL THEN pc.name END desc,
CASE WHEN $15::BOOL THEN is_closesource END asc,
CASE WHEN $16::BOOL THEN is_closesource END desc,
CASE WHEN $17::BOOL THEN location END asc,
CASE WHEN $18::BOOL THEN location END desc,
CASE WHEN $19::BOOL THEN editor_name END asc,
CASE WHEN $20::BOOL THEN editor_name END desc
OFFSET $21 LIMIT $22 ) as product
`

// List Products
func (handler *handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("GetProducts", zap.Any("GetProducts called", time.Now()))
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		logger.Log.Error("rest - geteditor - Meathod Not Allowed", zap.String("Reason: ", "Meathod Not Allowed"))
		sendErrorResponse(http.StatusMethodNotAllowed, "Meathod Not Allowed", w)
		return
	}
	search_params_name := (r.URL.Query().Get("search_params.name.filteringkey"))
	ispname := (search_params_name != "")
	pname := strings.ToLower(search_params_name)

	//Listing response

	var productsDbResponse []ProductsDBResponse
	apiresp := ListProductResponse{}
	search_params_editorName := (r.URL.Query().Get("search_params.editorName.filteringkey"))
	iseditorname := (search_params_editorName != "")
	pditorname := strings.ToLower(search_params_editorName)
	pditorId := (r.URL.Query().Get("search_params.editorId.filteringkey"))
	iseditorId := (pditorId != "")
	search_params_location := (r.URL.Query().Get("search_params.locationType.filteringkey"))
	islocation := (search_params_location != "")
	plocation := search_params_location

	if strings.ToUpper(search_params_location) == string(db.LocationTypeSAAS) {
		plocation = "SAAS"
	} else if strings.ToUpper(search_params_location) == strings.ToUpper(string(db.LocationTypeOnPremise)) {
		plocation = "On Premise"
	} else if strings.ToUpper(search_params_location) == string(db.LocationTypeNONE) {
		plocation = "NONE"
	} else if strings.ToUpper(search_params_location) == strings.ToUpper(string(db.LocationTypeBoth)) {
		plocation = "Both"
	}

	var is_closesource, is_opensource, both bool
	search_params_licensing := (r.URL.Query().Get("search_params.licensing.filteringkey"))
	islicensing := (search_params_licensing != "")
	if islicensing {
		if strings.ToLower(search_params_licensing) == "open source" {
			is_opensource = true
		} else if strings.ToLower(search_params_licensing) == "closed source" {
			is_closesource = true
		} else if strings.ToLower(search_params_licensing) == "open source and closed source" {
			both = true
		} else {

			apiresp.Product = make([]*ProductResponse, len(productsDbResponse))
			e, _ := json.Marshal(apiresp)
			w.Write([]byte(e))
			return
		}
	}

	sort_order := (r.URL.Query().Get("sort_order"))
	sort_by := (r.URL.Query().Get("sort_by"))
	p_n := (r.URL.Query().Get("page_num"))
	page_num, _ := strconv.Atoi(p_n)
	p_s := (r.URL.Query().Get("page_size"))
	page_size, _ := strconv.Atoi(p_s)

	CreatedOnAsc := strings.Contains(sort_by, "created_on") && strings.Contains(sort_order, "asc")
	CreatedOnDesc := strings.Contains(sort_by, "created_on") && strings.Contains(sort_order, "desc")
	NameAsc := strings.Contains(sort_by, "name") && strings.Contains(sort_order, "asc")
	NameDesc := strings.Contains(sort_by, "name") && strings.Contains(sort_order, "desc")
	eNameAsc := strings.Contains(sort_by, "editorName") && strings.Contains(sort_order, "asc")
	eNameDesc := strings.Contains(sort_by, "editorName") && strings.Contains(sort_order, "desc")
	licensingasc := strings.Contains(sort_by, "licensing") && strings.Contains(sort_order, "asc")
	licensingdesc := strings.Contains(sort_by, "licensing") && strings.Contains(sort_order, "desc")
	locationTypeAsc := strings.Contains(sort_by, "locationType") && strings.Contains(sort_order, "asc")
	locationTypeDesc := strings.Contains(sort_by, "locationType") && strings.Contains(sort_order, "desc")

	PageNum := page_size * (page_num - 1)
	PageSize := page_size
	logger.Log.Info("GetProducts", zap.Any("before listProducts query", time.Now()))
	rows, err := handler.Db.QueryContext(r.Context(), listProducts, ispname, pname, iseditorname, pditorname, islocation, plocation, islicensing, is_opensource, is_closesource, both, CreatedOnAsc, CreatedOnDesc, NameAsc, NameDesc, licensingasc, licensingdesc, locationTypeAsc, locationTypeDesc, eNameAsc, eNameDesc, PageNum, PageSize, iseditorId, pditorId)
	logger.Log.Info("GetProducts", zap.Any("after listProducts query", time.Now()))
	defer rows.Close()
	//PageNum := page_size * (page_num - 1)
	//PageSize := page_size
	if err != nil {
		logger.Log.Error("rest - getproducts - Internal Server Error", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}
	// this should be replaced by redis
	scopesmap := make(map[string]string)
	cronCtx, err := CreateSharedContext(handler.AuthAPI, handler.Application)
	if err != nil {
		logger.Log.Error("couldnt fetch token, will try next time when API will execute", zap.Any("error", err))
	}
	if cronCtx != nil {
		logger.Log.Info("GetProducts", zap.Any("before AddClaimsInContext", time.Now()))
		cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, handler.VerifyKey, handler.APIKey)
		if err != nil {
			logger.Log.Error("Cron AddClaims Failed", zap.Error(err))
		}
		logger.Log.Info("GetProducts", zap.Any("before ListScopes call", time.Now()))
		listScopesNames, err := handler.account.ListScopes(cronAPIKeyCtx, &accv1.ListScopesRequest{})
		logger.Log.Info("GetProducts", zap.Any("after ListScopes call", time.Now()))
		if err != nil {
			logger.Log.Error("couldnt fetch scope name, will try next time when API will execute", zap.Any("error", err))
		}
		if listScopesNames != nil {
			for _, v := range listScopesNames.Scopes {
				scopesmap[v.ScopeCode] = v.ScopeName
			}
		}
	}
	logger.Log.Info("GetProducts", zap.Any("before Parsing", time.Now()))

	for rows.Next() {
		var productResponse ProductsDBResponse
		var support_vendors, metrics, closeLicences, usefulLinks, versions, scopes []byte
		var swidTagProduct, openLicences sql.NullString
		if err := rows.Scan(
			&productResponse.TotalRecords,
			&productResponse.ID,
			&productResponse.Name,
			&productResponse.Editorid,
			&productResponse.GenearlInformation,
			&productResponse.ContractTips,
			&support_vendors,
			&metrics,
			&productResponse.IsOpensource,
			&openLicences,
			&productResponse.IsClosesource,
			&closeLicences,
			&productResponse.Location,
			&productResponse.CreatedOn,
			&productResponse.UpdatedOn,
			&productResponse.Recommendation,
			&usefulLinks,
			&swidTagProduct,
			&productResponse.EditorName,
			&productResponse.OpensourceType,
			&versions,
			&scopes,
		); err != nil {
			logger.Log.Error("rest - getProducts ", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
		}
		if len(support_vendors) != 0 {
			productResponse.SupportVendors = support_vendors
		} else {
			str := make([]string, 1)
			s, _ := json.Marshal(str)
			productResponse.SupportVendors = s
		}
		if len(metrics) != 0 {
			productResponse.Metrics = metrics
		} else {
			str := make([]string, 1)
			s, _ := json.Marshal(str)
			productResponse.Metrics = s
		}
		if len(closeLicences) != 0 {
			productResponse.LicensesClosesource = closeLicences
		} else {
			str := make([]string, 1)
			s, _ := json.Marshal(str)
			productResponse.LicensesClosesource = s
		}
		if len(usefulLinks) != 0 {
			productResponse.UsefulLinks = usefulLinks
		} else {
			str := make([]string, 1)
			s, _ := json.Marshal(str)
			productResponse.UsefulLinks = s
		}
		if len(versions) != 0 {
			productResponse.Versions = versions
		} else {
			str := make([]Version, 1)
			s, _ := json.Marshal(str)
			productResponse.Versions = s
		}

		if len(scopes) != 0 && scopesmap != nil {
			var scopeCode []string
			err = json.Unmarshal(scopes, &scopeCode)
			if err != nil {
				logger.Log.Error("rest - product - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
				return
			}
			var scopes []string
			for _, v := range scopeCode {
				scopes = append(scopes, scopesmap[v])
			}
			marScopeNames, _ := json.Marshal(scopes)
			productResponse.Scopes = marScopeNames

		} else {
			str := make([]string, 0)
			s, _ := json.Marshal(str)
			productResponse.Scopes = s
		}
		productResponse.SwidTagProduct = string(swidTagProduct.String)
		productResponse.LicencesOpensource = openLicences
		productsDbResponse = append(productsDbResponse, productResponse)
	}
	//Listing response
	apiresp.Product = make([]*ProductResponse, len(productsDbResponse))
	if len(productsDbResponse) > 0 {
		apiresp.TotalRecords = int32(productsDbResponse[0].TotalRecords)
		for v := range productsDbResponse {
			apiresp.Product[v] = &ProductResponse{}
			apiresp.Product[v].Version = make([]*Version, 0)
			//list of versions
			jsonStr, err := json.Marshal(productsDbResponse[v].Versions)
			if err != nil {
				logger.Log.Error("service/v1 - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusBadRequest, "Marshal Error", w)
			}
			var versions []VersionCatalog
			if string(jsonStr) == "[null]" {
				str := make([]*Version, 0)
				apiresp.Product[v].Version = str
			} else {
				json.Unmarshal(jsonStr, &versions)
			}
			for _, version := range versions {
				var ver Version
				eos, _ := time.Parse("2006-01-02T15:04:05Z07:00", version.EndOfSupport+"+00:00")
				eol, _ := time.Parse("2006-01-02T15:04:05Z07:00", version.EndOfLife+"+00:00")
				if version.EndOfLife == "1970-01-01T00:00:00" || version.EndOfLife == "0001-01-01T00:00:00" {
					ver.EndOfLife = nil
				} else {
					// pbtime, _ := ptypes.TimestampProto(ver.EndOfLife.Time)
					ver.EndOfLife = &eol
				}
				if version.EndOfSupport == "1970-01-01T00:00:00" || version.EndOfSupport == "0001-01-01T00:00:00" {
					ver.EndOfSupport = nil
				} else {
					ver.EndOfSupport = &eos
				}

				ver.Id = version.ID
				ver.Name = version.Name
				ver.SwidTagVersion = version.SwidTagVersion
				ver.Recommendation = version.Recommendation
				apiresp.Product[v].Version = append(apiresp.Product[v].Version, &ver)
			}
			apiresp.Product[v].CloseSource = new(CloseSource)
			apiresp.Product[v].OpenSource = new(OpenSource)
			apiresp.Product[v].Id = productsDbResponse[v].ID
			apiresp.Product[v].Name = productsDbResponse[v].Name
			apiresp.Product[v].ProductSwidTag = string(productsDbResponse[v].SwidTagProduct)
			apiresp.Product[v].EditorID = productsDbResponse[v].Editorid
			apiresp.Product[v].EditorName = productsDbResponse[v].EditorName
			apiresp.Product[v].Recommendation = productsDbResponse[v].Recommendation.String
			apiresp.Product[v].UpdatedOn = productsDbResponse[v].UpdatedOn
			apiresp.Product[v].CreatedOn = productsDbResponse[v].CreatedOn

			apiresp.Product[v].CloseSource.IsCloseSource = productsDbResponse[v].IsClosesource.Bool
			jsonStr, err = json.Marshal(productsDbResponse[v].LicensesClosesource)
			if err != nil {
				logger.Log.Error("service/v1 - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusBadRequest, "Marshal Error", w)
			}
			if string(jsonStr) == "null" {
				str := make([]string, 0)
				apiresp.Product[v].CloseSource.CloseLicences = str
			} else {
				json.Unmarshal(jsonStr, &apiresp.Product[v].CloseSource.CloseLicences)
			}

			apiresp.Product[v].OpenSource.IsOpenSource = productsDbResponse[v].IsOpensource.Bool
			apiresp.Product[v].OpenSource.OpenLicences = string(productsDbResponse[v].LicencesOpensource.String)
			apiresp.Product[v].OpenSource.OpensourceType = string(productsDbResponse[v].OpensourceType)

			apiresp.Product[v].ContracttTips = productsDbResponse[v].ContractTips.String
			apiresp.Product[v].GenearlInformation = productsDbResponse[v].GenearlInformation.String
			apiresp.Product[v].LocationType = string(productsDbResponse[v].Location)

			jsonStr, err = json.Marshal(productsDbResponse[v].Metrics)
			if err != nil {
				logger.Log.Error("service/v1 - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusBadRequest, "Marshal Error", w)
			}
			if string(jsonStr) == "null" {
				str := make([]string, 0)
				apiresp.Product[v].Metrics = str
			} else {
				json.Unmarshal(jsonStr, &apiresp.Product[v].Metrics)
			}
			jsonStr, err = json.Marshal(productsDbResponse[v].SupportVendors)
			if err != nil {
				logger.Log.Error("service/v1 - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusBadRequest, "Marshal Error", w)
			}
			if string(jsonStr) == "null" {
				str := make([]string, 0)
				apiresp.Product[v].SupportVendors = str
			} else {
				json.Unmarshal(jsonStr, &apiresp.Product[v].SupportVendors)
			}

			jsonStr, err = json.Marshal(productsDbResponse[v].UsefulLinks)
			if err != nil {
				logger.Log.Error("service/v1 - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusBadRequest, "Marshal Error", w)
			}
			if string(jsonStr) == "null" {
				str := make([]string, 0)
				apiresp.Product[v].UsefulLinks = str
			} else {
				json.Unmarshal(jsonStr, &apiresp.Product[v].UsefulLinks)
			}

			jsonStr, err = json.Marshal(productsDbResponse[v].Scopes)
			if err != nil {
				logger.Log.Error("service/v1 - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusBadRequest, "Marshal Error", w)
			}
			if string(jsonStr) == "null" {
				str := make([]string, 0)
				apiresp.Product[v].Scopes = str
			} else {
				json.Unmarshal(jsonStr, &apiresp.Product[v].Scopes)
			}
		}

	}
	logger.Log.Info("GetProducts", zap.Any("after Parsing", time.Now()))
	e, err := json.Marshal(apiresp)
	if err != nil {
		logger.Log.Error("rest - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusBadGateway, err.Error(), w)
		return
	}
	logger.Log.Info("GetProducts", zap.Any("end", time.Now()))
	w.Write([]byte(e))
}

const getVendorAssTable = `select count(*),v.name from editor_catalog_clone e
left join association_vendors_editors ave on ave.e_id =e.id
left join vendors v on ave.v_id =v.id
group by v.name
order by COUNT(*) DESC
OFFSET $1 LIMIT $2;`
const getVendorTable = `select count(v.name),v.name from editor_catalog_clone e
join vendors1 v on v.e_id= e.id
group by v.name
order by COUNT(*) asc
OFFSET $1 LIMIT $2;
`
const getVendorJson = `select count(*),(foo->>'name')::text as g from (select json_array_elements(vendors::json) as foo from editor_catalog_clone  WHERE vendors::TEXT <> 'null') as j group by (foo->>'name')::text order by COUNT(*) asc OFFSET $1 LIMIT $2;;`

func (handler *handler) GetTesting(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Benchmark GetTesting", zap.Any("start time", time.Now()))
	var response []resp
	var rows *sql.Rows
	var before time.Time
	var after time.Time
	p_n := (r.URL.Query().Get("page_num"))
	page_num, _ := strconv.Atoi(p_n)
	p_s := (r.URL.Query().Get("page_size"))
	page_size, _ := strconv.Atoi(p_s)
	switch r.URL.Query().Get("search_key") {
	case "json":
		before = time.Now()
		rows, _ = handler.Db.QueryContext(r.Context(), getVendorJson, page_num, page_size)
		after = time.Now()
	case "has many":
		before = time.Now()
		rows, _ = handler.Db.QueryContext(r.Context(), getVendorTable, page_num, page_size)
		after = time.Now()
	case "has many through":
		before = time.Now()
		rows, _ = handler.Db.QueryContext(r.Context(), getVendorAssTable, page_num, page_size)
		after = time.Now()
	}
	logger.Log.Info("Benchmark GetTesting", zap.Any("differnences", after.Sub(before)))
	defer rows.Close()
	logger.Log.Info("Benchmark GetTesting", zap.Any("before parshing", time.Now()))
	for rows.Next() {
		var resp resp
		if err := rows.Scan(
			&resp.Count,
			&resp.Name,
		); err != nil {
			logger.Log.Error("rest - getProducts ", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
		}
		response = append(response, resp)
	}
	logger.Log.Info("Benchmark GetTesting", zap.Any("after parshing", time.Now()))
	// e, err := json.Marshal(response)
	// if err != nil {
	// 	logger.Log.Error("rest - ListProducts - Marshal", zap.String("Reason: ", err.Error()))
	// 	sendErrorResponse(http.StatusBadGateway, err.Error(), w)
	// 	return
	// }
	json.NewEncoder(w).Encode(response)
	// w.Write([]byte(e))
	logger.Log.Info("Benchmark GetTesting", zap.Any("end time", time.Now()))

}
