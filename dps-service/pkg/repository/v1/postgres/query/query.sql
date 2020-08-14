-- name: InsertUploadedData :one
INSERT INTO uploaded_data_files (scope,data_type,file_name,uploaded_by)
VALUES($1,$2,$3,$4) returning *;

-- name: InsertUploadedMetaData :one
INSERT INTO uploaded_data_files (file_name,uploaded_by)
VALUES($1,$2) returning *;


-- name: UpdateFileStatus :exec
UPDATE uploaded_data_files SET status = $1 where upload_id = $2 AND file_name = $3;

-- name: GetFileStatus :one
SELECT status FROM uploaded_data_files WHERE upload_id = $1 AND file_name = $2;

-- name: UpdateFileTotalRecord :exec
UPDATE uploaded_data_files SET total_records = $1 where upload_id = $2 AND file_name = $3;

-- name: UpdateFileSuccessRecord :exec
UPDATE uploaded_data_files SET success_records = success_records + $3 where upload_id = $1 AND file_name = $2;

-- name: UpdateFileFailedRecord :exec
UPDATE uploaded_data_files SET failed_records = failed_records + $3 where upload_id = $1 AND file_name = $2;

-- name: ListUploadedDataFiles :many
SELECT count(*) OVER() AS totalRecords,* from 
uploaded_data_files
WHERE 
    scope = ANY(@scope::TEXT[])
    AND data_type = 'DATA'
ORDER BY
  CASE WHEN @upload_id_asc::bool THEN upload_id END asc,
  CASE WHEN @upload_id_desc::bool THEN upload_id END desc,
  CASE WHEN @scope_asc::bool THEN scope END asc,
  CASE WHEN @scope_desc::bool THEN scope END desc,
  CASE WHEN @file_name_asc::bool THEN file_name END asc,
  CASE WHEN @file_name_desc::bool THEN file_name END desc,
  CASE WHEN @status_asc::bool THEN status END asc,
  CASE WHEN @status_desc::bool THEN status END desc,  
  CASE WHEN @uploaded_by_asc::bool THEN uploaded_by END asc,
  CASE WHEN @uploaded_by_desc::bool THEN uploaded_by END desc,
  CASE WHEN @uploaded_on_asc::bool THEN uploaded_on END asc,
  CASE WHEN @uploaded_on_desc::bool THEN uploaded_on END desc
  LIMIT @page_size OFFSET @page_num;

-- name: ListUploadedMetaDataFiles :many
SELECT count(*) OVER() AS totalRecords,* from 
uploaded_data_files
WHERE data_type = 'METADATA'
ORDER BY
  CASE WHEN @upload_id_asc::bool THEN upload_id END asc,
  CASE WHEN @upload_id_desc::bool THEN upload_id END desc,
  CASE WHEN @scope_asc::bool THEN scope END asc,
  CASE WHEN @scope_desc::bool THEN scope END desc,
  CASE WHEN @file_name_asc::bool THEN file_name END asc,
  CASE WHEN @file_name_desc::bool THEN file_name END desc,
  CASE WHEN @status_asc::bool THEN status END asc,
  CASE WHEN @status_desc::bool THEN status END desc,  
  CASE WHEN @uploaded_by_asc::bool THEN uploaded_by END asc,
  CASE WHEN @uploaded_by_desc::bool THEN uploaded_by END desc,
  CASE WHEN @uploaded_on_asc::bool THEN uploaded_on END asc,
  CASE WHEN @uploaded_on_desc::bool THEN uploaded_on END desc
  LIMIT @page_size OFFSET @page_num;