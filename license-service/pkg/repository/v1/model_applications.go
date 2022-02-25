package v1

// ApplicationData gives the details os an application
type ApplicationData struct {
	ID               string
	Name             string
	ApplicationID    string
	ApplicationOwner string
	NumOfInstances   int32
	NumOfProducts    int32
}

// ApplicationInfo ...
type ApplicationInfo struct {
	NumOfRecords []TotalRecords
	Applications []ApplicationData
}

// QueryApplications ...
type QueryApplications struct { //
	PageSize  int32
	Offset    int32
	SortBy    string
	SortOrder string
	Filter    *AggregateFilter
}

// ProductsForApplicationData ...
type ProductsForApplicationData struct {
	Name            string
	Version         string
	Editor          string
	Edition         string
	SwidTag         string
	NumOfEquipments int32
	NumOfInstances  int32
	TotalCost       float64
}

// ApplicationDetails gives the details on an application
type ApplicationDetails struct {
	Name             string
	ApplicationID    string
	ApplicationOwner string
	NumOfInstances   int32
	NumOfProducts    int32
}

// ProductsForApplication ...
type ProductsForApplication struct {
	NumOfRecords []TotalRecords
	Products     []ProductsForApplicationData
}

type ApplicationsSortBy int32

const (
	ApplicationsSortByAppID        ApplicationsSortBy = 1
	ApplicationsSortByName         ApplicationsSortBy = 2
	ApplicationsSortByAppOwner     ApplicationsSortBy = 3
	ApplicationsSortByNumInstances ApplicationsSortBy = 4
	ApplicationsSortByNumEquips    ApplicationsSortBy = 5
	ApplicationsSortByNumProducts  ApplicationsSortBy = 6
)

type ApplicationSearchKey string

const (
	ApplicationSearchKeyID       ApplicationSearchKey = "id"
	ApplicationSearchKeyName     ApplicationSearchKey = "name"
	ApplicationSearchKeyAppOwner ApplicationSearchKey = "application_owner"
)

type QueryApplicationsForProductAggregation struct { //
	PageSize  int32
	Offset    int32
	SortBy    ApplicationsSortBy
	SortOrder SortOrder
	Filter    *AggregateFilter
}

func (appAggKey ApplicationSearchKey) String() string {
	return string(appAggKey)
}
