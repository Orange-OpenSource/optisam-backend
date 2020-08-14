-- name: GetReport :many
SELECT count(*) OVER() AS totalRecords,r.report_id,rt.report_type_name,r.report_status,r.created_by,r.created_on FROM
report r
JOIN
report_type rt 
ON r.report_type_id = rt.report_type_id
WHERE r.scope = ANY(@scope::TEXT[])
ORDER BY
  CASE WHEN @report_id_asc::bool THEN r.report_id END asc,
  CASE WHEN @report_id_desc::bool THEN r.report_id END desc,
  CASE WHEN @report_type_name_asc::bool THEN rt.report_type_name END asc,
  CASE WHEN @report_type_name_desc::bool THEN rt.report_type_name END desc,
  CASE WHEN @report_status_asc::bool THEN r.report_status END asc,
  CASE WHEN @report_status_desc::bool THEN r.report_status END desc,
  CASE WHEN @created_by_asc::bool THEN r.created_by END asc,
  CASE WHEN @created_by_desc::bool THEN r.created_by END desc,
  CASE WHEN @created_on_asc::bool THEN r.created_on END asc,
  CASE WHEN @created_on_desc::bool THEN r.created_on END desc
  LIMIT @page_size OFFSET @page_num
; 

-- name: DownloadReport :one
SELECT report_data
FROM report r
WHERE r.report_id = @report_id
AND r.scope = ANY(@scope::TEXT[]);

-- name: GetReportTypes :many
SELECT * FROM report_type;

-- name: GetReportType :one
SELECT * FROM report_type
WHERE report_type_id = @report_type_id;

-- name: SubmitReport :one
INSERT INTO
report(scope,report_type_id,report_status,report_metadata,created_by)
VALUES($1,$2,$3,$4,$5) RETURNING report_id;

-- name: InsertReportData :exec
UPDATE report
SET report_data = @report_data_json
WHERE report_id = @report_id;

-- name: UpdateReportStatus :exec
UPDATE report
SET report_status = @report_status
WHERE report_id = @report_id;