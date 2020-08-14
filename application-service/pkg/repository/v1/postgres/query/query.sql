-- name: GetApplicationsView :many
SELECT count(*) OVER() AS totalRecords,a.application_id,a.application_name,a.application_owner ,COUNT(ai.instance_id)::INTEGER as num_of_instances,COUNT(DISTINCT(ai.product))::INTEGER as num_of_products,COUNT(DISTINCT(ai.equipment))::INTEGER as num_of_equipments, 0::INTEGER as cost 
FROM applications a LEFT JOIN (select application_id,instance_id, products, UNNEST(products) as product,UNNEST(equipments) as equipment FROM applications_instances) ai
ON a.application_id = ai.application_id
WHERE 
  a.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_application_name::bool THEN lower(a.application_name) LIKE '%' || lower(@application_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_name::bool THEN lower(a.application_name) = lower(@application_name) ELSE TRUE END)
  AND (CASE WHEN @lk_application_owner::bool THEN lower(a.application_owner) LIKE '%' || lower(@application_owner::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_owner::bool THEN lower(a.application_owner) = lower(@application_owner) ELSE TRUE END)
  AND (CASE WHEN @is_product_id::bool THEN @product_id::TEXT = ANY(ai.products) ELSE TRUE END)
  GROUP BY a.application_id
  ORDER BY
  CASE WHEN @application_id_asc::bool THEN a.application_id END asc,
  CASE WHEN @application_id_desc::bool THEN a.application_id END desc,
  CASE WHEN @application_name_asc::bool THEN application_name END asc,
  CASE WHEN @application_name_desc::bool THEN application_name END desc,
  CASE WHEN @application_owner_asc::bool THEN application_owner END asc,
  CASE WHEN @application_owner_desc::bool THEN application_owner END desc,
  CASE WHEN @num_of_instances_asc::bool THEN 4 END asc,
  CASE WHEN @num_of_instances_desc::bool THEN 4 END desc,
  CASE WHEN @num_of_products_asc::bool THEN 5 END asc,
  CASE WHEN @num_of_products_desc::bool THEN 5 END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN 6 END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN 6 END desc,
  CASE WHEN @cost_asc::bool THEN 7 END asc,
  CASE WHEN @cost_desc::bool THEN 7 END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: GetInstancesView :many
SELECT count(*) OVER() AS totalRecords,ai.instance_id,ai.instance_environment,CARDINALITY(ai.products)::INTEGER as num_of_products,CARDINALITY(ai.equipments)::INTEGER as num_of_equipments 
FROM applications_instances ai
WHERE 
  ai.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @is_product_id::bool THEN @product_id::TEXT = ANY (ai.products) ELSE TRUE END)
  AND (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END)
  ORDER BY
  CASE WHEN @instance_id_asc::bool THEN ai.instance_id END asc,
  CASE WHEN @instance_id_desc::bool THEN ai.instance_id END desc,
  CASE WHEN @instance_environment_asc::bool THEN instance_environment END asc,
  CASE WHEN @instance_environment_desc::bool THEN instance_environment END desc,
  CASE WHEN @num_of_products_asc::bool THEN 3 END asc,
  CASE WHEN @num_of_products_desc::bool THEN 3 END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN 4 END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN 4 END desc
  LIMIT @page_size OFFSET @page_num
;

-- name: GetApplicationInstance :one
SELECT * from applications_instances
WHERE instance_id = $1;

-- name: UpsertApplication :exec
INSERT INTO applications (application_id, application_name, application_version, application_owner, scope, created_on)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (application_id)
DO
 UPDATE SET application_name = $2, application_version = $3, application_owner = $4;

-- name: UpsertApplicationInstance :exec
INSERT INTO applications_instances (application_id, instance_id, instance_environment, products, equipments,scope)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (instance_id)
DO
 UPDATE SET instance_environment = $3, products = $4,equipments = $5;
