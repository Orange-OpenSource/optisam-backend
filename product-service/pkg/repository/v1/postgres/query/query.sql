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
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, pe.allocated_metric, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.equipment_users,0)::INTEGER as equipment_users, COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments, COALESCE(acq.total_cost,0)::FLOAT as cost 
FROM products p 
LEFT JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE scope = ANY(@scope::TEXT[]) AND (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END)  GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
INNER JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments, sum(num_of_users) as equipment_users, allocated_metric FROM products_equipments WHERE  scope = ANY(@scope::TEXT[]) AND (CASE WHEN @is_equipment_id::bool THEN equipment_id = @equipment_id ELSE TRUE END)  GROUP BY swidtag, allocated_metric) pe
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
  GROUP BY p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition,pa.num_of_applications, pe.equipment_users, pe.allocated_metric, pe.num_of_equipments, acq.total_cost
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

-- name: ListProductsByApplicationInstance :many
Select
    count(*) OVER() AS totalRecords,
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_category,
    p.product_editor,
    p.product_edition,
    COALESCE(acq.total_cost, 0) :: FLOAT as total_cost
from
    products p
    LEFT JOIN (
        SELECT
            swidtag,
            sum(total_cost) as total_cost
        FROM
            acqrights
        WHERE
            scope = ANY(@scope::TEXT[])
        GROUP BY
            swidtag
    ) acq ON p.swidtag = acq.swidtag
where
  scope = ANY(@scope::TEXT[])
  AND p.swidtag = ANY(@swidtag::TEXT[])
  AND (CASE WHEN @lk_product_name::bool THEN lower(p.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(p.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
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
  CASE WHEN @total_cost_asc::bool THEN acq.total_cost END asc,
  CASE WHEN @total_cost_desc::bool THEN acq.total_cost END desc
  LIMIT @page_size OFFSET @page_num;

-- name: GetProductInformation :one
SELECT p.swidtag,p.product_name,p.product_editor,p.product_version, acq.metrics, COALESCE(papp.num_of_applications,0)::INTEGER as num_of_applications,COALESCE(peq.num_of_equipments,0)::INTEGER as num_of_equipments
FROM products p 
LEFT JOIN 
(SELECT pa.swidtag, count(pa.application_id) as num_of_applications FROM products_applications pa WHERE pa.scope = @scope GROUP BY pa.swidtag) papp
ON p.swidtag = papp.swidtag
LEFT JOIN 
(SELECT pe.swidtag, count(pe.equipment_id) as num_of_equipments FROM products_equipments pe WHERE pe.scope = @scope GROUP BY pe.swidtag) peq
ON p.swidtag = peq.swidtag
LEFT JOIN
(SELECT ac.swidtag,ARRAY_AGG(DISTINCT acmetrics)::TEXT[] as metrics FROM acqrights ac, unnest(string_to_array(ac.metric,',')) as acmetrics WHERE ac.scope = @scope GROUP BY ac.swidtag) acq
ON p.swidtag = acq.swidtag
WHERE p.swidtag = @swidtag
AND p.scope = @scope
GROUP BY p.swidtag, p.product_name, p.product_editor, p.product_version, acq.metrics, papp.num_of_applications, peq.num_of_equipments;

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

-- name: AggregatedRightDetails :one
SELECT agg.aggregation_name,
       agg.product_editor,
       agg.products as product_names,
       agg.swidtags as product_swidtags,
       COALESCE(
                    (SELECT y.metrics
                     FROM
                         (SELECT ARRAY_AGG(DISTINCT armetrics)::TEXT[] as metrics
                          FROM aggregated_rights ar, unnest(string_to_array(ar.metric,',')) as armetrics 
                          WHERE ar.scope = @scope
                              AND ar.aggregation_id = @id ) y),'{}')::TEXT[] as metrics,
       COALESCE(
                    (SELECT sum(y.num_of_applications)
                     FROM
                         (SELECT count(DISTINCT pa.application_id) as num_of_applications
                          FROM products_applications pa
                          WHERE pa.scope = @scope
                              AND pa.swidtag = ANY(agg.swidtags) ) y),0)::INTEGER as num_of_applications,
       COALESCE(
                    (SELECT sum(z.num_of_equipments)
                     FROM
                         (SELECT count(DISTINCT pe.equipment_id) as num_of_equipments
                          FROM products_equipments pe
                          WHERE pe.scope = @scope
                              AND pe.swidtag = ANY(agg.swidtags) ) z),0)::INTEGER as num_of_equipments,
       COALESCE(
                    (SELECT ARRAY_AGG(DISTINCT x.version)
                     FROM
                          (SELECT acq.version as version
                            FROM acqrights acq
                            WHERE acq.swidtag = ANY(agg.swidtags)
                            AND acq.scope = @scope
                            UNION SELECT prd.product_version as version
                            FROM products prd
                            WHERE prd.swidtag = ANY(agg.swidtags)
                            AND prd.scope = @scope ) x),'{}') ::TEXT[] as product_versions
FROM aggregations agg
WHERE scope = @scope
    AND id = @id
GROUP BY agg.aggregation_name,
         agg.product_editor,
         agg.products,
         agg.swidtags;

-- -- name: ProductAggregationChildOptions :many
-- SELECT p.swidtag,p.product_name,p.product_edition,p.product_editor,p.product_version
-- FROM products p 
-- WHERE p.option_of in (
-- SELECT p.swidtag
-- FROM products p
-- WHERE 
--   p.aggregation_id = @aggregation_id
--   AND p.scope = ANY(@scope::TEXT[]))
-- AND p.scope = ANY(@scope::TEXT[]) ;

-- name: UpsertProduct :exec
INSERT INTO products (swidtag, product_name, product_version, product_edition, product_category, product_editor,scope,option_of,created_on,created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (swidtag,scope)
DO
 UPDATE SET product_name = $2, product_version = $3, product_edition = $4,product_category = $5,product_editor= $6,option_of=$8,updated_on=$11,updated_by=$12;

-- name: UpsertProductPartial :exec
INSERT INTO products (swidtag,scope,created_by)
VALUES ($1,$2,$3)
ON CONFLICT (swidtag,scope)
DO NOTHING;


-- name: DeleteProductApplications :exec
DELETE FROM products_applications
WHERE swidtag = @product_id and application_id = ANY(@application_id::TEXT[]) and scope = @scope;

-- name: DeleteProductEquipments :exec
DELETE FROM products_equipments
WHERE swidtag = @product_id and equipment_id = ANY(@equipment_id::TEXT[]) and scope = @scope;

-- name: GetProductsByEditor :many
SELECT p.swidtag, p.product_name, p.product_version
FROM products p
JOIN 
(SELECT swidtag FROM products_equipments WHERE scope = ANY(@scopes::TEXT[]) GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
WHERE p.product_editor = @product_editor and p.scope = ANY(@scopes::TEXT[]);

-- name: UpsertAcqRights :exec
INSERT INTO acqrights (
    sku,
    swidtag,
    product_name,
    product_editor,
    scope,
    metric,
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
    version,
    comment,
    last_purchased_order,
    support_number,
    maintenance_provider,
    ordering_date,
    corporate_sourcing_contract,
    software_provider,
    file_name,
    file_data,
    repartition)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27) ON CONFLICT (sku,scope) DO
UPDATE
SET swidtag = $2,
    product_name = $3,
    product_editor = $4,
    scope = $5,
    metric = $6,
    num_licenses_acquired = $7,
    avg_unit_price = $8,
    avg_maintenance_unit_price = $9,
    total_purchase_cost = $10,
    total_maintenance_cost = $11,
    total_cost = $12,
    updated_by = $13,
    start_of_maintenance = $14,
    end_of_maintenance = $15,
    num_licences_maintainance = $16,
    version = $17,
    comment = $18,
    last_purchased_order = $19,
    support_number = $20,
    maintenance_provider = $21,
    ordering_date = $22,
    corporate_sourcing_contract = $23,
    software_provider = $24,
    file_name = $25,
    file_data = $26,
    repartition = $27;


-- name: ListAcqRightsIndividual :many
SELECT count(*) OVER() AS totalRecords,a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost ,a.start_of_maintenance, a.end_of_maintenance , a.version, a.comment, a.ordering_date, a.software_provider, a.corporate_sourcing_contract, a.last_purchased_order, a.support_number, a.maintenance_provider, a.file_name, a.repartition
FROM 
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
  AND (CASE WHEN @is_ordering_date::bool THEN a.ordering_date <= @ordering_date ELSE TRUE END)
  AND (CASE WHEN @lk_software_provider::bool THEN lower(a.software_provider) LIKE '%' || lower(@software_provider::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_software_provider::bool THEN lower(a.software_provider) = lower(@software_provider) ELSE TRUE END)
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
Select
    DISTINCT agg.aggregation_name,
    agg.product_editor,
    agg.swidtags,
    COALESCE(
                    (SELECT sum(y.num_of_applications)
                     FROM
                         (SELECT count(DISTINCT pa.application_id) as num_of_applications
                          FROM products_applications pa
                          WHERE pa.scope = @scope
                              AND pa.swidtag = ANY(agg.swidtags) ) y),0)::INTEGER as num_of_applications,
       COALESCE(
                    (SELECT sum(z.num_of_equipments)
                     FROM
                         (SELECT count(DISTINCT pe.equipment_id) as num_of_equipments
                          FROM products_equipments pe
                          WHERE pe.scope = @scope
                              AND pe.swidtag = ANY(agg.swidtags) ) z),0)::INTEGER as num_of_equipments,
    agg.id,
    COALESCE(ar.total_cost,0)::NUMERIC(15,2) as total_cost
from
    aggregations agg
    LEFT JOIN (
        select
            p.product_name as name,
            p.product_version as version,
            p.swidtag as p_id
        from
            products p
        where
            p.scope = @scope
    ) p on p_id = ANY(agg.swidtags :: TEXT [])
     LEFT JOIN (
        SELECT
            a.aggregation_id,
            sum(a.total_cost)::NUMERIC(15,2) as total_cost
        FROM
            aggregated_rights a
        WHERE
            a.scope = @scope
        GROUP BY
            a.aggregation_id
    ) ar ON agg.id = ar.aggregation_id
WHERE
    agg.scope = @scope
GROUP BY
    agg.aggregation_name,
    agg.product_editor,
    agg.swidtags,
    agg.id,
    ar.total_cost
LIMIT
    @page_size OFFSET @page_num;

-- name: ListAcqRightsAggregation :many
SELECT count(*) OVER() AS totalRecords,
    a.sku,
    a.aggregation_id, 
    a.metric,
    a.ordering_date,
    a.corporate_sourcing_contract,
    a.software_provider,
    a.scope,
    a.num_licenses_acquired,
    a.num_licences_computed,
    a.num_licences_maintenance,
    a.avg_unit_price,
    a.avg_maintenance_unit_price,
    a.total_purchase_cost,
    a.total_computed_cost,
    a.total_maintenance_cost,
    a.total_cost,
    a.start_of_maintenance,
    a.end_of_maintenance,
    a.last_purchased_order,
    a.support_number,
    a.maintenance_provider,
    a.comment,
    a.created_on,
    a.created_by,
    a.updated_on,
    a.updated_by,
    a.file_name,
    a.repartition,
    agg.aggregation_name,
    agg.product_editor,
    agg.products,
    agg.swidtags
 from aggregated_rights a
  LEFT JOIN (
        SELECT
            ag.id,
            ag.aggregation_name, 
            ag.scope,
            ag.product_editor,
            ag.products,
            ag.swidtags
        FROM
            aggregations ag
        WHERE
            ag.scope = @scope
        GROUP BY
            ag.id
    ) agg ON agg.id = a.aggregation_id
WHERE a.scope = @scope
  AND (CASE WHEN @is_agg_name::bool THEN lower(agg.aggregation_name) = lower(@aggregation_name) ELSE TRUE END)
  AND (CASE WHEN @lk_agg_name::bool THEN lower(agg.aggregation_name) LIKE '%' || lower(@aggregation_name) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_editor::bool THEN lower(agg.product_editor) = lower(@product_editor) ELSE TRUE END)
  AND (CASE WHEN @lk_editor::bool THEN lower(agg.product_editor) LIKE '%' || lower(@product_editor) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_sku::bool THEN lower(a.sku) = lower(@sku) ELSE TRUE END)
  AND (CASE WHEN @lk_sku::bool THEN lower(a.sku) LIKE '%' || lower(@sku::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_metric::bool THEN lower(@metric) = any(string_to_array(lower(a.metric) , ',') :: text[]) ELSE TRUE END)
  AND (CASE WHEN @lk_metric::bool THEN lower(a.metric) LIKE '%' || lower(@metric::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_ordering_date::bool THEN a.ordering_date <= @ordering_date ELSE TRUE END)
  AND (CASE WHEN @lk_software_provider::bool THEN lower(a.software_provider) LIKE '%' || lower(@software_provider::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_software_provider::bool THEN lower(a.software_provider) = lower(@software_provider) ELSE TRUE END)
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN ARRAY_Length(agg.swidtags, 1) END asc,
  CASE WHEN @swidtag_desc::bool THEN ARRAY_Length(agg.swidtags, 1) END desc,
  CASE WHEN @agg_name_asc::bool THEN agg.aggregation_name END asc,
  CASE WHEN @agg_name_desc::bool THEN agg.aggregation_name END desc,
  CASE WHEN @product_editor_asc::bool THEN agg.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN agg.product_editor END desc,
  CASE WHEN @sku_asc::bool THEN a.sku END asc,
  CASE WHEN @sku_desc::bool THEN a.sku END desc,
  CASE WHEN @metric_asc::bool THEN a.metric END asc,
  CASE WHEN @metric_desc::bool THEN a.metric END desc,
  CASE WHEN @num_licenses_acquired_asc::bool THEN a.num_licenses_acquired END asc,
  CASE WHEN @num_licenses_acquired_desc::bool THEN a.num_licenses_acquired END desc,
  CASE WHEN @num_licences_maintenance_asc::bool THEN a.num_licences_maintenance END asc,
  CASE WHEN @num_licences_maintenance_desc::bool THEN a.num_licences_maintenance END desc,
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
  CASE WHEN @end_of_maintenance_desc::bool THEN a.end_of_maintenance END desc,
  CASE WHEN @license_under_maintenance_asc::bool THEN age(end_of_maintenance, start_of_maintenance) END desc,
  CASE WHEN @license_under_maintenance_desc::bool  THEN age(end_of_maintenance, start_of_maintenance) END asc
LIMIT @page_size OFFSET @page_num;


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
(SELECT *  FROM aggregations WHERE scope = p.scope) ar
ON p.swidtag = ar.swidtag
WHERE p.aggregation_name = @aggregation_name and p.scope = ANY(@scope::TEXT[]);

-- name: InsertAggregation :one
INSERT INTO aggregations (aggregation_name, scope, product_editor, products, swidtags, created_by)
VALUES ($1,
        $2,
        $3,
        $4,
        $5,
        $6) RETURNING id;

-- name: UpsertAggregatedRights :exec
INSERT INTO aggregated_rights (
  sku,
  scope,
  aggregation_id,
  metric,
  ordering_date,
  corporate_sourcing_contract,
  software_provider,
  num_licenses_acquired,
  avg_unit_price,
  avg_maintenance_unit_price,
  total_purchase_cost,
  total_maintenance_cost,
  total_cost,
  created_by,
  start_of_maintenance,
  end_of_maintenance,
  last_purchased_order,
  support_number,
  maintenance_provider,
  num_licences_maintenance,
  comment,
  file_name,
  file_data,
  repartition)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24) ON CONFLICT (sku,scope) DO
UPDATE
SET aggregation_id = $3,
    metric = $4,
    ordering_date = $5,
    corporate_sourcing_contract = $6,
    software_provider = $7,
    num_licenses_acquired = $8,
    avg_unit_price = $9,
    avg_maintenance_unit_price = $10,
    total_purchase_cost = $11,
    total_maintenance_cost = $12,
    total_cost = $13,
    updated_by = $14,
    start_of_maintenance = $15,
    end_of_maintenance = $16,
    last_purchased_order = $17,
    support_number = $18,
    maintenance_provider = $19,
    num_licences_maintenance = $20,
    comment = $21,
    file_name = $22,
    file_data = $23,
    repartition = $24;

-- name: UpdateAggregation :exec
UPDATE aggregations
SET aggregation_name = @aggregation_name,
    product_editor = @product_editor,
    products = @product_names,
    swidtags = @swidtags,
    updated_by = @updated_by
WHERE id = @id
    AND scope = @scope;

-- name: DeleteAggregation :exec
DELETE
FROM aggregations
WHERE id = @id
    AND scope = @scope;

-- name: ListAggregations :many
SELECT id,
       aggregation_name,
       product_editor,
       products,
       swidtags,
       scope
FROM aggregations
WHERE scope = @scope
    AND (CASE WHEN @is_agg_name::bool THEN lower(aggregation_name) = lower(@aggregation_name) ELSE TRUE END)
    AND (CASE WHEN @ls_agg_name::bool THEN lower(aggregation_name) LIKE '%' || lower(@aggregation_name) || '%' ELSE TRUE END)
    AND (CASE WHEN @is_product_editor::bool THEN lower(product_editor) = lower(@product_editor) ELSE TRUE END)
    AND (CASE WHEN @lk_product_editor::bool THEN lower(product_editor) LIKE '%' || lower(@product_editor) || '%' ELSE TRUE END)
    ORDER BY
    CASE WHEN @agg_name_asc::bool THEN aggregation_name END asc,
    CASE WHEN @agg_name_desc::bool THEN aggregation_name END desc,
    CASE WHEN @product_editor_asc::bool THEN product_editor END asc,
    CASE WHEN @product_editor_desc::bool THEN product_editor END desc
  LIMIT @page_size OFFSET @page_num;

-- name: ListProductsForAggregation :many
SELECT acq.swidtag, acq.product_name, acq.product_editor 
FROM acqrights acq 
WHERE acq.scope = @scope 
AND acq.swidtag NOT IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope) 
AND acq.product_editor = @editor
UNION
SELECT prd.swidtag, prd.product_name, prd.product_editor 
FROM products prd 
WHERE prd.scope = @scope 
AND prd.swidtag NOT IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope)
AND prd.product_editor = @editor;

-- name: ListEditorsForAggregation :many
SELECT DISTINCT acq.product_editor 
FROM acqrights acq 
WHERE acq.scope = @scope AND acq.product_editor <> ''
UNION
SELECT DISTINCT prd.product_editor 
FROM products prd 
WHERE prd.scope = @scope AND prd.product_editor <> '';

-- name: ListDeployedAndAcquiredEditors :many
SELECT DISTINCT acq.product_editor 
FROM acqrights acq 
WHERE acq.scope = @scope AND acq.product_editor <> ''
INTERSECT
SELECT DISTINCT prd.product_editor 
FROM products prd 
WHERE prd.scope = @scope AND prd.product_editor <> '';

-- name: GetAcqRightsByEditor :many
SELECT 
      acq.sku,
      acq.swidtag,
      acq.metric,
      acq.avg_unit_price :: FLOAT AS avg_unit_price,
      acq.num_licenses_acquired
FROM acqrights acq
WHERE acq.product_editor = @product_editor AND acq.scope = @scope;
        
-- name: GetAggregationByEditor :many
SELECT
    agg.aggregation_name,
    array_to_string(agg.swidtags,',') as swidtags,
    COALESCE(agr.sku,'') as sku, 
    COALESCE(agr.metric,'') as metric,
    COALESCE(agr.avg_unit_price,0) :: FLOAT AS avg_unit_price,
    COALESCE(agr.num_licenses_acquired,0)
FROM 
    aggregations agg
LEFT JOIN (
    SELECT
        a.aggregation_id,
        a.sku,
        a.metric,
        a.avg_unit_price,
        a.num_licenses_acquired
    FROM
        aggregated_rights a
    WHERE
        a.scope = @scope 
) agr ON agr.aggregation_id = agg.id
WHERE agg.scope = @scope AND agg.product_editor = @product_editor;

-- name: ListMetricsForAggregation :many
SELECT DISTINCT acq.metric
FROM acqrights acq
WHERE acq.scope = $1;

-- name: GetLicensesCost :one
SELECT
    COALESCE(SUM(total_cost),0) :: Numeric(15, 2) as total_cost,
    COALESCE(SUM(total_maintenance_cost),0) :: Numeric(15, 2) as total_maintenance_cost
FROM
    (
        SELECT
            SUM(total_cost) :: Numeric(15, 2) as total_cost,
            SUM(total_maintenance_cost) :: Numeric(15, 2) as total_maintenance_cost
        FROM
            acqrights
        WHERE
            scope = ANY(@scope::TEXT[])
        GROUP BY
            scope
        UNION ALL
        SELECT
            SUM(total_cost) :: Numeric(15, 2) as total_cost,
            SUM(total_maintenance_cost) :: Numeric(15, 2) as total_maintenance_cost
        FROM
            aggregated_rights
        WHERE
            scope = ANY(@scope::TEXT[])
        GROUP BY
            scope
    ) a;
-- name: GetAcqRightsCost :one
SELECT SUM(total_cost)::Numeric(15,2) as total_cost,SUM(total_maintenance_cost)::Numeric(15,2) as total_maintenance_cost 
from acqrights 
WHERE scope = ANY(@scope::TEXT[])
GROUP BY scope; 

-- name: ProductsPerMetric :many
SELECT x.metric as metric,
SUM(x.composition) as composition
FROM(
SELECT acq.metric, count(acq.swidtag) as composition
FROM acqrights acq Where acq.scope = @scope
GROUP BY acq.metric
UNION ALL
SELECT agr.metric, count(agr.aggregation_id) as composition
FROM aggregated_rights agr Where agr.scope = @scope
GROUP BY agr.metric
)x
GROUP BY x.metric;



-- name: CounterFeitedProductsLicences :many
SELECT swidtags as swid_tags, 
product_names as product_names,
aggregation_name as aggregation_name,
num_computed_licences::Numeric(15,2) as num_computed_licences,
num_acquired_licences::Numeric(15,2) as num_acquired_licences,
(delta_number)::Numeric(15,2) as delta
FROM overall_computed_licences
WHERE scope = @scope AND editor = @editor AND cost_optimization= FALSE
GROUP BY swidtags,product_names,aggregation_name,num_acquired_licences,num_computed_licences,delta_number
HAVING
    (delta_number) < 0
ORDER BY
    delta ASC
LIMIT
    5;

-- name: CounterFeitedProductsCosts :many
SELECT swidtags as swid_tags, 
product_names as product_names,
aggregation_name as aggregation_name,
computed_cost::Numeric(15,2) as computed_cost,
purchase_cost::Numeric(15,2) as purchase_cost,
(delta_cost)::Numeric(15,2) as delta_cost
FROM overall_computed_licences
WHERE scope = @scope AND editor = @editor AND cost_optimization= FALSE
GROUP BY swidtags,product_names,aggregation_name,purchase_cost,computed_cost,delta_cost
HAVING
    (delta_cost) < 0
ORDER BY
    delta_cost ASC
LIMIT
    5;

-- name: OverDeployedProductsLicences :many
SELECT swidtags as swid_tags, 
product_names as product_names,
aggregation_name as aggregation_name,
num_computed_licences::Numeric(15,2) as num_computed_licences,
num_acquired_licences::Numeric(15,2) as num_acquired_licences,
(delta_number)::Numeric(15,2) as delta
FROM overall_computed_licences
WHERE scope = @scope AND editor = @editor AND cost_optimization= FALSE
GROUP BY swidtags,product_names,aggregation_name,num_acquired_licences,num_computed_licences,delta_number
HAVING
    (delta_number) > 0
ORDER BY
    delta ASC
LIMIT
    5;

-- name: OverDeployedProductsCosts :many
SELECT swidtags as swid_tags, 
product_names as product_names,
aggregation_name as aggregation_name,
computed_cost::Numeric(15,2) as computed_cost,
purchase_cost::Numeric(15,2) as purchase_cost,
(delta_cost)::Numeric(15,2) as delta_cost
FROM overall_computed_licences
WHERE scope = @scope AND editor = @editor AND cost_optimization= FALSE
GROUP BY swidtags,product_names,aggregation_name,purchase_cost,computed_cost,delta_cost
HAVING
    (delta_cost) > 0
ORDER BY
    delta_cost ASC
LIMIT
    5;

-- name: ListAcqrightsProducts :many
SELECT DISTINCT swidtag,scope
FROM acqrights;

-- name: ListAcqrightsProductsByScope :many
SELECT DISTINCT swidtag,scope
FROM acqrights where acqrights.scope = $1  and swidtag not in ( select unnest(aggregations.swidtags) from aggregations inner join aggregated_rights on aggregations.id = aggregated_rights.aggregation_id and aggregations.scope = $1);

-- name: AddComputedLicenses :exec
UPDATE 
  acqrights
SET 
  num_licences_computed = @computedLicenses,
  total_computed_cost = @computedCost
WHERE sku = @sku
AND scope = @scope;

-- name: CounterfeitPercent :one
select
coalesce(sum(num_acquired_licences),0)::Numeric(15,2) as acq,
coalesce(abs(sum(delta_number)), 0)::Numeric(15,2) as delta_rights
from overall_computed_licences ocl
where ocl.scope = @scope AND ocl.delta_number < 0;

-- name: OverdeployPercent :one
select coalesce(sum(num_acquired_licences),0)::Numeric(15,2) as acq,
coalesce(abs(sum(delta_number)), 0)::Numeric(15,2) as delta_rights
from overall_computed_licences ocl
where ocl.scope = @scope AND ocl.delta_number > 0;


-- name: UpsertProductApplications :exec
Insert into products_applications (swidtag, application_id,scope ) Values ($1,$2,$3) ON CONFLICT  (swidtag, application_id,scope)
Do NOTHING;

-- name: UpsertProductEquipments :exec
Insert into products_equipments (swidtag, equipment_id, num_of_users,scope, allocated_metric ) Values ($1,$2,$3,$4,$5 ) ON CONFLICT  (swidtag, equipment_id, scope)
Do Update set num_of_users = $3, allocated_metric = $5;

-- name: ProductsNotDeployed :many
SELECT DISTINCT(swidtag), product_name, product_editor, version FROM acqrights
WHERE acqrights.swidtag NOT IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND acqrights.scope = @scope;

-- name: ProductsNotAcquired :many
SELECT swidtag, product_name, product_editor, product_version  FROM products
WHERE products.swidtag NOT IN (SELECT swidtag FROM acqrights WHERE acqrights.scope = @scope UNION SELECT unnest(swidtags) as swidtags from aggregations INNER JOIN aggregated_rights ON aggregations.id = aggregated_rights.aggregation_id where aggregations.scope = @scope)
AND products.swidtag IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND products.scope = @scope;

-- name: DeleteProductsByScope :exec
DELETE FROM products WHERE scope = @scope;

-- name: DeleteAcqrightsByScope :exec
DELETE FROM acqrights WHERE scope = @scope;

-- name: DeleteAggregationByScope :exec
DELETE FROM aggregations WHERE scope = @scope;

-- name: DeleteAggregatedRightsByScope :exec
DELETE FROM aggregated_rights WHERE scope = @scope;

-- name: GetAggregationByName :one
SELECT *
FROM aggregations
WHERE aggregation_name = @aggregation_name
    AND scope = @scope;

-- name: GetAggregatedRightBySKU :one
SELECT sku,
    aggregation_id, 
    metric,
    ordering_date,
    corporate_sourcing_contract,
    software_provider,
    scope,
    num_licenses_acquired,
    num_licences_computed,
    num_licences_maintenance,
    avg_unit_price,
    avg_maintenance_unit_price,
    total_purchase_cost,
    total_computed_cost,
    total_maintenance_cost,
    total_cost,
    start_of_maintenance,
    end_of_maintenance,
    last_purchased_order,
    support_number,
    maintenance_provider,
    comment,
    file_name
FROM aggregated_rights
WHERE sku = @sku
    AND scope = @scope;

-- name: GetAggregatedRightsFileDataBySKU :one
SELECT file_data
FROM aggregated_rights
WHERE sku = @sku
    and scope = @scope;

-- name: GetAcqRightBySKU :one
SELECT sku,
       swidtag,
       product_name,
       product_editor,
       metric,
       num_licenses_acquired,
       num_licences_maintainance,
       avg_unit_price,
       avg_maintenance_unit_price,
       total_purchase_cost,
       total_maintenance_cost,
       total_cost,
       start_of_maintenance,
       end_of_maintenance,
       version,
       comment,
       ordering_date,
       software_provider,
       corporate_sourcing_contract,
       last_purchased_order,
       support_number,
       maintenance_provider,
       file_name
FROM acqrights
WHERE sku = @acqright_sku
    and scope = @scope;

-- name: GetAcqRightFileDataBySKU :one
SELECT file_data
FROM acqrights
WHERE sku = @acqright_sku
    and scope = @scope;

-- name: UpsertDashboardUpdates :exec
Insert into dashboard_audit (updated_at, next_update_at ,updated_by ,scope) values($1,$2,$3,$4)
on CONFLICT (scope) 
Do update set updated_at = $1, next_update_at = $2, updated_by = $3;

-- name: GetDashboardUpdates :one
select updated_at  at time zone $2::varchar as updated_at , next_update_at  at time zone $2::varchar as next_update_at from dashboard_audit where scope = $1;

-- name: DeleteAcqrightBySKU :exec
DELETE FROM acqrights WHERE scope = @scope AND sku = @sku;

-- name: DeleteAggregatedRightBySKU :exec
DELETE FROM aggregated_rights WHERE scope = @scope AND sku = @sku;

-- name: GetEquipmentsBySwidtag :one
SELECT 
    ARRAY_AGG(DISTINCT(equipment_id))::TEXT[] as equipments
from products_equipments
WHERE 
    scope = @scope and 
    swidtag = @swidtag;

-- name: GetAcqBySwidtags :many
Select sku,
       swidtag,
       product_name,
       product_editor,
       metric,
       num_licenses_acquired,
       num_licences_maintainance,
       avg_unit_price,
       avg_maintenance_unit_price,
       total_purchase_cost,
       total_maintenance_cost,
       total_cost,
       start_of_maintenance,
       end_of_maintenance,
       version,
       comment,
       ordering_date,
       software_provider,
       corporate_sourcing_contract,
       last_purchased_order,
       support_number,
       maintenance_provider,
       file_name
from acqrights where swidtag = ANY(@swidtag::TEXT[]) and scope = @scope
 AND (CASE WHEN @is_metric::bool THEN lower(metric) = lower(@metric) ELSE TRUE END);

-- name: GetIndividualProductDetailByAggregation :many
Select
	agg.aggregation_name,
    coalesce(num_of_applications,0) as num_of_applications,
    coalesce(num_of_equipments,0) as num_of_equipments,
    agg.product_editor,
    COALESCE(pname,'')::TEXT as name,
    COALESCE(pversion,'')::TEXT as version,
    p_id,
    COALESCE(ar.total_cost,0)::NUMERIC(15,2) as total_cost
from
	aggregations agg
	LEFT JOIN (
		select 
			p.product_name as pname,
			p.product_version as pversion,
			p.swidtag as p_id
			from products p 
		where 
			p.scope =  @scope
	) p on p_id  = ANY(agg.swidtags::TEXT[])
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
    LEFT JOIN (
        SELECT
            a.aggregation_id,
            SUM(a.total_cost)::Numeric(15,2) as total_cost
        FROM
            aggregated_rights a
        WHERE
            a.scope = @scope
        GROUP BY
            a.aggregation_id
    ) ar ON ar.aggregation_id = agg.id
WHERE
     agg.scope = @scope and agg.aggregation_name = @aggregation_name;

-- name: ListSelectedProductsForAggregration :many
SELECT acq.swidtag, acq.product_name, acq.product_editor
FROM acqrights acq 
WHERE acq.scope = @scope 
AND acq.swidtag IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope and agg.id = @id) 
UNION
SELECT prd.swidtag, prd.product_name, prd.product_editor
FROM products prd 
WHERE prd.scope = @scope 
AND prd.swidtag IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope and agg.id = @id);

-- name: GetAggregationByID :one
SELECT *
FROM aggregations
WHERE id = @id
    AND scope = @scope;

-- name: GetTotalCounterfietAmount :one
select coalesce(sum(ocl.delta_cost),0.0)::FLOAT  as counterfiet_amount from overall_computed_licences ocl
where ocl.scope = @scope AND ocl.delta_cost < 0 AND cost_optimization = FALSE;


-- name: GetTotalUnderusageAmount :one
select coalesce(sum(ocl.delta_cost),0.0)::FLOAT  as underusage_amount from overall_computed_licences ocl
where ocl.scope = @scope AND ocl.delta_cost > 0 AND cost_optimization = FALSE;

-- name: GetTotalDeltaCost :one
select coalesce(sum(ocl.delta_cost),0.0)::FLOAT
FROM overall_computed_licences ocl
where cost_optimization = TRUE AND ocl.scope = @scope;

-- name: ListAggregationNameByScope :many
SELECT aggregation_name from aggregations inner join aggregated_rights on aggregations.id = aggregated_rights.aggregation_id where aggregations.scope = $1;

-- name: ListAggregationNameWithScope :many
SELECT aggregation_name, aggregations.scope  from aggregations inner join aggregated_rights on aggregations.id = aggregated_rights.aggregation_id ;

-- name: AddComputedLicensesToAggregation :exec
UPDATE 
  aggregated_rights
SET 
  num_licences_computed = @computedLicenses,
  total_computed_cost = @computedCost
WHERE sku = @sku
AND scope = @scope;

-- name: GetAcqRightMetricsBySwidtag :many
SELECT sku,
       metric
FROM acqrights 
WHERE scope = @scope
    AND swidtag = @swidtag;

-- name: GetAggRightMetricsByAggregationId :many
SELECT sku,
       metric
FROM aggregated_rights
WHERE scope = @scope
    AND aggregation_id = @agg_id;

-- name: GetIndividualProductForAggregationCount :one
SELECT count(*) 
FROM products p 
WHERE p.scope = @scope 
    AND p.swidtag = ANY(@swidtags::TEXT[]);



-- name: InsertOverAllComputedLicences :exec
INSERT INTO overall_computed_licences (
    sku,
    swidtags,
    scope,
    product_names,
    aggregation_name,
    metrics,
    num_computed_licences,
    num_acquired_licences,
    total_cost,
    purchase_cost,
    computed_cost,
    delta_number,
    cost_optimization,
    delta_cost,
    avg_unit_price,
    computed_details,
    metic_not_defined,
    not_deployed,
    editor)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19);

-- name: DeleteOverallComputedLicensesByScope :exec
DELETE FROM overall_computed_licences WHERE scope = @scope;