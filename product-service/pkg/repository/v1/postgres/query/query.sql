-- name: EquipmentProducts :many
SELECT * from products_equipments
WHERE
equipment_id = $1;

-- name: ListEditors :many
SELECT DISTINCT ON (p.product_editor) p.product_editor 
FROM products p 
WHERE p.scope = ANY(@scope::TEXT[]) AND LENGTH(p.product_editor) > 0;

-- name: ListProductsView :many
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments ,COALESCE(acq.total_cost,0)::FLOAT as cost 
FROM products p 
LEFT JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[]) GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE scope = ANY(@scope::TEXT[]) GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN
(SELECT swidtag, sum(total_cost) as total_cost FROM acqrights WHERE scope = ANY(@scope::TEXT[]) GROUP BY swidtag) acq
ON p.swidtag = acq.swidtag
WHERE
  p.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(p.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(p.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(p.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
  GROUP BY p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition,pa.num_of_applications, pe.num_of_equipments,acq.total_cost
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN p.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN p.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN p.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN p.product_name END desc,
  CASE WHEN @product_edition_asc::bool THEN p.product_edition END asc,
  CASE WHEN @product_edition_desc::bool THEN p.product_edition END desc,
  CASE WHEN @product_category_asc::bool THEN p.product_category END asc,
  CASE WHEN @product_category_desc::bool THEN p.product_category END desc,
  CASE WHEN @product_version_asc::bool THEN p.product_version END asc,
  CASE WHEN @product_version_desc::bool THEN p.product_version END desc,
  CASE WHEN @product_editor_asc::bool THEN p.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN p.product_editor END desc,
  CASE WHEN @num_of_applications_asc::bool THEN num_of_applications END asc,
  CASE WHEN @num_of_applications_desc::bool THEN num_of_applications END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN num_of_equipments END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN num_of_equipments END desc,
  CASE WHEN @cost_asc::bool THEN acq.total_cost END asc,
  CASE WHEN @cost_desc::bool THEN acq.total_cost END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: ListProductsViewRedirectedApplication :many
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments ,COALESCE(acq.total_cost,0)::FLOAT as cost 
FROM products p 
INNER JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[]) AND  (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END) GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE scope = ANY(@scope::TEXT[]) AND (CASE WHEN @is_equipment_id::bool THEN equipment_id = @equipment_id ELSE TRUE END) GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN
(SELECT swidtag, sum(total_cost) as total_cost FROM acqrights WHERE scope = ANY(@scope::TEXT[]) GROUP BY swidtag) acq
ON p.swidtag = acq.swidtag
WHERE
  p.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(p.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(p.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(p.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
  GROUP BY p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition,pa.num_of_applications, pe.num_of_equipments, acq.total_cost
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN p.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN p.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN p.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN p.product_name END desc,
  CASE WHEN @product_edition_asc::bool THEN p.product_edition END asc,
  CASE WHEN @product_edition_desc::bool THEN p.product_edition END desc,
  CASE WHEN @product_category_asc::bool THEN p.product_category END asc,
  CASE WHEN @product_category_desc::bool THEN p.product_category END desc,
  CASE WHEN @product_version_asc::bool THEN p.product_version END asc,
  CASE WHEN @product_version_desc::bool THEN p.product_version END desc,
  CASE WHEN @product_editor_asc::bool THEN p.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN p.product_editor END desc,
  CASE WHEN @num_of_applications_asc::bool THEN num_of_applications END asc,
  CASE WHEN @num_of_applications_desc::bool THEN num_of_applications END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN num_of_equipments END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN num_of_equipments END desc,
  CASE WHEN @cost_asc::bool THEN acq.total_cost END asc,
  CASE WHEN @cost_desc::bool THEN acq.total_cost END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: ListProductsViewRedirectedEquipment :many
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments ,COALESCE(acq.total_cost,0)::FLOAT as cost 
FROM products p 
LEFT JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[]) AND (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END)  GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
INNER JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE  scope = ANY(@scope::TEXT[]) AND (CASE WHEN @is_equipment_id::bool THEN equipment_id = @equipment_id ELSE TRUE END)  GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN
(SELECT swidtag, sum(total_cost) as total_cost FROM acqrights WHERE scope = ANY(@scope::TEXT[]) GROUP BY swidtag) acq
ON p.swidtag = acq.swidtag
WHERE
  p.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(p.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(p.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(p.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
  GROUP BY p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition,pa.num_of_applications, pe.num_of_equipments, acq.total_cost
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN p.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN p.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN p.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN p.product_name END desc,
  CASE WHEN @product_edition_asc::bool THEN p.product_edition END asc,
  CASE WHEN @product_edition_desc::bool THEN p.product_edition END desc,
  CASE WHEN @product_category_asc::bool THEN p.product_category END asc,
  CASE WHEN @product_category_desc::bool THEN p.product_category END desc,
  CASE WHEN @product_version_asc::bool THEN p.product_version END asc,
  CASE WHEN @product_version_desc::bool THEN p.product_version END desc,
  CASE WHEN @product_editor_asc::bool THEN p.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN p.product_editor END desc,
  CASE WHEN @num_of_applications_asc::bool THEN num_of_applications END asc,
  CASE WHEN @num_of_applications_desc::bool THEN num_of_applications END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN num_of_equipments END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN num_of_equipments END desc,
  CASE WHEN @cost_asc::bool THEN acq.total_cost END asc,
  CASE WHEN @cost_desc::bool THEN acq.total_cost END desc
  LIMIT @page_size OFFSET @page_num
; 
-- name: GetProductInformation :one
SELECT p.swidtag,p.product_editor,p.product_edition,p.product_version
FROM products p 
WHERE p.swidtag = @swidtag
AND p.scope = ANY(@scope::TEXT[]);

-- name: GetProductOptions :many
SELECT p.swidtag,p.product_name,p.product_edition,p.product_editor,p.product_version
FROM products p 
WHERE p.option_of = @swidtag
AND p.scope = ANY(@scope::TEXT[]);

-- name: ListAggregationsView :many
SELECT count(*) OVER() AS totalRecords,p.aggregation_id,p.aggregation_name,p.product_editor,array_agg(distinct p.swidtag)::TEXT[] as swidtags, COALESCE(sum(pa.num_of_applications),0)::INTEGER as num_of_applications , COALESCE(sum(pe.num_of_equipments),0)::INTEGER as num_of_equipments , COALESCE(SUM(acq.total_cost),0)::FLOAT as total_cost
FROM products p 
LEFT JOIN
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[])  GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE scope = ANY(@scope::TEXT[])  GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN
(SELECT swidtag,total_cost FROM acqrights LEFT JOIN aggregations on acqrights.metric = aggregations.aggregation_metric WHERE acqrights.scope = ANY(@scope::TEXT[]) AND aggregations.aggregation_scope = ANY(@scope::TEXT[]) GROUP BY swidtag,total_cost) acq
ON p.swidtag = acq.swidtag
WHERE
  p.aggregation_id <> 0
  AND p.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(p.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_aggregation_name::bool THEN lower(p.aggregation_name) LIKE '%' || lower(@aggregation_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_aggregation_name::bool THEN lower(p.aggregation_name) = lower(@aggregation_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
  GROUP BY p.aggregation_id, p.aggregation_name, p.product_editor
  ORDER BY
  CASE WHEN @aggregation_name_asc::bool THEN p.aggregation_name END asc,
  CASE WHEN @aggregation_name_desc::bool THEN p.aggregation_name END desc,
  CASE WHEN @product_editor_asc::bool THEN p.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN p.product_editor END desc,
  CASE WHEN @num_of_applications_asc::bool THEN SUM(pa.num_of_applications) END asc,
  CASE WHEN @num_of_applications_desc::bool THEN SUM(pa.num_of_applications) END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN SUM(pe.num_of_equipments) END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN SUM(pe.num_of_equipments) END desc,
  CASE WHEN @cost_asc::bool THEN sum(acq.total_cost) END asc,
  CASE WHEN @cost_desc::bool THEN sum(acq.total_cost) END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: ListAggregationProductsView :many
SELECT p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition ,COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications,COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments, COALESCE(acq.total_cost,0)::FLOAT as cost
FROM products p 
LEFT JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[])  GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE scope = ANY(@scope::TEXT[])  GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN
(SELECT swidtag, sum(total_cost) as total_cost FROM acqrights LEFT JOIN aggregations on acqrights.metric = aggregations.aggregation_metric WHERE aggregations.aggregation_id =  @aggregation_id AND acqrights.scope = ANY(@scope::TEXT[]) AND aggregations.aggregation_scope = ANY(@scope::TEXT[]) GROUP BY swidtag) acq
ON p.swidtag = acq.swidtag
WHERE
  p.aggregation_id = @aggregation_id
  AND p.scope = ANY(@scope::TEXT[])
GROUP BY p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition,acq.total_cost,pa.num_of_applications, pe.num_of_equipments;


-- name: ProductAggregationDetails :one
SELECT p.aggregation_id,p.aggregation_name,p.product_editor,array_agg(distinct p.swidtag)::TEXT[] as swidtags,array_agg(distinct p.product_edition)::TEXT[] as editions, COALESCE(SUM(pa.num_of_applications),0)::INTEGER as num_of_applications,COALESCE(SUM(pe.num_of_equipments),0)::INTEGER as num_of_equipments, COALESCE(SUM(acq.total_cost),0)::FLOAT as total_cost  
FROM products p 
LEFT JOIN
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[])  GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE scope = ANY(@scope::TEXT[])  GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN
(SELECT swidtag, sum(total_cost) as total_cost FROM acqrights LEFT JOIN aggregations on acqrights.metric = aggregations.aggregation_metric WHERE aggregations.aggregation_id =  @aggregation_id AND acqrights.scope = ANY(@scope::TEXT[]) AND aggregations.aggregation_scope = ANY(@scope::TEXT[]) GROUP BY swidtag) acq
ON p.swidtag = acq.swidtag
WHERE
  p.aggregation_id = @aggregation_id AND p.aggregation_id <> 0
  AND p.scope = ANY(@scope::TEXT[])
GROUP BY p.aggregation_id,p.aggregation_name,p.aggregation_name,p.product_editor;

-- name: ProductAggregationChildOptions :many
SELECT p.swidtag,p.product_name,p.product_edition,p.product_editor,p.product_version
FROM products p 
WHERE p.option_of in (
SELECT p.swidtag
FROM products p
WHERE 
  p.aggregation_id = @aggregation_id
  AND p.scope = ANY(@scope::TEXT[]))
AND p.scope = ANY(@scope::TEXT[]) ;

-- name: UpsertProduct :exec
INSERT INTO products (swidtag, product_name, product_version, product_edition, product_category, product_editor,scope,option_of,created_on,created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
-- SCOPE BASED CHANGE
ON CONFLICT (swidtag,scope)
DO
 UPDATE SET product_name = $2, product_version = $3, product_edition = $4,product_category = $5,product_editor= $6,option_of=$8,updated_on=$11,updated_by=$12;

-- name: UpsertProductPartial :exec
INSERT INTO products (swidtag,scope,created_by)
VALUES ($1,$2,$3)
-- SCOPE BASED CHANGE
ON CONFLICT (swidtag,scope)
DO NOTHING;


-- name: DeleteProductApplications :exec
DELETE FROM products_applications
-- SCOPE BASED CHANGE
WHERE swidtag = @product_id and application_id = ANY(@application_id::TEXT[]) and scope = @scope;

-- name: DeleteProductEquipments :exec
DELETE FROM products_equipments
-- SCOPE BASED CHANGE
WHERE swidtag = @product_id and equipment_id = ANY(@equipment_id::TEXT[]) and scope = @scope;

-- name: GetProductsByEditor :many
SELECT swidtag, product_name
FROM products
WHERE product_editor = @product_editor and scope = ANY(@scopes::TEXT[]);

-- name: UpsertProductAggregation :exec
-- SCOPE BASED CHANGE
Update products set aggregation_id = @aggregation_id, aggregation_name = @aggregation_name WHERE
swidtag = ANY(@swidtags::TEXT[]) AND scope = @scope;

-- name: GetProductAggregation :many
SELECT swidtag
FROM products
WHERE aggregation_id = $1 and aggregation_name = $2;

-- name: DeleteProductAggregation :exec
Update products set aggregation_id = $1, aggregation_name = $2 WHERE
aggregation_id = $3;

-- name: UpsertAcqRights :exec
INSERT INTO acqrights (sku,swidtag,product_name,product_editor,entity,scope,metric,num_licenses_acquired,avg_unit_price,avg_maintenance_unit_price,total_purchase_cost,total_maintenance_cost,total_cost,created_by,start_of_maintenance,end_of_maintenance,num_licences_maintainance,version)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$17,$18,$19,$20)
ON CONFLICT (sku,scope)
DO
UPDATE SET swidtag = $2,product_name = $3,product_editor = $4,entity = $5,scope = $6,metric = $7,num_licenses_acquired = $8,
            avg_unit_price = $9,avg_maintenance_unit_price = $10,total_purchase_cost = $11,
            total_maintenance_cost = $12,total_cost = $13,updated_on = $15,updated_by = $16,start_of_maintenance = $17,end_of_maintenance = $18, num_licences_maintainance = $19, version = $20;


-- name: ListAcqRightsIndividual :many
SELECT count(*) OVER() AS totalRecords,a.entity,a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost ,start_of_maintenance, end_of_maintenance , version FROM 
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
  CASE WHEN @total_cost_desc::bool THEN a.total_cost END desc,
  CASE WHEN @start_of_maintenance_asc::bool THEN a.start_of_maintenance END asc,
  CASE WHEN @start_of_maintenance_desc::bool THEN a.start_of_maintenance END desc,
  CASE WHEN @end_of_maintenance_asc::bool THEN a.end_of_maintenance END asc,
  CASE WHEN @end_of_maintenance_desc::bool THEN a.end_of_maintenance END desc
  LIMIT @page_size OFFSET @page_num;

-- name: ListAcqRightsAggregation :many
SELECT count(*) OVER() AS totalRecords,ag.aggregation_id,ag.aggregation_name,a.product_editor,a.metric,array_agg(a.sku)::TEXT[] as skus,array_agg(a.swidtag)::TEXT[] as swidtags,SUM(a.total_cost)::Numeric(15,2) as total_cost FROM 
acqrights a JOIN (SELECT agg.aggregation_id,agg.aggregation_name,agg.aggregation_metric as metric,unnest(agg.products) as swidtag FROM aggregations agg where agg.aggregation_scope = ANY(@scope::TEXT[]) ) ag
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
SELECT a.entity,a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost, a.version FROM 
acqrights a
WHERE 
  a.swidtag IN (SELECT UNNEST(products) from aggregations where aggregation_id = @aggregation_id  AND a.metric = aggregation_metric AND aggregation_scope = ANY(@scope::TEXT[]))
  AND a.scope = ANY(@scope::TEXT[]);

-- name: InsertAggregation :one
INSERT INTO aggregations (aggregation_name,aggregation_metric,aggregation_scope,products,created_by)
VALUES ($1,$2,$3,$4,$5) RETURNING *;

-- name: UpdateAggregation :one
UPDATE aggregations
SET aggregation_name = @aggregation_name,products = @products
WHERE aggregation_id = @aggregation_id
AND aggregation_scope = @scope
RETURNING *;

-- name: DeleteAggregation :exec
DELETE FROM aggregations 
WHERE aggregation_id = @aggregation_id
AND aggregation_scope = ANY(@scope::TEXT[]);

-- name: ListAggregation :many
SELECT agg.aggregation_id,agg.aggregation_name,agg.aggregation_metric,acq.product_editor,agg.aggregation_scope,
ARRAY_AGG(acq.product_name)::TEXT[] as product_names,ARRAY_AGG(agg.swidtag)::TEXT[] as product_swidtags
FROM acqrights acq JOIN 
(SELECT ag.aggregation_id,ag.aggregation_name,ag.aggregation_metric,ag.aggregation_scope,unnest(ag.products) swidtag,ag.created_on,ag.created_by,updated_on,updated_by from aggregations ag WHERE ag.aggregation_scope = ANY(@scope::TEXT[])) agg
ON acq.swidtag = agg.swidtag AND acq.metric = agg.aggregation_metric WHERE scope = ANY(@scope::TEXT[])
GROUP BY agg.aggregation_id,agg.aggregation_name,agg.aggregation_metric,acq.product_editor,agg.aggregation_scope;

-- name: ListAcqRightsProducts :many
SELECT swidtag,product_name
FROM acqrights acq
WHERE swidtag NOT IN (SELECT UNNEST(products) from aggregations where aggregation_scope = @scope)
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

-- name: GetAcqRightsCost :one
SELECT SUM(total_cost)::Numeric(15,2) as total_cost,SUM(total_maintenance_cost)::Numeric(15,2) as total_maintenance_cost 
from acqrights 
WHERE scope = ANY(@scope::TEXT[])
GROUP BY scope; 

-- name: ProductsPerMetric :many
SELECT metric, COUNT(swidtag) as num_products
from acqrights
WHERE scope = ANY(@scope::TEXT[])
GROUP BY metric;

-- name: CounterFeitedProductsLicences :many
SELECT swidtag as swid_tag, 
product_name as product_name, 
SUM(num_licenses_acquired) as num_licenses_acquired,
SUM(num_licences_computed) as num_licences_computed,
SUM(num_licenses_acquired-num_licences_computed) as delta
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag,product_name
HAVING SUM(num_licenses_acquired-num_licences_computed) < 0
ORDER BY delta ASC LIMIT 5;

-- name: CounterFeitedProductsCosts :many
SELECT swidtag as swid_tag, 
product_name as product_name, 
SUM(total_purchase_cost)::Numeric(15,2) as  total_purchase_cost,
SUM(total_computed_cost)::Numeric(15,2) as total_computed_cost,
SUM(total_purchase_cost-total_computed_cost)::Numeric(15,2) as delta_cost
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag,product_name
HAVING SUM(total_purchase_cost-total_computed_cost) < 0
ORDER BY delta_cost ASC LIMIT 5;

-- name: OverDeployedProductsLicences :many
SELECT swidtag as swid_tag, 
product_name as product_name, 
SUM(num_licenses_acquired) as num_licenses_acquired,
SUM(num_licences_computed) as num_licences_computed,
SUM(num_licenses_acquired-num_licences_computed) as delta
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag,product_name
HAVING SUM(num_licenses_acquired-num_licences_computed) > 0
ORDER BY delta DESC LIMIT 5;

-- name: OverDeployedProductsCosts :many
SELECT swidtag as swid_tag, 
product_name as product_name, 
SUM(total_purchase_cost)::Numeric(15,2) as  total_purchase_cost,
SUM(total_computed_cost)::Numeric(15,2) as total_computed_cost,
SUM(total_purchase_cost-total_computed_cost)::Numeric(15,2) as delta_cost
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag,product_name
HAVING SUM(total_purchase_cost-total_computed_cost) > 0
ORDER BY delta_cost DESC LIMIT 5;

-- name: ListAcqrightsProducts :many
SELECT DISTINCT swidtag,scope
FROM acqrights;

-- name: AddComputedLicenses :exec
UPDATE 
  acqrights
SET 
  num_licences_computed = @computedLicenses,
  total_computed_cost = @computedCost
WHERE sku = @sku
AND scope = @scope;

-- name: CounterfeitPercent :one
SELECT tpc, delta_cost from (
SELECT
    sum(total_purchase_cost)::Numeric(15,2) as tpc,
    abs(sum(case when total_purchase_cost < total_computed_cost then total_purchase_cost - total_computed_cost else 0 end))::Numeric(15,2) as delta_cost
    from (
        select sum(total_purchase_cost) as total_purchase_cost,
        sum(total_computed_cost) as total_computed_cost
        from acqrights
        where scope= @scope
        group by swidtag
    ) y
)x WHERE tpc IS NOT NULL;

-- name: OverdeployPercent :one
SELECT tpc, delta_cost from (
SELECT
    sum(total_purchase_cost)::Numeric(15,2) as tpc,
    sum(case when total_purchase_cost > total_computed_cost then total_purchase_cost - total_computed_cost else 0 end)::Numeric(15,2) as delta_cost
    from (
        select sum(total_purchase_cost) as total_purchase_cost,
        sum(total_computed_cost) as total_computed_cost
        from acqrights
        where scope= @scope
        group by swidtag
    ) y
)x WHERE tpc IS NOT NULL;

-- name: UpsertProductApplications :exec

Insert into products_applications (swidtag, application_id,scope ) Values ($1,$2,$3) ON CONFLICT  (swidtag, application_id,scope)
Do NOTHING;

-- name: UpsertProductEquipments :exec

Insert into products_equipments (swidtag, equipment_id, num_of_users,scope ) Values ($1,$2,$3,$4 ) ON CONFLICT  (swidtag, equipment_id, scope)
Do Update set num_of_users = $3;

-- name: GetProductQualityOverview :one

select total_records,
       count(swid1) as not_acquired,
       count(swid2) as not_deployed,
       (count(swid1) * 100.0/ total_records) :: NUMERIC(5,2) as not_deployed_percentage,
       (count(swid2) * 100.0/ total_records) :: NUMERIC(5,2) as not_acquired_percentage
from
    (select count(*) over() as total_records, pe.swidtag as swid1, acq.swidtag as swid2 from products_equipments pe
     full outer join acqrights acq on acq.swidtag = pe.swidtag and acq.scope = pe.scope
     where acq."scope" = $1  or pe.scope = $1
     group by (2,3)) p
where swid1 is NULL or swid2 is null
group by (1);


-- name: ProductsNotDeployed :many
SELECT DISTINCT(swidtag), product_name FROM acqrights
WHERE acqrights.swidtag NOT IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND acqrights.scope = @scope;

-- name: ProductsNotAcquired :many
SELECT swidtag, product_name FROM products
WHERE products.swidtag NOT IN (SELECT swidtag FROM acqrights WHERE acqrights.scope = @scope)
AND products.swidtag IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND products.scope = @scope;

-- name: DeleteProductsByScope :exec
DELETE FROM products WHERE scope = @scope;

-- name: DeleteAcqrightsByScope :exec
DELETE FROM acqrights WHERE scope = @scope;

-- name: DeleteProductAggregationByScope :exec
DELETE FROM aggregations WHERE aggregation_scope = @scope;
