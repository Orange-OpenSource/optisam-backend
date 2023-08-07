package rest

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"net/http"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	repo "optisam-backend/catalog-service/pkg/repository/v1/postgres"
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
	pCRepo      *repo.ProductCatalogRepository
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
	Id                 string      `json:"id"`
	EditorID           string      `json:"editorID"`
	Name               string      `json:"name"`
	Metrics            []string    `json:"metrics"`
	GenearlInformation string      `json:"genearlInformation"`
	ContracttTips      string      `json:"contracttTips"`
	LocationType       string      `json:"locationType"`
	OpenSource         *OpenSource `json:"openSource"`
	Version            []*Version  `json:"version"`
	Recommendation     string      `json:"recommendation"`
	UsefulLinks        []string    `json:"usefulLinks"`
	SupportVendors     []string    `json:"supportVendors"`
	CreatedOn          time.Time   `json:"createdOn"`
	UpdatedOn          time.Time   `json:"UpdatedOn"`
	ProductSwidTag     string      `json:"productSwidTag"`
	EditorName         string      `json:"editorName"`
	Scopes             []string    `json:"scopes"`
	Licensing          string      `json:"licensing"`
}

type OpenSource struct {
	OpenLicences   string `json:"openLicences"`
	OpensourceType string `json:"openSourceType"`
}
type CloseSource struct {
	IsCloseSource bool     `json:"isCloseSource"`
	CloseLicences []string `json:"closeLicences,omitempty"`
}

type Version struct {
	Id             string     `json:"id"`
	SwidTagVersion string     `json:"swidTagVersion"`
	Name           string     `json:"name"`
	Recommendation string     `json:"recommendation"`
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
	TotalRecords       int64           `json:"totalRecords"`
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Editorid           string          `json:"editorid"`
	GenearlInformation sql.NullString  `json:"genearl_information"`
	ContractTips       sql.NullString  `json:"contract_tips"`
	SupportVendors     json.RawMessage `json:"support_vendors"`
	Metrics            json.RawMessage `json:"metrics"`
	// IsOpensource        sql.NullBool      `json:"is_opensource"`
	LicencesOpensource sql.NullString `json:"licences_opensource"`
	// IsClosesource       sql.NullBool      `json:"is_closesource"`
	// LicensesClosesource json.RawMessage   `json:"licenses_closesource"`
	Location       db.LocationType            `json:"location"`
	OpensourceType db.OpensourceType          `json:"opensource_type"`
	CreatedOn      time.Time                  `json:"created_on"`
	UpdatedOn      time.Time                  `json:"updated_on"`
	Recommendation sql.NullString             `json:"recommendation"`
	UsefulLinks    json.RawMessage            `json:"useful_links"`
	SwidTagProduct string                     `json:"swid_tag_product"`
	EditorName     string                     `json:"editor_name"`
	Versions       json.RawMessage            `json:"versions"`
	Scopes         json.RawMessage            `json:"scopes"`
	Licensing      db.ProductCatalogLicensing `json:"licensing"`
}

type LocationType string
type Licensing string
type OpensourceType string

type Product struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Editorid           string          `json:"editorid"`
	GenearlInformation sql.NullString  `json:"genearl_information"`
	ContractTips       sql.NullString  `json:"contract_tips"`
	SupportVendors     json.RawMessage `json:"support_vendors"`
	Metrics            json.RawMessage `json:"metrics"`
	// IsOpensource        sql.NullBool    `json:"is_opensource"`
	LicencesOpensource sql.NullString `json:"licences_opensource"`
	// IsClosesource       sql.NullBool    `json:"is_closesource"`
	// LicensesClosesource json.RawMessage `json:"licenses_closesource"`
	Location       LocationType    `json:"location"`
	Licensing      Licensing       `json:" licensing"`
	CreatedOn      time.Time       `json:"created_on"`
	UpdatedOn      time.Time       `json:"updated_on"`
	Recommendation sql.NullString  `json:"recommendation"`
	UsefulLinks    json.RawMessage `json:"useful_links"`
	SwidTagProduct sql.NullString  `json:"swid_tag_product"`
	Source         sql.NullString  `json:"source"`
	EditorName     string          `json:"editor_name"`
	OpensourceType OpensourceType  `json:"opensource_type"`
	Scopes         []string        `json:"scopes"`
}

type AuditResponse struct {
	Entity string  `json:"entity"`
	Date   *string `json:"date"`
	Year   int     `json:"year"`
}
type ListEditorResponse struct {
	TotalRecords int      `json:"totalrecords"`
	Editors      []Editor `json:"editors"`
}
type Editor struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	GeneralInformation   string          `json:"general_information"`
	PartnerManagers      json.RawMessage `json:"partner_managers"`
	Audits               json.RawMessage `json:"audits"`
	Vendors              json.RawMessage `json:"vendors"`
	CreatedOn            time.Time       `json:"created_on"`
	UpdatedOn            time.Time       `json:"updated_on"`
	ProductCount         int             `json:"product_count"`
	Scopes               []string        `json:"scopes"`
	CountryCode          string          `json:"country_code"`
	Address              string          `json:"address"`
	GroupContract        bool            `json:"groupContract"`
	GlobalAccountManager json.RawMessage `json:"global_account_manager"`
	Sourcers             json.RawMessage `json:"sourcers"`
}

type Vendors struct {
	Name string `json:"name"`
}

type Managers struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Audits struct {
	Entity string     `json:"entity"`
	Date   *time.Time `json:"date"`
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

type Productsfilters struct {
	DeploymentType FilterDetail `json:"deploymentType,omitempty"`
	Licensing      FilterDetail `json:"licensing,omitempty"`
	Recommendation FilterDetail `json:"recommendation,omitempty"`
	Entities       FilterDetail `json:"entities,omitempty"`
	Vendors        FilterDetail `json:"vendors,omitempty"`
}
type Editorfilters struct {
	GroupContract FilterDetail `json:"groupContract,omitempty"`
	Year          FilterDetail `json:"year,omitempty"`
	CountryCode   FilterDetail `json:"countryCode,omitempty"`
	Entities      FilterDetail `json:"entities,omitempty"`
}

type FilterDetail struct {
	TotalCount int      `json:"total_count"`
	Filter     []Filter `json:"filter"`
}
type Filter struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
type FilterDbResponse struct {
	TotalCount int    `json:"total_count"`
	Name       string `json:"name"`
	Count      int    `json:"count"`
}
