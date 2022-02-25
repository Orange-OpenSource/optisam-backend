package v1

// ProductSearchKey ...
type ProductSearchKey string

const (
	ProductSearchKeySwidTag ProductSearchKey = "swidtag"
	ProductSearchKeyName    ProductSearchKey = "name"
	ProductSearchKeyEditor  ProductSearchKey = "editor"
)

// ProductChildData ...
type ProductChildData struct {
	SwidTag   string
	Name      string
	Edition   string
	Editor    string
	Version   string
	Metric    string
	TotalCost float64
}

// ProductAdditionalData ...
type ProductAdditionalData struct {
	Name              string
	Swidtag           string
	Version           string
	Editor            string
	NumOfApplications int32
	NumofEquipments   int32
	NumofOptions      int32
	Child             []ProductChildData
}

// ProductAcquiredRight represents product acquired rights.
type ProductAcquiredRight struct {
	SKU          string
	Metric       string
	AcqLicenses  uint64
	TotalCost    float64
	AvgUnitPrice float64
}

// ProductAdditionalInfo ...
type ProductAdditionalInfo struct {
	Products []ProductAdditionalData
}

// TotalRecords ...
type TotalRecords struct {
	TotalCnt int32
}

// ProductData ...
type ProductData struct {
	Name              string
	Version           string
	Category          string
	Editor            string
	Swidtag           string
	NumOfEquipments   int32
	NumOfApplications int32
	TotalCost         float64
}

// ProductInfo ...
type ProductInfo struct {
	NumOfRecords []TotalRecords
	Products     []ProductData
}

// QueryProducts ....
type QueryProducts struct {
	PageSize  int32
	Offset    int32
	SortBy    string
	SortOrder string
	Filter    *AggregateFilter
	AcqFilter *AggregateFilter
	AggFilter *AggregateFilter
}

// QueryApplicationsForProduct ...
type QueryApplicationsForProduct struct { //
	SwidTag   string
	PageSize  int32
	Offset    int32
	SortBy    string
	SortOrder SortOrder
	Filter    *AggregateFilter
}

// QueryInstancesForApplicationProduct ...
type QueryInstancesForApplicationProduct struct { //
	SwidTag   string
	AppID     string
	PageSize  int32
	Offset    int32
	SortBy    int32
	SortOrder SortOrder
}

// ApplicationsForProductData ...
type ApplicationsForProductData struct {
	ApplicationID   string
	Name            string
	Owner           string
	NumOfEquipments int32
	NumOfInstances  int32
}

// ApplicationsForProduct ...
type ApplicationsForProduct struct {
	NumOfRecords []TotalRecords
	Applications []ApplicationsForProductData
}

// InstancesForApplicationProductData ...
type InstancesForApplicationProductData struct {
	Name            string
	ID              string
	Environment     string
	NumOfEquipments int32
	NumOfProducts   int32
}

// InstancesForApplicationProduct ...
type InstancesForApplicationProduct struct {
	NumOfRecords []TotalRecords
	Instances    []InstancesForApplicationProductData
}

// AggregateFilter is a collection of filters
type AggregateFilter struct {
	Filters []Queryable
}

func (a *AggregateFilter) Len() int {
	return len(a.Filters)
}

func (a *AggregateFilter) Less(i, j int) bool {
	return a.Filters[i].Priority() > a.Filters[j].Priority()
}

func (a *AggregateFilter) Swap(i, j int) {
	a.Filters[i], a.Filters[j] = a.Filters[j], a.Filters[i]
}

type Filtertype int32

const (
	RegexFilter Filtertype = 0
	EqFilter    Filtertype = 1
)

// Filter has filtering key and value
type Filter struct {
	FilteringPriority   int32
	FilterMatchingType  Filtertype
	FilterKey           string      // Key of filter
	FilterValue         interface{} // Search value for filter
	FilterValueMultiple []interface{}
}

// Key Queryable key method.
func (f *Filter) Key() string {
	return f.FilterKey
}

// Value Queryable Value method.
func (f *Filter) Value() interface{} {
	return f.FilterValue
}

// Values Queryable Values method.
func (f *Filter) Values() []interface{} {
	return f.FilterValueMultiple
}

// Priority Queryable Value method.
func (f *Filter) Priority() int32 {
	return f.FilteringPriority
}

// Type Queryable Value method.
func (f *Filter) Type() Filtertype {
	return f.FilterMatchingType
}

func (prodKey ProductSearchKey) ToString() string {
	return string(prodKey)
}
