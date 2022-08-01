-- name: ListConfig :many
SELECT id,name,equipment_type,status,created_by,created_on,updated_by,updated_on from config_master 
WHERE (CASE WHEN @is_equip_type::bool THEN equipment_type = @equipment_type ELSE TRUE END) AND
status = @status AND scope = @scope;

-- name: GetMetadatabyConfigID :many
Select id,equipment_type,attribute_name,config_filename from config_metadata where config_id=$1;

-- name: GetConfig :one
SELECT id,name,equipment_type,status,created_by,created_on,updated_by,updated_on,scope from config_master where id = @id AND status = @status AND scope = @scope;

-- name: GetDataByMetadataID :many
SELECT attribute_value, json_data from config_data where metadata_id=$1;

-- name: DeleteConfig :exec
UPDATE config_master SET status = @status where id = @id AND scope = @scope;

-- name: DeleteConfigData :exec
DELETE FROM config_metadata WHERE config_id=$1; 