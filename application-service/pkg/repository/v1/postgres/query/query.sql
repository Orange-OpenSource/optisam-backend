-- name: GetApplicationsView :many
SELECT count(*) OVER() AS totalRecords,a.application_id,a.application_name,a.application_owner,a.application_domain ,a.obsolescence_risk,COUNT(DISTINCT(ai.instance_id))::INTEGER as num_of_instances,COUNT(DISTINCT(ai.product))::INTEGER as num_of_products,COUNT(DISTINCT(ai.equipment))::INTEGER as num_of_equipments
FROM applications a LEFT JOIN 
(select application_id,instance_id, products, UNNEST(coalesce(products,'{null}')) as product,UNNEST(coalesce(equipments,'{null}')) as equipment FROM applications_instances WHERE  scope = ANY(@scope::TEXT[])) ai
ON a.application_id = ai.application_id
WHERE 
  a.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_application_name::bool THEN lower(a.application_name) LIKE '%' || lower(@application_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_name::bool THEN lower(a.application_name) = lower(@application_name) ELSE TRUE END)
  AND (CASE WHEN @lk_application_owner::bool THEN lower(a.application_owner) LIKE '%' || lower(@application_owner::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_owner::bool THEN lower(a.application_owner) = lower(@application_owner) ELSE TRUE END)
  AND (CASE WHEN @is_product_id::bool THEN @product_id::TEXT = ANY(ai.products) ELSE TRUE END)
  AND (CASE WHEN @lk_application_domain::bool THEN lower(a.application_domain) LIKE '%' || lower(@application_domain::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_domain::bool THEN lower(a.application_domain) = lower(@application_domain) ELSE TRUE END)
  AND (CASE WHEN @lk_obsolescence_risk::bool THEN lower(a.obsolescence_risk) LIKE '%' || lower(@obsolescence_risk::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_obsolescence_risk::bool THEN lower(a.obsolescence_risk) = lower(@obsolescence_risk) ELSE TRUE END)
  GROUP BY a.application_id,a.application_name,a.application_owner,a.application_domain,a.obsolescence_risk
  ORDER BY
  CASE WHEN @application_id_asc::bool THEN a.application_id END asc,
  CASE WHEN @application_id_desc::bool THEN a.application_id END desc,
  CASE WHEN @application_name_asc::bool THEN application_name END asc,
  CASE WHEN @application_name_desc::bool THEN application_name END desc,
  CASE WHEN @application_owner_asc::bool THEN application_owner END asc,
  CASE WHEN @application_owner_desc::bool THEN application_owner END desc,
  CASE WHEN @application_domain_desc::bool THEN application_domain END desc,
  CASE WHEN @application_domain_asc::bool THEN application_domain END asc,
  CASE WHEN @obsolescence_risk_desc::bool THEN obsolescence_risk END desc,
  CASE WHEN @obsolescence_risk_asc::bool THEN obsolescence_risk END asc,
  CASE WHEN @num_of_instances_asc::bool THEN count(ai.instance_id) END asc,
  CASE WHEN @num_of_instances_desc::bool THEN count(ai.instance_id) END desc,
  CASE WHEN @num_of_products_asc::bool THEN COUNT(DISTINCT(ai.product)) END asc,
  CASE WHEN @num_of_products_desc::bool THEN COUNT(DISTINCT(ai.product)) END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN COUNT(DISTINCT(ai.equipment)) END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN COUNT(DISTINCT(ai.equipment)) END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: GetInstancesView :many
SELECT count(*) OVER() AS totalRecords,ai.instance_id,ai.instance_environment,CARDINALITY(coalesce(ai.products,'{}'))::INTEGER as num_of_products
FROM applications_instances ai
WHERE 
  ai.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @is_product_id::bool THEN @product_id::TEXT = ANY (ai.products) ELSE TRUE END)
  AND (CASE WHEN @is_application_id::bool THEN ai.application_id = @application_id ELSE TRUE END)
  ORDER BY
  CASE WHEN @instance_id_asc::bool THEN ai.instance_id END asc,
  CASE WHEN @instance_id_desc::bool THEN ai.instance_id END desc,
  CASE WHEN @instance_environment_asc::bool THEN ai.instance_environment END asc,
  CASE WHEN @instance_environment_desc::bool THEN ai.instance_environment END desc,
  CASE WHEN @num_of_products_asc::bool THEN CARDINALITY(ai.products) END asc,
  CASE WHEN @num_of_products_desc::bool THEN CARDINALITY(ai.products) END desc
  LIMIT @page_size OFFSET @page_num
;

-- name: GetInstanceViewEquipments :many
SELECT count(*) OVER()
FROM
  ( SELECT UNNEST(equipments)
    FROM applications_instances
    WHERE scope = @scope
      AND instance_id = @instance_id
      AND (CASE WHEN @is_product_id::bool THEN @product_id::TEXT = ANY (products) ELSE TRUE END)
      AND (CASE WHEN @is_application_id::bool THEN application_id = @application_id ELSE TRUE END)
  )x
WHERE (CASE WHEN @is_product_id::bool THEN UNNEST = ANY(@equipment_ids::TEXT[]) ELSE TRUE END);

-- name: GetApplicationInstances :many
SELECT * from applications_instances
WHERE application_id = @application_id
AND scope = @scope;

-- name: GetApplicationInstance :one
SELECT * from applications_instances
WHERE instance_id = $1;

-- name: UpsertApplication :exec
INSERT INTO applications (application_id, application_name, application_version, application_owner,application_domain, scope, created_on)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (application_id ,scope)
DO
 UPDATE SET application_name = $2, application_version = $3, application_owner = $4,application_domain = $5;

-- name: UpsertApplicationInstance :exec
INSERT INTO applications_instances (application_id, instance_id, instance_environment, products, equipments,scope)
VALUES ($1,$2,$3,$4,$5,$6)
-- SCOPE BASED CHANGE
ON CONFLICT (instance_id,scope)
DO
 UPDATE SET application_id = $1 ,instance_environment = $3, products = $4,equipments = $5;

-- name: GetMaintenanceLevelByMonth :one
SELECT
  maintenance_level_meta.maintenance_level_id,
  maintenance_level_meta.maintenance_level_name
FROM
  maintenance_level_meta
  INNER JOIN maintenance_time_criticity ON maintenance_level_meta.maintenance_level_id = maintenance_time_criticity.level_id
WHERE
  maintenance_time_criticity.start_month <= @calMonth
  AND @calMonth <= maintenance_time_criticity.end_month
  AND maintenance_time_criticity.scope = @scope;

-- name: GetDomainCriticityByDomain :one
SELECT 
  domain_criticity_meta.domain_critic_id,
  domain_criticity_meta.domain_critic_name
FROM
  domain_criticity_meta
  INNER JOIN domain_criticity ON domain_criticity_meta.domain_critic_id = domain_criticity.domain_critic_id
WHERE 
  @applicationDomain :: TEXT = ANY(domain_criticity.domains)
  AND domain_criticity.scope = @scope;

-- name: GetMaintenanceLevelByMonthByName :one
SELECT 
  maintenance_level_id,
  maintenance_level_name
FROM
  maintenance_level_meta
WHERE maintenance_level_name = @levelName;

-- name: GetApplicationDomains :many
SELECT distinct a.application_domain from applications a where a.scope = $1;

-- name: GetDomainCriticityMeta :many
SELECT * from domain_criticity_meta;

-- name: GetMaintenanceCricityMeta :many
SELECT * from maintenance_level_meta;

-- name: GetDomainCriticity :many
SELECT domain_critic_id,domains from domain_criticity where scope = $1;

-- name: GetMaintenanceTimeCriticity :many
SELECT * from maintenance_time_criticity where scope = $1;

-- name: GetRiskMeta :many
SELECT * from risk_meta;

-- name: GetRiskMatrix :many
SELECT * from risk_matrix;

-- name: GetRiskMatrixConfig :many
SELECT rmc.configuration_id,rmc.domain_critic_id,dcm.domain_critic_name,rmc.maintenance_level_id, mlm.maintenance_level_name,rmc.risk_id, rme.risk_name 
FROM risk_matrix_config rmc, risk_matrix rm,domain_criticity_meta dcm,maintenance_level_meta mlm, risk_meta rme
WHERE rmc.configuration_id = rm.configuration_id
AND rmc.maintenance_level_id = mlm.maintenance_level_id
AND rmc.domain_critic_id = dcm.domain_critic_id
AND rmc.risk_id = rme.risk_id
AND rm.scope = $1;

-- name: GetObsolescenceRiskForApplication :one
SELECT 
  risk_meta.risk_name
FROM 
  risk_meta
  INNER JOIN risk_matrix_config ON risk_meta.risk_id = risk_matrix_config.risk_id
  INNER JOIN risk_matrix ON risk_matrix_config.configuration_id = risk_matrix.configuration_id
WHERE risk_matrix_config.domain_critic_id = @domainCriticID
AND risk_matrix_config.maintenance_level_id = @maintenanceLevelID
AND risk_matrix.scope = @scope;

-- name: AddApplicationbsolescenceRisk :exec
UPDATE 
  applications
SET obsolescence_risk = @riskValue
WHERE application_id = @applicationID
AND scope = @scope;
-- name: GetDomainCriticityMetaIDs :many
SELECT domain_critic_id from domain_criticity_meta;

-- name: GetMaintenanceCricityMetaIDs :many
SELECT maintenance_level_id from maintenance_level_meta;

-- name: GetRiskLevelMetaIDs :many
SELECT risk_id from risk_meta;

-- name: InsertDomainCriticity :exec
INSERT INTO domain_criticity(scope,domain_critic_id,domains,created_by)
VALUES ($1,$2,$3,$4)
ON CONFLICT(scope,domain_critic_id)
DO
UPDATE SET domains = $3;

-- name: InsertMaintenanceTimeCriticity :exec
INSERT INTO maintenance_time_criticity(scope,level_id,start_month,end_month,created_by)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT(scope,level_id)
DO
UPDATE SET start_month=$3,end_month=$4;

-- name: InsertRiskMatrix :one
INSERT INTO risk_matrix(scope,created_by) VALUES ($1,$2) 
ON CONFLICT(scope)
DO
UPDATE SET scope=$1
returning configuration_id;

-- name: InsertRiskMatrixConfig :exec
INSERT INTO risk_matrix_config(configuration_id,domain_critic_id,maintenance_level_id,risk_id)
VALUES ($1,$2,$3,$4)
ON CONFLICT(configuration_id,domain_critic_id,maintenance_level_id)
DO 
UPDATE SET risk_id=$4;

-- name: GetApplicationsDetails :many
SELECT 
  application_id,
  application_name,
  application_version,
  application_owner,
  application_domain,
  scope
FROM 
  applications
GROUP BY 
  application_id,
  application_name,
  application_version,
  application_owner,
  application_domain,
  scope;


-- name: DeleteApplicationsByScope :exec
DELETE FROM applications WHERE scope = @scope;

-- name: DeleteInstancesByScope :exec
DELETE FROM applications_instances WHERE scope = @scope;

-- name: DeleteDomainCriticityByScope :exec
DELETE FROM domain_criticity where scope = $1;

-- name: DeleteMaintenanceCirticityByScope :exec
DELETE FROM maintenance_time_criticity where scope = $1;

-- name: DeleteRiskMatricbyScope :exec
Delete FROM risk_matrix where scope = $1;

-- name: GetEquipmentsByApplicationID :one
SELECT 
    ARRAY_AGG(DISTINCT(equipment_ids))::TEXT[] as equipments
from applications_instances, UNNEST(equipments) as equipment_ids
WHERE 
    scope = @scope and 
    application_id = @application_id;

-- name: GetApplicationsByProduct :many
SELECT count(*) OVER() AS totalRecords,a.application_id,a.application_name,a.application_owner,a.application_domain,a.obsolescence_risk,COUNT(DISTINCT(ai.instance_id))::INTEGER as num_of_instances,COUNT(DISTINCT(ai.equipment))::INTEGER as num_of_equipments
FROM applications a INNER JOIN 
(select application_id,instance_id,UNNEST(coalesce(equipments,'{null}')) as equipment FROM applications_instances, UNNEST(products) as product_swidtags WHERE scope = ANY(@scope::TEXT[]) 
  AND product_swidtags = ANY(@productSwidtags::TEXT[])
) ai
ON a.application_id = ai.application_id
WHERE 
  a.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_application_name::bool THEN lower(a.application_name) LIKE '%' || lower(@application_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_name::bool THEN lower(a.application_name) = lower(@application_name) ELSE TRUE END)
  AND (CASE WHEN @lk_application_owner::bool THEN lower(a.application_owner) LIKE '%' || lower(@application_owner::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_owner::bool THEN lower(a.application_owner) = lower(@application_owner) ELSE TRUE END)
  AND (CASE WHEN @lk_application_domain::bool THEN lower(a.application_domain) LIKE '%' || lower(@application_domain::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_application_domain::bool THEN lower(a.application_domain) = lower(@application_domain) ELSE TRUE END)
  AND (CASE WHEN @lk_obsolescence_risk::bool THEN lower(a.obsolescence_risk) LIKE '%' || lower(@obsolescence_risk::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_obsolescence_risk::bool THEN lower(a.obsolescence_risk) = lower(@obsolescence_risk) ELSE TRUE END)
  GROUP BY a.application_id,a.application_name,a.application_owner,a.application_domain,a.obsolescence_risk
  ORDER BY
  CASE WHEN @application_id_asc::bool THEN a.application_id END asc,
  CASE WHEN @application_id_desc::bool THEN a.application_id END desc,
  CASE WHEN @application_name_asc::bool THEN application_name END asc,
  CASE WHEN @application_name_desc::bool THEN application_name END desc,
  CASE WHEN @application_owner_asc::bool THEN application_owner END asc,
  CASE WHEN @application_owner_desc::bool THEN application_owner END desc,
  CASE WHEN @application_domain_desc::bool THEN application_domain END desc,
  CASE WHEN @application_domain_asc::bool THEN application_domain END asc,
  CASE WHEN @obsolescence_risk_desc::bool THEN obsolescence_risk END desc,
  CASE WHEN @obsolescence_risk_asc::bool THEN obsolescence_risk END asc,
  CASE WHEN @num_of_instances_asc::bool THEN count(ai.instance_id) END asc,
  CASE WHEN @num_of_instances_desc::bool THEN count(ai.instance_id) END desc,
  CASE WHEN @num_of_equipments_asc::bool THEN COUNT(DISTINCT(ai.equipment)) END asc,
  CASE WHEN @num_of_equipments_desc::bool THEN COUNT(DISTINCT(ai.equipment)) END desc
  LIMIT @page_size OFFSET @page_num
;