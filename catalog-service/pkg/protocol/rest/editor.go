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
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

const getEditor = `-- name: GetEditor :one
SELECT editor_catalog.id, editor_catalog.name, editor_catalog.general_information, editor_catalog.partner_managers, editor_catalog.audits, editor_catalog.vendors, editor_catalog.created_on, editor_catalog.updated_on, COUNT(product_catalog.id), (Select json_agg(t.scope) as a from (
    (Select scope from products where product_editor = editor_catalog.name )
    UNION
    (select scope from acqrights where product_editor = editor_catalog.name)
) t) from editor_catalog
LEFT JOIN product_catalog ON editor_catalog.id = product_catalog.editorID
where editor_catalog.id = $1
GROUP BY editor_catalog.id
`

// const listEditors = `-- name: GetEditor :many
// SELECT count(*) OVER() AS totalRecords, editor_catalog.id, editor_catalog.name, editor_catalog.general_information, editor_catalog.partner_managers, editor_catalog.audits, editor_catalog.vendors, editor_catalog.created_on, editor_catalog.updated_on, COUNT(product_catalog.id),(Select json_agg(t.scope) as a from (
//     (Select scope from products where product_editor = editor_catalog.name )
//     UNION
//     (select scope from acqrights where product_editor = editor_catalog.name)
// ) t) from editor_catalog
// LEFT JOIN product_catalog ON editor_catalog.id = product_catalog.editorID
// where
// (CASE WHEN $5::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($6::TEXT) || '%' ELSE TRUE END)
// GROUP BY editor_catalog.id
// ORDER BY
//   CASE WHEN $1::bool THEN editor_catalog.created_on END asc,
//   CASE WHEN $2::bool THEN editor_catalog.created_on END desc,
//   CASE WHEN $7::bool THEN editor_catalog.name END asc,
//   CASE WHEN $8::bool THEN editor_catalog.name END desc,
//   CASE WHEN $9::bool THEN COUNT(product_catalog.id) END asc,
//   CASE WHEN $10::bool THEN COUNT(product_catalog.id) END desc
// LIMIT $3 OFFSET $4
// `

const listEditors = `-- name: GetEditor :many
select * ,(Select json_agg(t.scope) as a from (
    (Select scope from products where product_editor = editor.name )
    UNION
    (select scope from acqrights where product_editor = editor.name)
) t) from (
SELECT count(editor_catalog.id) OVER() AS totalRecords, editor_catalog.id, editor_catalog.name,
editor_catalog.general_information, editor_catalog.partner_managers, editor_catalog.audits,
editor_catalog.vendors, editor_catalog.created_on,
editor_catalog.updated_on, COUNT(product_catalog.id) as pcount from editor_catalog
LEFT JOIN product_catalog ON editor_catalog.id = product_catalog.editorID
where
(CASE WHEN $5::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($6::TEXT) || '%' ELSE TRUE END)
GROUP BY editor_catalog.id
order by 
CASE WHEN $1::bool THEN editor_catalog.created_on END asc,
CASE WHEN $2::bool THEN editor_catalog.created_on END desc,
CASE WHEN $7::bool THEN editor_catalog.name END asc,
CASE WHEN $8::bool THEN editor_catalog.name END desc,
CASE WHEN $9::bool THEN COUNT(product_catalog.id) END asc,
CASE WHEN $10::bool THEN COUNT(product_catalog.id) END desc
LIMIT $3 OFFSET $4) as editor
`

const listEditorNames = `-- name: GetEditorNames :many
SELECT id, name from editor_catalog
where
(CASE WHEN $1::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($2::TEXT) || '%' ELSE TRUE END)
GROUP BY editor_catalog.id
ORDER BY
  CASE WHEN $3::bool THEN editor_catalog.name END asc,
  CASE WHEN $4::bool THEN editor_catalog.name END desc
LIMIT $5 OFFSET $6
`
const listEditorNamesAll = `-- name: GetEditorNames :many
SELECT id, name from editor_catalog
where
(CASE WHEN $1::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($2::TEXT) || '%' ELSE TRUE END)
GROUP BY editor_catalog.id
ORDER BY
  CASE WHEN $3::bool THEN editor_catalog.name END asc,
  CASE WHEN $4::bool THEN editor_catalog.name END desc
`

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
	var vendors, audits, managers, scopes []byte
	var createdOn, updatedOn time.Time
	var genInfo sql.NullString
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
		var man = make([]PartnerManagers, 0)
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
			audit := &AuditResponse{Date: el, Entity: a.Entity}
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
			editor.Scopes = listScopesNames.ScopeNames
		}

	}
	editor.GeneralInformation = genInfo.String
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
		logger.Log.Error("rest - geteditor - Meathod Not Allowed", zap.String("Reason: ", "Meathod Not Allowed"))
		sendErrorResponse(http.StatusMethodNotAllowed, "Meathod Not Allowed", w)
		return
	}

	search_params_name := (r.URL.Query().Get("search_params.name.filteringkey"))
	isename := (search_params_name != "")
	ename := strings.ToLower(search_params_name)

	createdonAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "createdOn") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	createdonDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "createdOn") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))
	NameAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "name") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	NameDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "name") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))
	pcountAsc := (strings.Contains(r.URL.Query().Get("sortBy"), "productsCount") && (strings.Contains(r.URL.Query().Get("sortOrder"), "asc")))
	pcountDesc := (strings.Contains(r.URL.Query().Get("sortBy"), "productsCount") && (strings.Contains(r.URL.Query().Get("sortOrder"), "desc")))
	pageNum := strcomp.StringToNum(r.URL.Query().Get("pageSize")) * (strcomp.StringToNum(r.URL.Query().Get("pageNum")) - 1)
	pageSize := strcomp.StringToNum(r.URL.Query().Get("pageSize"))
	logger.Log.Info("ListEditors", zap.Any("before list editors query", time.Now()))
	rows, err := handler.Db.QueryContext(r.Context(), listEditors, createdonAsc, createdonDesc, pageSize, pageNum, isename, ename, NameAsc, NameDesc, pcountAsc, pcountDesc)
	logger.Log.Info("ListEditors", zap.Any("after list editors query", time.Now()))
	if err != nil {
		logger.Log.Error("rest - geteditor - Internal Server Error", zap.String("Reason: ", err.Error()))
		sendErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}
	defer rows.Close()
	// this should be replaced by redis
	scopesmap := make(map[string]string)
	cronCtx, err := CreateSharedContext(handler.AuthAPI, handler.Application)
	if err != nil {
		logger.Log.Error("couldnt fetch token, will try next time when API will execute", zap.Any("error", err))
	}
	if cronCtx != nil {
		logger.Log.Info("ListEditors", zap.Any("before AddClaimsInContext", time.Now()))
		cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, handler.VerifyKey, handler.APIKey)
		logger.Log.Info("ListEditors", zap.Any("after AddClaimsInContext", time.Now()))
		if err != nil {
			logger.Log.Error("Cron AddClaims Failed", zap.Error(err))
		}
		logger.Log.Info("ListEditors", zap.Any("before list scopes call", time.Now()))
		listScopesNames, err := handler.account.ListScopes(cronAPIKeyCtx, &accv1.ListScopesRequest{})
		logger.Log.Info("ListEditors", zap.Any("after list scopes call", time.Now()))
		if err != nil {
			logger.Log.Error("couldnt fetch scope name, will try next time when API will execute", zap.Any("error", err))
		}
		if listScopesNames != nil {
			for _, v := range listScopesNames.Scopes {
				scopesmap[v.ScopeCode] = v.ScopeName
			}
		}
	}
	var response ListEditorResponse
	var totalRecords int
	logger.Log.Info("ListEditors", zap.Any("before data parsing", time.Now()))
	for rows.Next() {
		var editor Editor
		var vendors, audits, managers, scopes []byte
		var createdOn, updatedOn time.Time
		var genInfo sql.NullString
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
			var ven = make([]PartnerManagers, 0)
			ra, _ := json.Marshal(ven)
			editor.PartnerManagers = ra
		}

		editor.GeneralInformation = genInfo.String
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
				var el *string
				var e string
				if a.Date != nil {
					e = a.Date.AsTime().String()
					if e != "1970-01-01 12:00:00 +0000 UTC" {
						el = &e
					}
				}
				audit := &AuditResponse{Date: el, Entity: a.Entity}
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
	// resp, err := json.Marshal(response)
	// if err != nil {
	// 	logger.Log.Error("rest - geteditor - Marshal response", zap.String("Reason: ", err.Error()))
	// 	sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
	// 	return
	// }
	// w.Write([]byte(resp))
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
	// resp, err := json.Marshal(response)
	// if err != nil {
	// 	logger.Log.Error("rest - geteditor - Marshal response", zap.String("Reason: ", err.Error()))
	// 	sendErrorResponse(http.StatusMethodNotAllowed, err.Error(), w)
	// 	return
	// }
	// w.Write([]byte(resp))
}

func validateRequest(r *http.Request) error {
	if (r.URL.Query().Get("sortBy") == "") || (r.URL.Query().Get("sortOrder") == "") || (r.URL.Query().Get("pageSize") == "") || (r.URL.Query().Get("pageNum") == "") {
		return errors.New("parameter missing")
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
