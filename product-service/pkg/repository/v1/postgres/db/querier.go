// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"context"
)

type Querier interface {
	AddComputedLicenses(ctx context.Context, arg AddComputedLicensesParams) error
	AddComputedLicensesToAggregation(ctx context.Context, arg AddComputedLicensesToAggregationParams) error
	AggregatedRightDetails(ctx context.Context, arg AggregatedRightDetailsParams) (AggregatedRightDetailsRow, error)
	CounterFeitedProductsCosts(ctx context.Context, arg CounterFeitedProductsCostsParams) ([]CounterFeitedProductsCostsRow, error)
	CounterFeitedProductsLicences(ctx context.Context, arg CounterFeitedProductsLicencesParams) ([]CounterFeitedProductsLicencesRow, error)
	CounterfeitPercent(ctx context.Context, scope string) (CounterfeitPercentRow, error)
	DeleteAcqrightBySKU(ctx context.Context, arg DeleteAcqrightBySKUParams) error
	DeleteAcqrightsByScope(ctx context.Context, scope string) error
	DeleteAggregatedRightBySKU(ctx context.Context, arg DeleteAggregatedRightBySKUParams) error
	DeleteAggregatedRightsByScope(ctx context.Context, scope string) error
	DeleteAggregation(ctx context.Context, arg DeleteAggregationParams) error
	DeleteAggregationByScope(ctx context.Context, scope string) error
	DeleteOverallComputedLicensesByScope(ctx context.Context, scope string) error
	DeleteProductApplications(ctx context.Context, arg DeleteProductApplicationsParams) error
	DeleteProductEquipments(ctx context.Context, arg DeleteProductEquipmentsParams) error
	DeleteProductsByScope(ctx context.Context, scope string) error
	EquipmentProducts(ctx context.Context, equipmentID string) ([]ProductsEquipment, error)
	GetAcqBySwidtags(ctx context.Context, arg GetAcqBySwidtagsParams) ([]GetAcqBySwidtagsRow, error)
	GetAcqRightBySKU(ctx context.Context, arg GetAcqRightBySKUParams) (GetAcqRightBySKURow, error)
	GetAcqRightFileDataBySKU(ctx context.Context, arg GetAcqRightFileDataBySKUParams) ([]byte, error)
	GetAcqRightMetricsBySwidtag(ctx context.Context, arg GetAcqRightMetricsBySwidtagParams) ([]GetAcqRightMetricsBySwidtagRow, error)
	GetAcqRightsByEditor(ctx context.Context, arg GetAcqRightsByEditorParams) ([]GetAcqRightsByEditorRow, error)
	GetAcqRightsCost(ctx context.Context, scope []string) (GetAcqRightsCostRow, error)
	GetAggRightMetricsByAggregationId(ctx context.Context, arg GetAggRightMetricsByAggregationIdParams) ([]GetAggRightMetricsByAggregationIdRow, error)
	GetAggregatedRightBySKU(ctx context.Context, arg GetAggregatedRightBySKUParams) (GetAggregatedRightBySKURow, error)
	GetAggregatedRightsFileDataBySKU(ctx context.Context, arg GetAggregatedRightsFileDataBySKUParams) ([]byte, error)
	GetAggregationByEditor(ctx context.Context, arg GetAggregationByEditorParams) ([]GetAggregationByEditorRow, error)
	GetAggregationByID(ctx context.Context, arg GetAggregationByIDParams) (Aggregation, error)
	GetAggregationByName(ctx context.Context, arg GetAggregationByNameParams) (Aggregation, error)
	GetDashboardUpdates(ctx context.Context, arg GetDashboardUpdatesParams) (GetDashboardUpdatesRow, error)
	GetEquipmentsBySwidtag(ctx context.Context, arg GetEquipmentsBySwidtagParams) ([]string, error)
	GetIndividualProductDetailByAggregation(ctx context.Context, arg GetIndividualProductDetailByAggregationParams) ([]GetIndividualProductDetailByAggregationRow, error)
	GetIndividualProductForAggregationCount(ctx context.Context, arg GetIndividualProductForAggregationCountParams) (int64, error)
	GetLicensesCost(ctx context.Context, scope []string) (GetLicensesCostRow, error)
	GetProductInformation(ctx context.Context, arg GetProductInformationParams) (GetProductInformationRow, error)
	GetProductInformationFromAcqright(ctx context.Context, arg GetProductInformationFromAcqrightParams) (GetProductInformationFromAcqrightRow, error)
	GetProductOptions(ctx context.Context, arg GetProductOptionsParams) ([]GetProductOptionsRow, error)
	GetProductsByEditor(ctx context.Context, arg GetProductsByEditorParams) ([]GetProductsByEditorRow, error)
	GetTotalCounterfietAmount(ctx context.Context, scope string) (float64, error)
	GetTotalDeltaCost(ctx context.Context, scope string) (float64, error)
	GetTotalUnderusageAmount(ctx context.Context, scope string) (float64, error)
	InsertAggregation(ctx context.Context, arg InsertAggregationParams) (int32, error)
	InsertOverAllComputedLicences(ctx context.Context, arg InsertOverAllComputedLicencesParams) error
	ListAcqRightsAggregation(ctx context.Context, arg ListAcqRightsAggregationParams) ([]ListAcqRightsAggregationRow, error)
	ListAcqRightsIndividual(ctx context.Context, arg ListAcqRightsIndividualParams) ([]ListAcqRightsIndividualRow, error)
	ListAcqrightsProducts(ctx context.Context) ([]ListAcqrightsProductsRow, error)
	ListAcqrightsProductsByScope(ctx context.Context, scope string) ([]ListAcqrightsProductsByScopeRow, error)
	ListAggregationNameByScope(ctx context.Context, scope string) ([]string, error)
	ListAggregationNameWithScope(ctx context.Context) ([]ListAggregationNameWithScopeRow, error)
	ListAggregations(ctx context.Context, arg ListAggregationsParams) ([]ListAggregationsRow, error)
	ListDeployedAndAcquiredEditors(ctx context.Context, scope string) ([]string, error)
	ListEditors(ctx context.Context, scope []string) ([]string, error)
	ListEditorsForAggregation(ctx context.Context, scope string) ([]string, error)
	ListMetricsForAggregation(ctx context.Context, scope string) ([]string, error)
	ListProductAggregation(ctx context.Context, arg ListProductAggregationParams) ([]ListProductAggregationRow, error)
	ListProductsAggregationIndividual(ctx context.Context, arg ListProductsAggregationIndividualParams) ([]ListProductsAggregationIndividualRow, error)
	ListProductsByApplicationInstance(ctx context.Context, arg ListProductsByApplicationInstanceParams) ([]ListProductsByApplicationInstanceRow, error)
	ListProductsForAggregation(ctx context.Context, arg ListProductsForAggregationParams) ([]ListProductsForAggregationRow, error)
	ListProductsView(ctx context.Context, arg ListProductsViewParams) ([]ListProductsViewRow, error)
	ListProductsViewRedirectedApplication(ctx context.Context, arg ListProductsViewRedirectedApplicationParams) ([]ListProductsViewRedirectedApplicationRow, error)
	ListProductsViewRedirectedEquipment(ctx context.Context, arg ListProductsViewRedirectedEquipmentParams) ([]ListProductsViewRedirectedEquipmentRow, error)
	ListSelectedProductsForAggregration(ctx context.Context, arg ListSelectedProductsForAggregrationParams) ([]ListSelectedProductsForAggregrationRow, error)
	OverDeployedProductsCosts(ctx context.Context, arg OverDeployedProductsCostsParams) ([]OverDeployedProductsCostsRow, error)
	OverDeployedProductsLicences(ctx context.Context, arg OverDeployedProductsLicencesParams) ([]OverDeployedProductsLicencesRow, error)
	OverdeployPercent(ctx context.Context, scope string) (OverdeployPercentRow, error)
	ProductsNotAcquired(ctx context.Context, scope string) ([]ProductsNotAcquiredRow, error)
	ProductsNotDeployed(ctx context.Context, scope string) ([]ProductsNotDeployedRow, error)
	ProductsPerMetric(ctx context.Context, scope string) ([]ProductsPerMetricRow, error)
	UpdateAggregation(ctx context.Context, arg UpdateAggregationParams) error
	UpsertAcqRights(ctx context.Context, arg UpsertAcqRightsParams) error
	UpsertAggregatedRights(ctx context.Context, arg UpsertAggregatedRightsParams) error
	UpsertDashboardUpdates(ctx context.Context, arg UpsertDashboardUpdatesParams) error
	UpsertProduct(ctx context.Context, arg UpsertProductParams) error
	UpsertProductApplications(ctx context.Context, arg UpsertProductApplicationsParams) error
	UpsertProductEquipments(ctx context.Context, arg UpsertProductEquipmentsParams) error
	UpsertProductPartial(ctx context.Context, arg UpsertProductPartialParams) error
}

var _ Querier = (*Queries)(nil)
