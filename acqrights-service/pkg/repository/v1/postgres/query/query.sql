-- name: UpsertAcqRights :exec
INSERT INTO acqrights (sku,swidtag,product_name,product_editor,entity,scope,metric,num_licenses_acquired,num_licences_maintainance,avg_unit_price,avg_maintenance_unit_price,total_purchase_cost,total_maintenance_cost,total_cost,created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
ON CONFLICT (sku)
DO
UPDATE SET swidtag = $2,product_name = $3,product_editor = $4,entity = $5,scope = $6,metric = $7,num_licenses_acquired = $8,
            num_licences_maintainance = $9,avg_unit_price = $10,avg_maintenance_unit_price = $11,total_purchase_cost = 12,
            total_maintenance_cost = $13,total_cost = $14,updated_on = $16,updated_by = $17;


-- name: ListAcqRightsIndividual :many
SELECT count(*) OVER() AS totalRecords,a.entity,a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost FROM 
acqrights a
WHERE 
  a.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(a.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(a.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(a.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(a.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(a.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(a.product_editor) = lower(@product_editor) ELSE TRUE END)
  AND (CASE WHEN @lk_sku::bool THEN lower(a.sku) LIKE '%' || lower(@sku::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_sku::bool THEN lower(a.sku) = lower(@sku) ELSE TRUE END)
  AND (CASE WHEN @lk_metric::bool THEN lower(a.metric) LIKE '%' || lower(@metric::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_metric::bool THEN lower(a.metric) = lower(@metric) ELSE TRUE END)
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN a.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN a.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN a.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN a.product_name END desc,
  CASE WHEN @product_editor_asc::bool THEN a.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN a.product_editor END desc,
  CASE WHEN @sku_asc::bool THEN a.sku END asc,
  CASE WHEN @sku_desc::bool THEN a.sku END desc,
  CASE WHEN @metric_asc::bool THEN a.metric END asc,
  CASE WHEN @metric_desc::bool THEN a.metric END desc,
  CASE WHEN @entity_asc::bool THEN a.entity END asc,
  CASE WHEN @entity_desc::bool THEN a.entity END desc,
  CASE WHEN @num_licenses_acquired_asc::bool THEN a.num_licenses_acquired END asc,
  CASE WHEN @num_licenses_acquired_desc::bool THEN a.num_licenses_acquired END desc,
  CASE WHEN @num_licences_maintainance_asc::bool THEN a.num_licences_maintainance END asc,
  CASE WHEN @num_licences_maintainance_desc::bool THEN a.num_licences_maintainance END desc,
  CASE WHEN @avg_unit_price_asc::bool THEN a.avg_unit_price END asc,
  CASE WHEN @avg_unit_price_desc::bool THEN a.avg_unit_price END desc,  
  CASE WHEN @avg_maintenance_unit_price_asc::bool THEN a.avg_maintenance_unit_price END asc,
  CASE WHEN @avg_maintenance_unit_price_desc::bool THEN a.avg_maintenance_unit_price END desc,
  CASE WHEN @total_purchase_cost_asc::bool THEN a.total_purchase_cost END asc,
  CASE WHEN @total_purchase_cost_desc::bool THEN a.total_purchase_cost END desc,
  CASE WHEN @total_maintenance_cost_asc::bool THEN a.total_maintenance_cost END asc,
  CASE WHEN @total_maintenance_cost_desc::bool THEN a.total_maintenance_cost END desc,
  CASE WHEN @total_cost_asc::bool THEN a.total_cost END asc,
  CASE WHEN @total_cost_desc::bool THEN a.total_cost END desc
  LIMIT @page_size OFFSET @page_num;

-- name: ListAcqRightsAggregation :many
SELECT count(*) OVER() AS totalRecords,aggregation_id,aggregation_name,a.product_editor,a.metric,array_agg(a.sku)::TEXT[] as skus,array_agg(a.swidtag)::TEXT[] as swidtags,SUM(a.total_cost)::REAL as total_cost FROM 
acqrights a JOIN (SELECT aggregation_id,aggregation_name,aggregation_metric as metric,unnest(products) as swidtag FROM aggregations) ag
ON a.swidtag = ag.swidtag AND a.metric = ag.metric
WHERE 
    a.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(a.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(a.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_aggregation_name::bool THEN lower(ag.aggregation_name) LIKE '%' || lower(@aggregation_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_aggregation_name::bool THEN lower(ag.aggregation_name) = lower(@aggregation_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(a.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(a.product_editor) = lower(@product_editor) ELSE TRUE END)
  AND (CASE WHEN @lk_sku::bool THEN lower(a.sku) LIKE '%' || lower(@sku::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_sku::bool THEN lower(a.sku) = lower(@sku) ELSE TRUE END)
  AND (CASE WHEN @lk_metric::bool THEN lower(a.metric) LIKE '%' || lower(@metric::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_metric::bool THEN lower(a.metric) = lower(@metric) ELSE TRUE END)
  GROUP BY ag.aggregation_id,ag.aggregation_name,a.product_editor,a.metric
  ORDER BY
  CASE WHEN @aggregation_name_asc::bool THEN aggregation_name END asc,
  CASE WHEN @aggregation_name_desc::bool THEN aggregation_name END desc,
  CASE WHEN @product_editor_asc::bool THEN a.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN a.product_editor END desc,
  CASE WHEN @metric_asc::bool THEN a.metric END asc,
  CASE WHEN @metric_desc::bool THEN a.metric END desc,
  CASE WHEN @total_cost_asc::bool THEN 8 END asc,
  CASE WHEN @total_cost_desc::bool THEN 8 END desc
  LIMIT @page_size OFFSET @page_num;


-- name: ListAcqRightsAggregationIndividual :many
SELECT a.entity,a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost FROM 
acqrights a
WHERE 
  a.swidtag IN (SELECT UNNEST(products) from aggregations where aggregation_id = @aggregation_id)
  AND a.scope = ANY(@scope::TEXT[]);

-- name: InsertAggregation :one
INSERT INTO aggregations (aggregation_name,aggregation_metric,aggregation_scope,products,created_by)
VALUES ($1,$2,$3,$4,$5) RETURNING *;

-- name: UpdateAggregation :one
UPDATE aggregations
SET aggregation_name = @aggregation_name,products = @products
WHERE aggregation_id = @aggregation_id
AND aggregation_scope = ANY(@scope::TEXT[])
RETURNING *;

-- name: DeleteAggregation :exec
DELETE FROM aggregations 
WHERE aggregation_id = @aggregation_id
AND aggregation_scope = ANY(@scope::TEXT[]);

-- name: ListAggregation :many
SELECT aggregation_id,aggregation_name,aggregation_metric,acq.product_editor,aggregation_scope,
ARRAY_AGG(acq.product_name)::TEXT[] as product_names,ARRAY_AGG(agg.swidtag)::TEXT[] as product_swidtags
FROM acqrights acq JOIN 
(SELECT aggregation_id,aggregation_name,aggregation_metric,aggregation_scope,unnest(products) swidtag,created_on,created_by,updated_on,updated_by from aggregations WHERE aggregation_scope = ANY(@scope::TEXT[])) agg
ON acq.swidtag = agg.swidtag AND acq.metric = agg.aggregation_metric
GROUP BY agg.aggregation_id,agg.aggregation_name,agg.aggregation_metric,acq.product_editor,agg.aggregation_scope;

-- name: ListAcqRightsProducts :many
SELECT swidtag,product_name
FROM acqrights acq
WHERE swidtag NOT IN (SELECT UNNEST(products) from aggregations)
AND acq.metric = @metric
AND acq.product_editor = @editor
AND acq.scope = @scope;

-- name: ListAcqRightsEditors :many
SELECT DISTINCT acq.product_editor
FROM acqrights acq
WHERE acq.scope = $1;

-- name: ListAcqRightsMetrics :many
SELECT DISTINCT acq.metric
FROM acqrights acq
WHERE acq.scope = $1;