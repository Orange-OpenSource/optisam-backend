-- name: EquipmentProducts :many
SELECT * from products_equipments
WHERE
equipment_id = $1;

-- name: ListEditors :many
SELECT DISTINCT ON (p.product_editor) p.product_editor 
FROM products p 
WHERE p.scope = ANY(@scope::TEXT[]);

-- name: ListProductsView :many
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments , sum(cost)::FLOAT as cost 
FROM products p 
LEFT JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
WHERE
  p.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(p.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(p.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(p.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
  GROUP BY p.swidtag,pa.num_of_applications, pe.num_of_equipments
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN p.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN p.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN p.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN p.product_name END desc,
  CASE WHEN @product_edition_asc::bool THEN p.product_edition END asc,
  CASE WHEN @product_edition_desc::bool THEN p.product_edition END desc,
  CASE WHEN @product_version_asc::bool THEN p.product_version END asc,
  CASE WHEN @product_version_desc::bool THEN p.product_version END desc,
  CASE WHEN @product_editor_asc::bool THEN 4 END asc,
  CASE WHEN @product_editor_desc::bool THEN 4 END desc,
  CASE WHEN @num_of_applications_asc::bool THEN 5 END asc,
  CASE WHEN @num_of_applications_desc::bool THEN 5 END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN 6 END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN 6 END desc,
  CASE WHEN @cost_asc::bool THEN 7 END asc,
  CASE WHEN @cost_desc::bool THEN 7 END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: ListProductsViewRedirectedApplication :many
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments , sum(cost)::FLOAT as cost 
FROM products p 
INNER JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END) GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE (CASE WHEN @is_equipment_id::bool THEN equipment_id = @equipment_id ELSE TRUE END) GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
WHERE
  p.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(p.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(p.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(p.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
  GROUP BY p.swidtag,pa.num_of_applications, pe.num_of_equipments
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN p.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN p.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN p.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN p.product_name END desc,
  CASE WHEN @product_edition_asc::bool THEN p.product_edition END asc,
  CASE WHEN @product_edition_desc::bool THEN p.product_edition END desc,
  CASE WHEN @product_version_asc::bool THEN p.product_version END asc,
  CASE WHEN @product_version_desc::bool THEN p.product_version END desc,
  CASE WHEN @product_editor_asc::bool THEN 4 END asc,
  CASE WHEN @product_editor_desc::bool THEN 4 END desc,
  CASE WHEN @num_of_applications_asc::bool THEN 5 END asc,
  CASE WHEN @num_of_applications_desc::bool THEN 5 END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN 6 END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN 6 END desc,
  CASE WHEN @cost_asc::bool THEN 7 END asc,
  CASE WHEN @cost_desc::bool THEN 7 END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: ListProductsViewRedirectedEquipment :many
SELECT count(*) OVER() AS totalRecords,p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition, COALESCE(pa.num_of_applications,0)::INTEGER as num_of_applications , COALESCE(pe.num_of_equipments,0)::INTEGER as num_of_equipments , sum(cost)::FLOAT as cost 
FROM products p 
LEFT JOIN 
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications WHERE (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END) GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
INNER JOIN 
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments WHERE (CASE WHEN @is_equipment_id::bool THEN equipment_id = @equipment_id ELSE TRUE END) GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
WHERE
  p.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(p.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(p.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(p.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(p.product_editor) = lower(@product_editor) ELSE TRUE END)
  GROUP BY p.swidtag,pa.num_of_applications, pe.num_of_equipments
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN p.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN p.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN p.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN p.product_name END desc,
  CASE WHEN @product_edition_asc::bool THEN p.product_edition END asc,
  CASE WHEN @product_edition_desc::bool THEN p.product_edition END desc,
  CASE WHEN @product_version_asc::bool THEN p.product_version END asc,
  CASE WHEN @product_version_desc::bool THEN p.product_version END desc,
  CASE WHEN @product_editor_asc::bool THEN 4 END asc,
  CASE WHEN @product_editor_desc::bool THEN 4 END desc,
  CASE WHEN @num_of_applications_asc::bool THEN 5 END asc,
  CASE WHEN @num_of_applications_desc::bool THEN 5 END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN 6 END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN 6 END desc,
  CASE WHEN @cost_asc::bool THEN 7 END asc,
  CASE WHEN @cost_desc::bool THEN 7 END desc
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
SELECT count(*) OVER() AS totalRecords,p.aggregation_id,p.aggregation_name,p.product_editor,array_agg(distinct p.swidtag)::TEXT[] as swidtags, COALESCE(SUM(pa.num_of_applications),0)::INTEGER as num_of_applications , COALESCE(SUM(pe.num_of_equipments),0)::INTEGER as num_of_equipments , sum(cost)::INTEGER as total_cost  
FROM products p 
LEFT JOIN
(SELECT swidtag, count(application_id) as num_of_applications FROM products_applications  GROUP BY swidtag) pa
ON p.swidtag = pa.swidtag
LEFT JOIN
(SELECT swidtag, count(equipment_id) as num_of_equipments FROM products_equipments  GROUP BY swidtag) pe
ON p.swidtag = pe.swidtag
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
  CASE WHEN @num_of_applications_asc::bool THEN 5 END asc,
  CASE WHEN @num_of_applications_desc::bool THEN 5 END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN 6 END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN 6 END desc,
  CASE WHEN @cost_asc::bool THEN 7 END asc,
  CASE WHEN @cost_desc::bool THEN 7 END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: ListAggregationProductsView :many
SELECT p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition ,COUNT(distinct pa.application_id)::INTEGER as num_of_applications,COUNT(distinct pe.equipment_id)::INTEGER as num_of_equipments, sum(cost)::FLOAT as cost 
FROM products p 
LEFT JOIN products_applications pa
ON p.swidtag = pa.swidtag
LEFT JOIN products_equipments pe
ON p.swidtag = pe.swidtag
WHERE
  p.aggregation_id = @aggregation_id
  AND p.scope = ANY(@scope::TEXT[])
GROUP BY p.swidtag,p.product_name,p.product_version,p.product_category,p.product_editor,p.product_edition
;


-- name: ProductAggregationDetails :one
SELECT p.aggregation_id,p.aggregation_name,p.product_editor,array_agg(distinct p.swidtag)::TEXT[] as swidtags,array_agg(distinct p.product_edition)::TEXT[] as editions, COUNT(pa.application_id)::INTEGER as num_of_applications,COUNT(pe.equipment_id)::INTEGER as num_of_equipments, SUM(p.cost)::INTEGER as total_cost  
FROM products p 
LEFT JOIN products_applications pa
ON p.swidtag = pa.swidtag
LEFT JOIN products_equipments pe
ON p.swidtag = pe.swidtag
WHERE
  p.aggregation_id = @aggregation_id
  AND p.scope = ANY(@scope::TEXT[])
GROUP BY p.aggregation_id,p.aggregation_name,p.product_editor;

-- name: ProductAggregationChildOptions :many
SELECT p.swidtag,p.product_name,p.product_editor,p.product_edition,p.product_version
FROM products p
WHERE 
  p.aggregation_id = @aggregation_id
  AND p.scope = ANY(@scope::TEXT[]);

-- name: UpsertProduct :exec
INSERT INTO products (swidtag, product_name, product_version, product_edition, product_category, product_editor,scope,option_of,created_on,created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (swidtag)
DO
 UPDATE SET product_name = $2, product_version = $3, product_edition = $4,product_category = $5,product_editor= $6,scope=$7,option_of=$8,updated_on=$11,updated_by=$12;

-- name: UpsertProductPartial :exec
INSERT INTO products (swidtag,scope,created_by)
VALUES ($1,$2,$3)
ON CONFLICT (swidtag)
DO NOTHING;


-- name: DeleteProductApplications :exec
DELETE FROM products_applications
WHERE swidtag = @product_id and application_id = ANY(@application_id::TEXT[]);

-- name: DeleteProductEquipments :exec
DELETE FROM products_equipments
WHERE swidtag = @product_id and equipment_id = ANY(@equipment_id::TEXT[]);

-- name: GetProductsByEditor :many
SELECT swidtag, product_name
FROM products
WHERE product_editor = @product_editor and scope = ANY(@scopes::TEXT[]);

-- name: UpsertProductAggregation :exec
Update products set aggregation_id = @aggregation_id, aggregation_name = @aggregation_name WHERE
swidtag = ANY(@swidtags::TEXT[]);

-- name: GetProductAggregation :many
SELECT swidtag
FROM products
WHERE aggregation_id = $1 and aggregation_name = $2;

-- name: DeleteProductAggregation :exec
Update products set aggregation_id = $1, aggregation_name = $2 WHERE
aggregation_id = $3;
