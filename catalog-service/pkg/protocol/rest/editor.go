package rest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/common/optisam/config"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func (handler *handler) GetEditor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var editor Editor

	if r.Method != http.MethodGet {
		logger.Log.Error("rest - geteditor - Meathod Not Allowed", zap.String("Reason: ", "Meathod Not Allowed"))
		sendErrorResponse(http.StatusMethodNotAllowed, "Meathod Not Allowed", w)
		return
	}

	id := (r.URL.Query().Get("id"))

	if id == "" {
		logger.Log.Error("rest - geteditor - QueryParams", zap.String("Reason: ", "Unable to fetch Query Params"))
		sendErrorResponse(http.StatusMethodNotAllowed, "Unable to fetch Query Params", w)
		return
	}
	var vendors, audits, managers, scopes, generalAccountManager, sourcers []byte
	var createdOn, updatedOn time.Time
	var genInfo, address, code sql.NullString

	row := handler.Db.QueryRowContext(r.Context(), getEditor, id)
	err := row.Scan(
		&editor.ID,
		&editor.Name,
		&genInfo,
		&managers,
		&audits,
		&vendors,
		&createdOn,
		&updatedOn,
		&code,
		&address,
		&editor.GroupContract,
		&generalAccountManager,
		&sourcers,
		&editor.ProductCount,
		&scopes,
	)
	if err != nil {
		logger.Log.Error("rest- geteditor - geteditor", zap.String("reason", err.Error()))
		sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
		return
	}
	if len(managers) != 0 && string(managers) != "null" {
		err = json.Unmarshal(managers, &editor.PartnerManagers)
		if err != nil {
			logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
			return
		}
	} else {
		var man = make([]Managers, 0)
		ra, _ := json.Marshal(man)
		editor.PartnerManagers = ra
	}

	if len(audits) != 0 && string(audits) != "null" {
		var pbaudits []*v1.Audits
		var respAudits []*AuditResponse
		err = json.Unmarshal(audits, &pbaudits)
		if err != nil {
			logger.Log.Error("rest - geteditor - Marshal,audits", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
			return
		}
		for _, a := range pbaudits {
			var el *string
			var e string
			if a.Date != nil {
				e = a.Date.AsTime().String()
				if e != "1970-01-01 12:00:00 +0000 UTC" {
					el = &e
				}
			}

			audit := &AuditResponse{Date: el, Entity: a.Entity, Year: int(a.Year)}
			respAudits = append(respAudits, audit)
		}

		ra, err := json.Marshal(respAudits)
		if err != nil {
			logger.Log.Error("rest - geteditor - Marshal single audit", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
			return
		}
		editor.Audits = ra
	} else {
		var aud = make([]Audits, 0)
		ra, _ := json.Marshal(aud)
		editor.Audits = ra
	}
	if len(vendors) != 0 && string(vendors) != "null" {
		err = json.Unmarshal(vendors, &editor.Vendors)
		if err != nil {
			logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
			return
		}
	} else {
		var ven = make([]Vendors, 0)
		ra, _ := json.Marshal(ven)
		editor.Vendors = ra
	}

	if len(generalAccountManager) != 0 && string(generalAccountManager) != "null" {
		err = json.Unmarshal(generalAccountManager, &editor.GlobalAccountManager)
		if err != nil {
			logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
			return
		}
	} else {
		var man = make([]Managers, 0)
		ra, _ := json.Marshal(man)
		editor.GlobalAccountManager = ra
	}

	if len(sourcers) != 0 && string(sourcers) != "null" {
		err = json.Unmarshal(sourcers, &editor.Sourcers)
		if err != nil {
			logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
			return
		}
	} else {
		var man = make([]Managers, 0)
		ra, _ := json.Marshal(man)
		editor.GlobalAccountManager = ra
	}

	if scopes != nil {
		var scopeCode []string
		err = json.Unmarshal(scopes, &scopeCode)
		if err != nil {
			logger.Log.Error("rest - product - Marshal", zap.String("Reason: ", err.Error()))
			sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
			return
		}
		scopes_response, err := handler.pCRepo.GetScope(r.Context(), scopeCode)
		if err != nil {
			logger.Log.Info("couldn't fetch data from redis", zap.Any("error", err))
		}
		if len(scopes_response) == len(scopeCode) {
			listScopeMap := map[string]string{}
			for _, v := range scopes_response {
				listScopeMap[v.ScopeCode] = v.ScopeName
			}
			for i, v := range scopeCode {
				scopeCode[i] = listScopeMap[v]
			}
			editor.Scopes = scopeCode
		} else {
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
				editor.Scopes = listScopesNames.ScopeNames
			}
		}
	}
	editor.GeneralInformation = genInfo.String
	editor.Address = address.String
	editor.CountryCode = code.String
	editor.CreatedOn = createdOn
	editor.UpdatedOn = updatedOn

	e, err := json.Marshal(editor)
	if err != nil {
		logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(e))
}

func (handler *handler) ListEditors(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("ListEditors", zap.Any("started listing of editors", time.Now()))
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		logger.Log.Error("rest - ListEditors - Meathod Not Allowed", zap.String("Reason: ", "Meathod Not Allowed"))
		sendErrorResponse(http.StatusMethodNotAllowed, "Meathod Not Allowed", w)
		return
	}
	err := validatePagination(r)
	if err != nil {
		logger.Log.Error("rest - ListEditors - pagination error", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
		return
	}
	//name
	search_params_name := (r.URL.Query().Get("search_params.name.filteringkey"))
	isename := (search_params_name != "")
	ename := strings.ToLower(search_params_name)

	//group contract
	search_params_group_contract := (r.URL.Query().Get("search_params.group_contract.filteringkey"))
	is_search_params_group_contract := (search_params_group_contract != "")
	// group_contract := strings.Split(search_params_group_contract, ",")
	var group_contract bool
	if is_search_params_group_contract {
		if strings.Contains(search_params_group_contract, "false,true") {
			is_search_params_group_contract = false
		} else {
			var err error
			group_contract, err = strconv.ParseBool(search_params_group_contract)
			if err != nil {
				logger.Log.Error("rest - geteditors - Internal Server Error", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
				return
			}
		}
	}

	entities := r.URL.Query().Get("search_params.entities.filteringkey")
	entitiesType := strings.Split(entities, ",")
	isEntitiesType := (entities != "")

	countrycodes := r.URL.Query().Get("search_params.countryCodes.filteringkey")
	countrycodesType := strings.Split(countrycodes, ",")
	iscountrycodesType := (countrycodes != "")

	audityears := r.URL.Query().Get("search_params.audityears.filteringkey")
	audityearsType := strings.Split(audityears, ",")
	isaudityearsType := (audityears != "")

	createdonAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "createdOn") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	createdonDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "createdOn") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))
	NameAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "name") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	NameDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "name") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))
	pcountAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "productsCount") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	pcountDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "productsCount") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))
	gcontractAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "groupContract") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	gcontractDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "groupContract") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))

	// pagination
	pageNum := strcomp.StringToNum(r.URL.Query().Get("pageSize")) * (strcomp.StringToNum(r.URL.Query().Get("pageNum")) - 1)
	pageSize := strcomp.StringToNum(r.URL.Query().Get("pageSize"))

	var rows *sql.Rows
	var query string

	scopesmap := make(map[string]string)
	scopes, err := handler.pCRepo.GetAllScope(r.Context())
	if err != nil {
		logger.Log.Error("couldn't fetch scopes from redis", zap.Any("error", err))
	}
	if len(scopes) > 0 {
		for _, v := range scopes {
			scopesmap[v.ScopeCode] = v.ScopeName
		}
	} else {
		scopesmap = createScopeMap(handler)
	}

	logger.Log.Info("ListEditors", zap.Any("before editors query", time.Now()))
	if !isEntitiesType {
		query = listEditors
		rows, err = handler.Db.QueryContext(r.Context(), query, createdonAsc, createdonDesc, pageSize, pageNum, isename, ename, NameAsc, NameDesc, pcountAsc, pcountDesc, is_search_params_group_contract, group_contract, gcontractAsc, gcontractDesc, iscountrycodesType, pq.Array(countrycodesType), isaudityearsType, pq.Array(audityearsType))
	} else {
		var scopeCodes []string
		reversescopesmap := make(map[string]string)
		for k, v := range scopesmap {
			reversescopesmap[v] = k
		}
		for _, entity := range entitiesType {
			scopeCodes = append(scopeCodes, reversescopesmap[entity])
		}
		query = innerScopeEditor
		rows, err = handler.Db.QueryContext(r.Context(), query, createdonAsc, createdonDesc, pageSize, pageNum, isename, ename, NameAsc, NameDesc, pcountAsc, pcountDesc, is_search_params_group_contract, group_contract, gcontractAsc, gcontractDesc, iscountrycodesType, pq.Array(countrycodesType), isaudityearsType, pq.Array(audityearsType), pq.Array(scopeCodes))
	}
	if err != nil {
		logger.Log.Error("rest - ListEditors - Internal Server Error", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}
	logger.Log.Info("ListEditors", zap.Any("after list editors query", time.Now()))

	defer rows.Close()

	var response ListEditorResponse
	var totalRecords int
	logger.Log.Info("ListEditors", zap.Any("before data parsing", time.Now()))
	for rows.Next() {
		var editor Editor
		var vendors, audits, managers, scopes, sourcers, account_manager []byte
		var createdOn, updatedOn time.Time
		var genInfo, address, code sql.NullString
		err := rows.Scan(
			&totalRecords,
			&editor.ID,
			&editor.Name,
			&genInfo,
			&managers,
			&audits,
			&vendors,
			&createdOn,
			&updatedOn,
			&code,
			&address,
			&editor.GroupContract,
			&account_manager,
			&sourcers,
			&editor.ProductCount,
			&scopes,
		)
		if err != nil {
			logger.Log.Error("rest- geteditor - geteditor", zap.String("reason", err.Error()))
			sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
			return
		}
		if len(managers) != 0 && string(managers) != "null" {
			err = json.Unmarshal(managers, &editor.PartnerManagers)
			if err != nil {
				logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
				return
			}
		} else {
			var ven = make([]Managers, 0)
			ra, _ := json.Marshal(ven)
			editor.PartnerManagers = ra
		}
		if len(sourcers) != 0 && string(sourcers) != "null" {
			err = json.Unmarshal(sourcers, &editor.Sourcers)
			if err != nil {
				logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
				return
			}
		} else {
			var ven = make([]Managers, 0)
			ra, _ := json.Marshal(ven)
			editor.Sourcers = ra
		}
		if len(account_manager) != 0 && string(account_manager) != "null" {
			err = json.Unmarshal(account_manager, &editor.GlobalAccountManager)
			if err != nil {
				logger.Log.Error("rest - geteditor - Marshal", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
				return
			}
		} else {
			var ven = make([]Managers, 0)
			ra, _ := json.Marshal(ven)
			editor.GlobalAccountManager = ra
		}

		editor.GeneralInformation = genInfo.String
		editor.Address = address.String
		editor.CountryCode = code.String
		var pbaudits []*v1.Audits
		var respAudits []*AuditResponse
		if len(audits) != 0 && string(audits) != "null" {
			err = json.Unmarshal(audits, &pbaudits)
			if err != nil {
				logger.Log.Error("rest - geteditor - Marshal,audits", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
				return
			}
			for _, a := range pbaudits {
				// var audit *AuditResponse
				// if a.Date > 1970 {
				// 	audit = &AuditResponse{Date: &a.Date, Entity: a.Entity}
				// } else {
				// 	audit = &AuditResponse{Date: nil, Entity: a.Entity}
				// }
				var el *string
				var e string
				if a.Date != nil {
					e = a.Date.AsTime().String()
					if e != "1970-01-01 12:00:00 +0000 UTC" {
						el = &e

					}
				}
				if el == nil {
					a.Year = 0
				}
				audit := &AuditResponse{Date: el, Entity: a.Entity, Year: int(a.Year)}

				respAudits = append(respAudits, audit)
			}

			ra, err := json.Marshal(respAudits)
			if err != nil {
				logger.Log.Error("rest - geteditor - Marshal single audit", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
				return
			}
			editor.Audits = ra
		} else {
			var aud = make([]Audits, 0)
			ra, _ := json.Marshal(aud)
			editor.Audits = ra
		}
		if len(vendors) != 0 && string(vendors) != "null" {
			err = json.Unmarshal(vendors, &editor.Vendors)
			if err != nil {
				logger.Log.Error("rest - geteditor - Marshal vendors", zap.String("Reason: ", err.Error()))
				sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
				return
			}
		} else {
			var ven = make([]Vendors, 0)
			ra, _ := json.Marshal(ven)
			editor.Vendors = ra
		}

		if scopes != nil && scopesmap != nil {
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
			editor.Scopes = scopes
		}
		editor.CreatedOn = createdOn
		editor.UpdatedOn = updatedOn
		response.Editors = append(response.Editors, editor)
	}

	response.TotalRecords = totalRecords
	if response.TotalRecords == 0 {
		editor := []Editor{}
		response.Editors = editor
	}
	logger.Log.Info("ListEditors", zap.Any("after data parsing", time.Now()))
	w.WriteHeader(http.StatusOK)
	logger.Log.Info("ListEditors", zap.Any("end", time.Now()))
	json.NewEncoder(w).Encode(response)
}

func (handler *handler) ListEditorNames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		logger.Log.Error("rest - geteditor - Meathod Not Allowed", zap.String("Reason: ", "Meathod Not Allowed"))
		sendErrorResponse(http.StatusMethodNotAllowed, "Meathod Not Allowed", w)
		return
	}
	search_params_name := (r.URL.Query().Get("searchParams.name.filteringkey"))
	isename := (search_params_name != "")
	ename := strings.ToLower(search_params_name)
	nameAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "name") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	nameDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "name") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))
	pageNum := strcomp.StringToNum(r.URL.Query().Get("pageSize")) * (strcomp.StringToNum(r.URL.Query().Get("pageNum")) - 1)
	pageSize := strcomp.StringToNum(r.URL.Query().Get("pageSize"))
	var rows *sql.Rows
	var err error
	if pageSize == 0 {
		rows, err = handler.Db.QueryContext(r.Context(), listEditorNamesAll, isename, ename, nameAsc, nameDesc)
	} else {
		rows, err = handler.Db.QueryContext(r.Context(), listEditorNames, isename, ename, nameAsc, nameDesc, pageSize, pageNum)
	}
	if err != nil {
		logger.Log.Error("rest - geteditor - Internal Server Error", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}
	defer rows.Close()
	var editor EditorNames
	var response ListEditorNames
	for rows.Next() {
		err := rows.Scan(
			&editor.ID,
			&editor.Name,
		)
		if err != nil {
			logger.Log.Error("rest- geteditor - geteditor", zap.String("reason", err.Error()))
			sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
			return
		}
		response.Editors = append(response.Editors, editor)
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
}

func validateRequest(r *http.Request) error {
	if (r.URL.Query().Get("sortBy") == "") || (r.URL.Query().Get("sortOrder") == "") || (r.URL.Query().Get("pageSize") == "") || (r.URL.Query().Get("pageNum") == "") {
		return errors.New("parameter missing")
	}
	return nil
}

func validatePagination(r *http.Request) error {
	ps, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		return err
	}
	if ps < 10 || ps > 200 {
		return errors.New("page size is allowed between 10 and 200")
	}
	pn, err := strconv.Atoi(r.URL.Query().Get("pageNum"))
	if err != nil {
		return err
	}
	if pn < 1 || pn > 1000 {
		return errors.New("page number is allowed between 1 and 1000")
	}
	return nil

}

// CreateSharedContext will return admin auth token
func CreateSharedContext(api string, appCred config.Application) (*context.Context, error) {
	logger.Log.Info("CreateSharedContext", zap.Any("CreateSharedContext called", time.Now()))
	ctx := context.Background()
	respMap := make(map[string]interface{})
	data := url.Values{
		"username":   {appCred.UserNameSuperAdmin},
		"password":   {appCred.PasswordSuperAdmin},
		"grant_type": {"password"},
	}
	logger.Log.Info("CreateSharedContext", zap.Any("before auth api", time.Now()))
	resp, err := http.PostForm(api, data) // nolint: gosec
	logger.Log.Info("CreateSharedContext", zap.Any("after auth api", time.Now()))
	if err != nil {
		log.Println("Failed to get user claims  ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	// log.Println(" Token Data received", string(bodyBytes))
	err = json.Unmarshal(bodyBytes, &respMap)
	if err != nil {
		log.Println("failed to unmarshal byte data", err)
		return nil, err
	}
	authStr := fmt.Sprintf("Bearer %s", respMap["access_token"].(string))
	md := metadata.Pairs("Authorization", authStr)

	ctx = metadata.NewIncomingContext(ctx, md)
	logger.Log.Info("CreateSharedContext", zap.Any("CreateSharedContext executed", time.Now()))

	return &ctx, nil
}
