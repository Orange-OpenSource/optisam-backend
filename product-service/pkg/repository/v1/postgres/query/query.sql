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
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments,pe.equipment_ids, COALESCE(acq.total_cost,0)::FLOAT as cost 
FROM products p 
INNER JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[]) AND  (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END) GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments,  ARRAY_AGG(equipment_id)::TEXT[] as equipment_ids FROM products_equipments WHERE scope = ANY(@scope::TEXT[]) AND (CASE WHEN @is_equipment_id::bool THEN equipment_id = ANY(@equipment_ids::TEXT[]) ELSE TRUE END) GROUP BY swidtag) pe
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
  GROUP BY p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition,pa.num_of_applications, pe.num_of_equipments, pe.equipment_ids, acq.total_cost
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
SELECT p.swidtag,p.product_name,p.product_editor,p.product_version, acq.metrics, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications,COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments
FROM products p 
LEFT JOIN 
(SELECT pa.swidtag, count(pa.application_id) as num_of_applications FROM products_applications pa WHERE pa.scope = @scope GROUP BY pa.swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT pe.swidtag, count(pe.equipment_id) as num_of_equipments FROM products_equipments pe WHERE pe.scope = @scope GROUP BY pe.swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN
(SELECT ac.swidtag,ARRAY_AGG(DISTINCT acmetrics)::TEXT[] as metrics FROM acqrights ac, unnest(string_to_array(ac.metric,',')) as acmetrics WHERE ac.scope = @scope GROUP BY ac.swidtag) acq
ON p.swidtag = acq.swidtag
WHERE p.swidtag = @swidtag
AND p.scope = @scope
GROUP BY p.swidtag, p.product_name, p.product_editor, p.product_version, acq.metrics, pa.num_of_applications, pe.num_of_equipments;

-- name: GetProductInformationFromAcqright :one
SELECT ac.swidtag,
       ac.product_name,
       ac.product_editor,
       ac.version,
       ARRAY_AGG(DISTINCT acmetrics)::TEXT[] as metrics
FROM acqrights ac, unnest(string_to_array(ac.metric,',')) as acmetrics
WHERE ac.scope = @scope
    AND ac.swidtag = @swidtag
GROUP BY ac.swidtag,
         ac.product_name,
         ac.product_editor,
         ac.version;

-- name: GetProductOptions :many
SELECT p.swidtag,p.product_name,p.product_edition,p.product_editor,p.product_version
FROM products p 
WHERE p.option_of = @swidtag
AND p.scope = @scope;

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
SELECT x.product_editor as editor,
    ARRAY_AGG(DISTINCT x.product_name)::TEXT [] as product_names,
    ARRAY_AGG(DISTINCT x.swidtag)::TEXT [] as product_swidtags,
    ARRAY_AGG(DISTINCT x.version)::TEXT[] as product_versions,
    COALESCE(SUM(pa.num_of_applications),0)::INTEGER as num_of_applications,
    COALESCE(SUM(pe.num_of_equipments),0)::INTEGER as num_of_equipments,
    x.aggregation_name as aggregation_name,
    x.aggregation_id as aggregation_id,
    x.aggregation_metric as metric
FROM(
        SELECT  acq.swidtag,
             acq.product_name,
            acq.product_editor,
            acq.version,
            agg.aggregation_id,
            agg.aggregation_name,
            agg.aggregation_metric
        FROM aggregations agg
            JOIN acqrights acq ON acq.swidtag = ANY(agg.products)
        WHERE acq.scope = @scope
            AND agg.aggregation_scope = @scope
            AND agg.aggregation_id = @aggregation_id
            AND acq.metric = agg.aggregation_metric
        UNION
        SELECT prd.swidtag,
            prd.product_name,
            prd.product_editor,
            prd.product_version as version,
            agg.aggregation_id,
            agg.aggregation_name,
            agg.aggregation_metric
        FROM aggregations agg
            JOIN products prd ON prd.swidtag = ANY(agg.products)
        WHERE prd.scope = @scope
            AND agg.aggregation_scope = @scope
            AND agg.aggregation_id = @aggregation_id
            AND prd.aggregation_id = agg.aggregation_id
    ) x
LEFT JOIN
(SELECT pa.swidtag, count(pa.application_id) as num_of_applications FROM products_applications pa WHERE pa.scope = @scope GROUP BY pa.swidtag) pa
ON pa.swidtag = x.swidtag
LEFT JOIN 
(SELECT pe.swidtag, count(pe.equipment_id) as num_of_equipments FROM products_equipments pe WHERE pe.scope = @scope GROUP BY pe.swidtag) pe
ON pe.swidtag = x.swidtag
GROUP BY x.product_editor,
    x.aggregation_id,
    x.aggregation_name,
    x.aggregation_metric;

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
SELECT p.swidtag, p.product_name, p.product_version
FROM products p
JOIN 
(SELECT swidtag FROM products_equipments WHERE scope = ANY(@scopes::TEXT[]) GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
WHERE p.product_editor = @product_editor and p.scope = ANY(@scopes::TEXT[]);

-- name: UpsertProductAggregation :exec
-- SCOPE BASED CHANGE
Update products set aggregation_id = @aggregation_id, aggregation_name = @aggregation_name WHERE
swidtag = ANY(@swidtags::TEXT[]) AND scope = @scope;

-- name: GetProductAggregation :many
SELECT swidtag
FROM products
WHERE aggregation_id = @aggregation_id AND aggregation_name = @aggregation_name AND scope = @scope;

-- name: UpdateAggregationForProduct :exec
Update products set aggregation_id = @aggregation_id, aggregation_name = @aggregation_name WHERE
aggregation_id = @old_aggregation_id AND scope = @scope;

-- name: UpsertAcqRights :exec
INSERT INTO acqrights (sku,swidtag,product_name,product_editor,scope,metric,num_licenses_acquired,avg_unit_price,avg_maintenance_unit_price,total_purchase_cost,total_maintenance_cost,total_cost,created_by,start_of_maintenance,end_of_maintenance,num_licences_maintainance,version,comment)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
ON CONFLICT (sku,scope)
DO
UPDATE SET swidtag = $2,product_name = $3,product_editor = $4,scope = $5,metric = $6,num_licenses_acquired = $7,
            avg_unit_price = $8,avg_maintenance_unit_price = $9,total_purchase_cost = $10,
            total_maintenance_cost = $11,total_cost = $12,updated_by = $13,start_of_maintenance = $14,end_of_maintenance = $15, num_licences_maintainance = $16, version = $17, comment = $18;


-- name: ListAcqRightsIndividual :many
SELECT count(*) OVER() AS totalRecords,a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost ,start_of_maintenance, end_of_maintenance , version, comment FROM 
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

-- name: ListProductAggregation :many
SELECT
    DISTINCT ag.aggregation_name as aggregation_name,
   	ag.product_editor,
   	swidtags,
   	coalesce(pe.num_of_equipments,0) as num_of_equipments,
   	coalesce(pa.num_of_applications,0) as num_of_applications,
   	ag.id,
   	ag.total_cost
FROM
    aggregated_rights ag
    INNER JOIN (
        SELECT
            pa.swidtag,
            count(pa.application_id) as num_of_applications
        FROM
            products_applications pa
        WHERE
            pa.scope = @scope
        GROUP BY
            pa.swidtag
    ) pa ON pa.swidtag = ANY(ag.swidtags)
    INNER JOIN (
        SELECT
            pe.swidtag,
            count(pe.equipment_id) as num_of_equipments
        FROM
            products_equipments pe
        WHERE
            pe.scope = @scope
        GROUP BY
            pe.swidtag
    ) pe ON pe.swidtag = ANY(ag.swidtags)
WHERE
    scope = @scope
LIMIT
    @page_size OFFSET @page_num;

-- name: ListAcqRightsAggregation :many
SELECT count(*) OVER() AS totalRecords,* from aggregated_rights
WHERE scope = ANY(@scope::TEXT[]) LIMIT @page_size OFFSET @page_num;


-- name: ListAcqRightsAggregationIndividual :many
SELECT a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost, a.version FROM 
acqrights a
WHERE 
  a.swidtag IN (SELECT UNNEST(products) from aggregations where aggregation_id = @aggregation_id  AND a.metric = aggregation_metric AND aggregation_scope = ANY(@scope::TEXT[]))
  AND a.scope = ANY(@scope::TEXT[]);

-- name: ListProductsAggregationIndividual :many
SELECT * ,COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications,COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments
FROM products p 
LEFT JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = p.scope  GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE scope = p.scope  GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
LEFT JOIN 
(SELECT *  FROM aggregated_rights WHERE scope = p.scope  GROUP BY swidtag) ar
ON p.swidtag = ar.swidtag
WHERE p.aggregation_name = @aggregation_name and p.scope = ANY(@scope::TEXT[]);

-- name: UpsertAggregation :one
INSERT INTO aggregated_rights (
  aggregation_name,
  sku,
  scope,
  product_editor,
  metric,
  products,
  swidtags,
  num_licenses_acquired,
  avg_unit_price,
  avg_maintenance_unit_price,
  total_purchase_cost,
  total_maintenance_cost,
  total_cost,
  created_by,
  start_of_maintenance,
  end_of_maintenance,
  num_licences_maintainance,
  comment)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) ON CONFLICT (aggregation_name,sku,scope) DO
UPDATE
SET product_editor = $4,
    metric = $5,
    products = $6,
    swidtags = $7,
    num_licenses_acquired = $8,
    avg_unit_price = $9,
    avg_maintenance_unit_price = $10,
    total_purchase_cost = $11,
    total_maintenance_cost = $12,
    total_cost = $13,
    updated_by = $14,
    start_of_maintenance = $15,
    end_of_maintenance = $16,
    num_licences_maintainance = $17,
    comment = $18
RETURNING id;

-- name: UpdateAggregation :one
UPDATE aggregations
SET aggregation_name = @aggregation_name,products = @products
WHERE aggregation_id = @aggregation_id
AND aggregation_scope = @scope
RETURNING *;

-- name: DeleteAggregation :exec
DELETE FROM aggregated_rights 
WHERE id = @aggregation_id
AND scope = @scope;

-- name: ListAggregations :many
SELECT id,aggregation_name,sku,product_editor,metric,products,swidtags,scope,num_licenses_acquired,
  num_licences_computed,num_licences_maintainance,avg_unit_price,avg_maintenance_unit_price,total_purchase_cost,
  total_computed_cost,total_maintenance_cost,total_cost,start_of_maintenance,end_of_maintenance,comment
FROM aggregated_rights
WHERE scope = @scope;

-- name: ListProductsForAggregation :many
SELECT acq.swidtag, acq.product_name, acq.product_editor 
FROM acqrights acq 
WHERE acq.scope = @scope 
AND acq.swidtag NOT IN (SELECT UNNEST(agg.swidtags) from aggregated_rights agg where agg.scope = @scope and agg.metric = @metric) 
AND acq.metric = @metric
AND acq.product_editor = @editor
UNION
SELECT prd.swidtag, prd.product_name, prd.product_editor 
FROM products prd 
JOIN 
(SELECT swidtag FROM products_equipments WHERE scope = @scope GROUP BY swidtag) pe
ON prd.swidtag = pe.swidtag
WHERE prd.scope = @scope 
AND prd.swidtag NOT IN (SELECT UNNEST(agg.swidtags) from aggregated_rights agg where agg.scope = @scope and agg.metric = @metric)
AND prd.product_editor = @editor;

-- name: ListEditorsForAggregation :many
SELECT DISTINCT acq.product_editor 
FROM acqrights acq 
WHERE acq.scope = @scope AND acq.product_editor <> ''
UNION
SELECT DISTINCT prd.product_editor 
FROM products prd 
WHERE prd.scope = @scope AND prd.product_editor <> '';

-- name: ListMetricsForAggregation :many
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
num_licences_computed as num_licences_computed,
(SUM(num_licenses_acquired)-num_licences_computed) as delta
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag, metric, product_name,num_licences_computed
HAVING (SUM(num_licenses_acquired)-num_licences_computed) < 0
ORDER BY delta ASC LIMIT 5;

-- name: CounterFeitedProductsCosts :many
SELECT swidtag as swid_tag, 
product_name as product_name, 
SUM(total_purchase_cost)::Numeric(15,2) as  total_purchase_cost,
total_computed_cost::Numeric(15,2) as total_computed_cost,
(SUM(total_purchase_cost)-total_computed_cost)::Numeric(15,2) as delta_cost
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag, metric, product_name, total_computed_cost
HAVING (SUM(total_purchase_cost)-total_computed_cost) < 0
ORDER BY delta_cost ASC LIMIT 5;

-- name: OverDeployedProductsLicences :many
SELECT swidtag as swid_tag, 
product_name as product_name, 
SUM(num_licenses_acquired) as num_licenses_acquired,
num_licences_computed as num_licences_computed,
(SUM(num_licenses_acquired)-num_licences_computed) as delta
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag, metric, product_name,num_licences_computed
HAVING (SUM(num_licenses_acquired)-num_licences_computed) > 0
ORDER BY delta ASC LIMIT 5;

-- name: OverDeployedProductsCosts :many
SELECT swidtag as swid_tag, 
product_name as product_name, 
SUM(total_purchase_cost)::Numeric(15,2) as  total_purchase_cost,
total_computed_cost::Numeric(15,2) as total_computed_cost,
(SUM(total_purchase_cost)-total_computed_cost)::Numeric(15,2) as delta_cost
FROM acqrights 
WHERE
scope = @scope
AND
product_editor = @product_editor
GROUP BY swidtag, metric, product_name, total_computed_cost
HAVING (SUM(total_purchase_cost)-total_computed_cost) > 0
ORDER BY delta_cost DESC LIMIT 5;

-- name: ListAcqrightsProducts :many
SELECT DISTINCT swidtag,scope
FROM acqrights;

-- name: ListAcqrightsProductsByScope :many
SELECT DISTINCT swidtag,scope
FROM acqrights where scope = $1;

-- name: AddComputedLicenses :exec
UPDATE 
  acqrights
SET 
  num_licences_computed = @computedLicenses,
  total_computed_cost = @computedCost
WHERE sku = @sku
AND scope = @scope;

-- name: CounterfeitPercent :one
SELECT acq, delta_rights from (
SELECT
    sum(num_licenses_acquired)::Numeric(15,2) as acq,
    abs(sum(case when num_licenses_acquired < total_computed_licenses then num_licenses_acquired - total_computed_licenses else 0 end))::Numeric(15,2) as delta_rights
    from (
        select sum(num_licenses_acquired) as num_licenses_acquired,
        num_licences_computed as total_computed_licenses
        from acqrights
        where scope= @scope AND metric = ANY(@metrics::TEXT[])
        group by swidtag,metric, num_licences_computed
    ) y
)x WHERE acq IS NOT NULL;

-- name: OverdeployPercent :one
SELECT acq, delta_rights from (
SELECT
    sum(num_licenses_acquired)::Numeric(15,2) as acq,
    abs(sum(case when num_licenses_acquired > total_computed_licenses then num_licenses_acquired - total_computed_licenses else 0 end))::Numeric(15,2) as delta_rights
    from (
        select sum(num_licenses_acquired) as num_licenses_acquired,
        num_licences_computed as total_computed_licenses
        from acqrights
        where scope= @scope AND metric = ANY(@metrics::TEXT[])
        group by swidtag,metric, num_licences_computed
    ) y
)x WHERE acq IS NOT NULL;

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
SELECT DISTINCT(swidtag), product_name, product_editor, version FROM acqrights
WHERE acqrights.swidtag NOT IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND acqrights.scope = @scope;

-- name: ProductsNotAcquired :many
SELECT swidtag, product_name, product_editor, product_version  FROM products
WHERE products.swidtag NOT IN (SELECT swidtag FROM acqrights WHERE acqrights.scope = @scope)
AND products.swidtag IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND products.scope = @scope;

-- name: DeleteProductsByScope :exec
DELETE FROM products WHERE scope = @scope;

-- name: DeleteAcqrightsByScope :exec
DELETE FROM acqrights WHERE scope = @scope;

-- name: DeleteProductAggregationByScope :exec
DELETE FROM aggregated_rights WHERE scope = @scope;

-- name: GetAggregationByName :one
SELECT *
FROM aggregated_rights
WHERE aggregation_name = @aggregation_name
    AND scope = @scope;

-- name: GetAggregationBySKU :one
SELECT *
FROM aggregated_rights
WHERE sku = @sku
    AND scope = @scope;

-- name: GetAcqRightBySKU :one
SELECT *
FROM acqrights
WHERE sku = @acqright_sku and scope = @scope;

-- name: InsertAcqRight :exec
INSERT INTO acqrights (sku,swidtag,product_name,product_editor,scope,metric,num_licenses_acquired,avg_unit_price,avg_maintenance_unit_price,total_purchase_cost,total_maintenance_cost,total_cost,created_by,start_of_maintenance,end_of_maintenance,num_licences_maintainance,version,comment)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18);

-- name: UpsertDashboardUpdates :exec
Insert into dashboard_audit (updated_at, next_update_at ,updated_by ,scope) values($1,$2,$3,$4)
on CONFLICT (scope) 
Do update set updated_at = $1, next_update_at = $2, updated_by = $3;

-- name: GetDashboardUpdates :one
select updated_at  at time zone $2::varchar as updated_at , next_update_at  at time zone $2::varchar as next_update_at from dashboard_audit where scope = $1;

-- name: DeleteAcqrightBySKU :exec
DELETE FROM acqrights WHERE scope = @scope AND sku = @sku;

-- name: GetEquipmentsBySwidtag :one
SELECT 
    ARRAY_AGG(DISTINCT(equipment_id))::TEXT[] as equipments
from products_equipments
WHERE 
    scope = @scope and 
    swidtag = @swidtag;

-- name: GetAcqBySwidtags :many
Select * from acqrights where swidtag = ANY(@swidtag::TEXT[]) and scope = @scope;

-- name: GetIndividualProductDetailByAggregation :many
Select
	ar.aggregation_name,
    coalesce(num_of_applications,0) as num_of_applications,
    coalesce(num_of_equipments,0) as num_of_equipments,
    ar.product_editor,
    name,
    version,
    p_id,
    ar.total_cost
from
	aggregated_rights ar
	LEFT JOIN (
		select 
			p.product_name as name,
			p.product_version as version,
			p.swidtag  as p_id
			from products p 
		where 
			p.scope =  @scope
	) p on p_id  = ANY(ar.swidtags::TEXT[])
    LEFT JOIN (
        SELECT
            pa.swidtag ,
            count(application_id) as num_of_applications
        FROM
            products_applications pa
        WHERE
            pa.scope = @scope
        GROUP BY
            pa.swidtag
    ) pa ON p_id = pa.swidtag
    LEFT JOIN (
        SELECT
            pe.swidtag ,
            count(equipment_id) as num_of_equipments
        FROM
            products_equipments pe
        WHERE
            pe.scope = @scope
        GROUP BY
            pe.swidtag
    ) pe ON p_id = pe.swidtag
WHERE
     ar.scope = @scope and ar.aggregation_name = @aggregation_name;

-- name: ListSelectedProductsForAggregration :many
SELECT acq.swidtag, acq.product_name, acq.product_editor
FROM acqrights acq 
WHERE acq.scope = @scope 
AND acq.swidtag IN (SELECT UNNEST(agg.swidtags) from aggregated_rights agg where agg.scope = @scope and agg.id = @id) 
UNION
SELECT prd.swidtag, prd.product_name, prd.product_editor
FROM products prd 
WHERE prd.scope = @scope 
AND prd.swidtag IN (SELECT UNNEST(agg.swidtags) from aggregated_rights agg where agg.scope = @scope and agg.id = @id);

-- name: GetAggregationByID :one
SELECT *
FROM aggregated_rights
WHERE id = @id
    AND scope = @scope;
