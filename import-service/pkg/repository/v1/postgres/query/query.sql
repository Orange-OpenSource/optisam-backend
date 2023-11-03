-- name: InsertNominativeUserRequest :one
INSERT INTO
    nominative_user_requests (
        upload_id,
        scope,
        status,
        swidtag,
        product_name,
        product_version,
        aggregation_id,
        editor,
        file_name,
        file_location,
        sheet_name,
        created_by,
		aggregation_name
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13
    ) RETURNING request_id;

-- name: InsertNominativeUserRequestDetails :exec
INSERT INTO
  nominative_user_requests_details (
  request_id,
  headers,
  host,
  remote_addr
 ) VALUES(
    $1,
    $2,
    $3,
    $4
 );

-- name: UpdateNominativeUserRequestPostgresSuccess :one
UPDATE nominative_user_requests
SET
    postgres_success = @postgres_success
WHERE upload_id = @upload_id RETURNING dgraph_completed_batches,total_dgraph_batches,request_id;

-- name: UpdateNominativeUserRequestDgraphBatchSuccess :one

UPDATE nominative_user_requests
SET
    dgraph_completed_batches = (dgraph_completed_batches +1)
WHERE upload_id = @upload_id RETURNING dgraph_completed_batches,total_dgraph_batches,postgres_success,request_id;

-- name: UpdateNominativeUserRequestDgraphSuccess :exec

UPDATE nominative_user_requests
SET
    dgraph_success= @dgraph_success
WHERE upload_id = @upload_id;

-- name: GetDgraphCompletedBatches :one

SELECT
   dgraph_completed_batches
FROM nominative_user_requests
WHERE
    upload_id = @upload_id;

-- name: UpdateNominativeUserRequestAnalysis :one

UPDATE nominative_user_requests
SET
	total_dgraph_batches = @total_dgraph_batches
WHERE upload_id = @upload_id RETURNING request_id;

-- name: UpdateNominativeUserDetailsRequestAnalysis :exec
UPDATE nominative_user_requests_details
SET
    record_succeed = @record_succeed,
	record_failed = @record_failed
WHERE request_id = @request_id;

-- name: UpdateNominativeUserRequestSuccess :exec

UPDATE nominative_user_requests
SET
    status = $1,
	dgraph_success = $2
WHERE request_id = $3;

-- name: ListNominativeUsersUploadedFiles :many

SELECT
   	count(*) OVER() AS totalRecords,
    nominative_user_requests.request_id,
    upload_id,
    scope,
    swidtag,
    aggregation_id,
    editor,
    created_by,
    created_at,
    record_failed,
    case 
      when record_succeed::jsonb = 'null' then 0
      else jsonb_array_length(record_succeed)
    end as record_succeed,
     case
      when record_failed::jsonb = 'null' then 0
      else jsonb_array_length(record_failed)
    end as record_failed,
    file_name,
    sheet_name,
        CASE 
   	     WHEN aggregation_id IS NOT NULL THEN aggregation_name
        ELSE  product_name
  	    END AS pname,
        CASE 
   	    WHEN aggregation_id IS NOT NULL THEN 'aggregation'
        ELSE  'individual'
  	    END AS nametype,
    product_name,
    product_version,
    aggregation_name,
    status
FROM
    nominative_user_requests
INNER JOIN nominative_user_requests_details
ON nominative_user_requests.request_id = nominative_user_requests_details.request_id 
where
    scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @file_upload_id:: bool THEN nominative_user_requests.request_id = @id
            ELSE TRUE
        END
    )
    ORDER BY
    CASE
        WHEN @file_name_asc:: bool THEN file_name
    END asc,
    CASE
        WHEN @file_name_desc:: bool THEN file_name
    END desc,
    CASE
        WHEN @file_status_asc:: bool THEN status
    END asc,
    CASE
        WHEN @file_status_desc:: bool THEN status
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN editor
    END desc,
    CASE
        WHEN @name_asc:: bool THEN  CASE 
                WHEN aggregation_id IS NOT NULL THEN aggregation_name
                ELSE  product_name
            END
    END asc,
    CASE
        WHEN @name_desc:: bool THEN  CASE 
                WHEN aggregation_id IS NOT NULL THEN aggregation_name
                ELSE  product_name
            END
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN product_version
    END desc,
    CASE
        WHEN @created_by_asc:: bool THEN created_by
    END asc,
    CASE
        WHEN @created_by_desc:: bool THEN created_by
    END desc,
    CASE
        WHEN @created_on_asc:: bool THEN created_at
    END asc,
    CASE
        WHEN @created_on_desc:: bool THEN created_at
    END desc,
    CASE
        WHEN @productType_asc:: bool THEN CASE 
   	    WHEN aggregation_id IS NOT NULL THEN 'aggregation'
        ELSE  'individual'
  	    END
    END asc,
    CASE
        WHEN @productType_desc:: bool THEN CASE 
   	    WHEN aggregation_id IS NOT NULL THEN 'aggregation'
        ELSE  'individual'
  	    END
    END desc
    LIMIT @page_size
    OFFSET @page_num;

