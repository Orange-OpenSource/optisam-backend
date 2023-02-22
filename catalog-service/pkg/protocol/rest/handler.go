package rest

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"net/http"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/config"
	"time"
)

type handler struct {
	Db          *sql.DB
	account     accv1.AccountServiceClient
	AuthAPI     string
	VerifyKey   *rsa.PublicKey
	APIKey      string
	Application config.Application
}
type errorresponse struct {
	Message string
	Code    int
}

func sendErrorResponse(code int, message string, w http.ResponseWriter) {
	w.WriteHeader(code)
	var response errorresponse
	response.Message = message
	response.Code = code
	e, _ := json.Marshal(response)
	w.Write([]byte(e))
}

type ProductResponse struct {
	Id                 string       `json:"id,omitempty"`
	EditorID           string       `json:"editorID,omitempty"`
	Name               string       `json:"name,omitempty"`
	Metrics            []string     `json:"metrics,omitempty"`
	GenearlInformation string       `json:"genearlInformation,omitempty"`
	ContracttTips      string       `json:"contracttTips,omitempty"`
	LocationType       string       `json:"locationType,omitempty"`
	OpenSource         *OpenSource  `json:"openSource,omitempty"`
	CloseSource        *CloseSource `json:"closeSource,omitempty"`
	Version            []*Version   `json:"version,omitempty"`
	Recommendation     string       `json:"recommendation,omitempty"`
	UsefulLinks        []string     `json:"usefulLinks"`
	SupportVendors     []string     `json:"supportVendors"`
	CreatedOn          time.Time    `json:"createdOn,omitempty"`
	UpdatedOn          time.Time    `json:"UpdatedOn,omitempty"`
	ProductSwidTag     string       `json:"productSwidTag"`
	EditorName         string       `json:"editorName,omitempty"`
	Scopes             []string     `json:"scopes"`
}

type OpenSource struct {
	IsOpenSource   bool   `json:"isOpenSource"`
	OpenLicences   string `json:"openLicences"`
	OpensourceType string `json:"openSourceType"`
}
type CloseSource struct {
	IsCloseSource bool     `json:"isCloseSource"`
	CloseLicences []string `json:"closeLicences,omitempty"`
}

type Version struct {
	Id             string     `json:"id,omitempty"`
	SwidTagVersion string     `json:"swidTagVersion"`
	Name           string     `json:"name,omitempty"`
	Recommendation string     `json:"recommendation,omitempty"`
	EndOfLife      *time.Time `json:"endOfLife"`
	EndOfSupport   *time.Time `json:"endOfSupport"`
}

type ListProductResponse struct {
	TotalRecords int32              `json:"total_records"`
	Product      []*ProductResponse `json:"product"`
}

type ProductSearchParams struct {
	Name       *StringFilter `json:"name,omitempty"`
	EditorName *StringFilter `json:"editorName,omitempty"`
}
type StringFilter struct {
	Filteringkey string `json:"filteringkey,omitempty"`
}

type ProductsDBResponse struct {
	TotalRecords        int64             `json:"totalRecords"`
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Editorid            string            `json:"editorid"`
	GenearlInformation  sql.NullString    `json:"genearl_information"`
	ContractTips        sql.NullString    `json:"contract_tips"`
	SupportVendors      json.RawMessage   `json:"support_vendors"`
	Metrics             json.RawMessage   `json:"metrics"`
	IsOpensource        sql.NullBool      `json:"is_opensource"`
	LicencesOpensource  sql.NullString    `json:"licences_opensource"`
	IsClosesource       sql.NullBool      `json:"is_closesource"`
	LicensesClosesource json.RawMessage   `json:"licenses_closesource"`
	Location            db.LocationType   `json:"location"`
	OpensourceType      db.OpensourceType `json:"opensource_type"`
	CreatedOn           time.Time         `json:"created_on"`
	UpdatedOn           time.Time         `json:"updated_on"`
	Recommendation      sql.NullString    `json:"recommendation"`
	UsefulLinks         json.RawMessage   `json:"useful_links"`
	SwidTagProduct      string            `json:"swid_tag_product"`
	EditorName          string            `json:"editor_name"`
	Versions            json.RawMessage   `json:"versions"`
	Scopes              json.RawMessage   `json:"scopes"`
}

type LocationType string
type OpensourceType string

type Product struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Editorid            string          `json:"editorid"`
	GenearlInformation  sql.NullString  `json:"genearl_information"`
	ContractTips        sql.NullString  `json:"contract_tips"`
	SupportVendors      json.RawMessage `json:"support_vendors"`
	Metrics             json.RawMessage `json:"metrics"`
	IsOpensource        sql.NullBool    `json:"is_opensource"`
	LicencesOpensource  sql.NullString  `json:"licences_opensource"`
	IsClosesource       sql.NullBool    `json:"is_closesource"`
	LicensesClosesource json.RawMessage `json:"licenses_closesource"`
	Location            LocationType    `json:"location"`
	CreatedOn           time.Time       `json:"created_on"`
	UpdatedOn           time.Time       `json:"updated_on"`
	Recommendation      sql.NullString  `json:"recommendation"`
	UsefulLinks         json.RawMessage `json:"useful_links"`
	SwidTagProduct      sql.NullString  `json:"swid_tag_product"`
	Source              sql.NullString  `json:"source"`
	EditorName          string          `json:"editor_name"`
	OpensourceType      OpensourceType  `json:"opensource_type"`
	Scopes              []string        `json:"scopes"`
}

type AuditResponse struct {
	Entity string  `json:"entity,omitempty"`
	Date   *string `json:"date"`
}
type ListEditorResponse struct {
	TotalRecords int      `json:"totalrecords"`
	Editors      []Editor `json:"editors"`
}
type Editor struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	GeneralInformation string          `json:"general_information"`
	PartnerManagers    json.RawMessage `json:"partner_managers"`
	Audits             json.RawMessage `json:"audits"`
	Vendors            json.RawMessage `json:"vendors"`
	CreatedOn          time.Time       `json:"created_on"`
	UpdatedOn          time.Time       `json:"updated_on"`
	ProductCount       int             `json:"product_count"`
	Scopes             []string        `json:"scopes"`
}

type Vendors struct {
	Name string `json:"name,omitempty"`
}

type PartnerManagers struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type Audits struct {
	Entity string     `json:"entity"`
	Date   *time.Time `json:"date,omitempty"`
}

type EditorNames struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type ListEditorNames struct {
	Editors []EditorNames `json:"editors"`
}

// func (x *http.Request) GetSearchParams() *ProductSearchParams {
// 	search_params := (x.URL.Query().Get("search_params"))

// 	if x != nil {
// 		return x.SearchParams
// 	}
// 	return nil
// }

type VersionCatalog struct {
	ID             string `json:"id"`
	SwidTagSystem  string `json:"swid_tag_system"`
	PID            string `json:"p_id"`
	Name           string `json:"name"`
	EndOfLife      string `json:"end_of_life"`
	EndOfSupport   string `json:"end_of_support"`
	Recommendation string `json:"recommendation"`
	SwidTagVersion string `json:"swid_tag_version"`
	Source         string `json:"source"`
}
type resp struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}
