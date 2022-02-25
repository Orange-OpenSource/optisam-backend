package v1

// AggregationInfoFull contains info of that aggregation and individual rights associated with products
type AggregationInfo struct {
	ID                   int
	Name                 string
	SKU                  string
	ProductNames         []string
	Swidtags             []string
	ProductIDs           []string
	Editor               string
	Metric               []string
	Licenses             int32
	MaintenanceLicenses  int32
	UnitPrice            float64
	MaintenanceUnitPrice float64
	PurchaseCost         float64
	MaintenanceCost      float64
	TotalCost            float64
	StartOfMaintenance   string
	EndOfMaintenance     string
	NumOfApplications    int32
	NumOfEquipments      int32
}

// ProductAggregation is the logical grouping of products
type ProductAggregation struct {
	ID                string
	Name              string
	Editor            string
	Product           string
	Metric            string
	MetricName        string
	NumOfApplications int
	NumOfEquipments   int
	TotalCost         float64
	Products          []string // list of ids of the prioduct which  are in aggregations
	ProductsFull      []*ProductData
	AcqRights         []string
	AcqRightsFull     []*AcquiredRights
}

// UpdateProductAggregationRequest contains members which needs to be updated in product aggregation
type UpdateProductAggregationRequest struct {
	Name            string
	AddedProducts   []string
	RemovedProducts []string
	// Products will come as comma separated strings
	Product string
}

// ProductAggSortBy gives the attributes on which sorting is allowed
type ProductAggSortBy int32

const (
	// ProductAggSortByName ...
	ProductAggSortByName        ProductAggSortBy = 1
	ProductAggSortByEditor      ProductAggSortBy = 2
	ProductAggSortByNumApp      ProductAggSortBy = 3
	ProductAggSortByNumEquips   ProductAggSortBy = 4
	ProductAggSortByProductName ProductAggSortBy = 5
	ProductAggSortByMetric      ProductAggSortBy = 6
	ProductAggSortByTotalCost   ProductAggSortBy = 7
)

// ProductAggSearchKey only needed for search key
type ProductAggSearchKey string

const (
	ProductAggSearchKeyName        ProductAggSearchKey = "name"
	ProductAggSearchKeyEditor      ProductAggSearchKey = "editor"
	ProductAggSearchKeyProductName ProductAggSearchKey = "product_name"
	ProductAggSearchKeySwidTag     ProductAggSearchKey = "swidtag"
)

// QueryProductAggregations are query params required for quering aggregations
type QueryProductAggregations struct {
	PageSize        int32
	Offset          int32
	SortBy          ProductAggSortBy
	SortOrder       SortOrder
	Filter          *AggregateFilter
	ProductFilter   *AggregateFilter
	AcqRightsFilter *AggregateFilter
	MetricFilter    *AggregateFilter
}

func (prodAggKey ProductAggSearchKey) ToString() string {
	return string(prodAggKey)
}
