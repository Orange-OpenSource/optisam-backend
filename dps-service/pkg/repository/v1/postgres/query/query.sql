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
UPDATE uploaded_data_files SET total_records = $1 , failed_records = $2  where upload_id = $3 AND file_name = $4;

-- name: UpdateFileSuccessRecord :exec
UPDATE uploaded_data_files SET success_records = success_records + $3 where upload_id = $1 AND file_name = $2;

-- name: UpdateFileFailedRecord :exec
UPDATE uploaded_data_files SET failed_records = failed_records + $3 where upload_id = $1 AND file_name = $2;

-- name: UpdateFileFailure :exec
UPDATE uploaded_data_files SET status = $1 , comments = $2 where upload_id = $3 AND file_name = $4;

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
WHERE 
  scope = ANY(@scope::TEXT[])
  AND data_type = 'METADATA'
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


-- name: GetFailedRecord :many
SELECT count(*) OVER() AS totalRecords, comments, data -> 'Data' as record from jobs where status = 'FAILED' and data -> 'UploadID' = $1 and type = 'API_WORKER' limit $2 offset $3;

-- name: GetEntityMonthWise :many
select sum(success_records), lower(file_name) as filename, EXTRACT(month from uploaded_on) as month, EXTRACT(year from uploaded_on) as year from  uploaded_data_files where  DATE(uploaded_on)  < make_date($1,$2,1) and  uploaded_on >= make_date($3,$4,1)  and scope = $5  and status = 'COMPLETED'  and  file_name SIMILAR TO $6
group by ( 2,3,4)  order by 3 desc , 4 DESC ;

-- name: GetFailureReasons :many
select count(TYPE) as failed_records,comments from jobs where status = 'FAILED' and type in ('FILE_WORKER', 'API_WORKER') and end_time >= make_date($1,$2,$3) and (data -> 'Data' ->> 'scope'  = $4 or data ->> 'scope' = $4 ) and data -> 'Data' -> 'metadata_type' is NULL group by (2);

-- name: GetDataFileRecords :one
select coalesce(sum(total_records),0)::BIGINT as total_records, coalesce(sum(failed_records),0) ::BIGINT as failed_records from  uploaded_data_files where  date(uploaded_on) >= make_date($1,$2,$3)   and scope = $4  and  file_name SIMILAR TO $5;

-- name: ListUploadedGlobalDataFiles :many
SELECT count(*) OVER() AS totalRecords,* from 
uploaded_data_files
WHERE 
  scope = ANY(@scope::TEXT[])
  AND data_type = 'GLOBALDATA'
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