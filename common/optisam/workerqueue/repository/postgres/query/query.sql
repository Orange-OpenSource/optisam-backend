-- name: GetJob :one
SELECT * FROM jobs
WHERE job_id = $1;

-- name: GetJobs :many
SELECT * FROM jobs ;

-- name: GetJobsForRetry :many
SELECT * FROM Jobs WHERE status  not in ('FAILED' ,'COMPLETED');

-- name: CreateJob :one
INSERT INTO jobs (type,status,data,comments,start_time,end_time,meta_data,ppid) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING job_id;


-- name: UpdateJobStatusRunning :exec
UPDATE jobs SET status = $2,start_time = $3 WHERE job_id = $1;

-- name: UpdateJobStatusCompleted :exec
UPDATE jobs SET status = $2,end_time = $3 WHERE job_id = $1;

-- name: UpdateJobStatusRetry :exec
UPDATE jobs SET status = $2,retry_count = retry_count + 1 WHERE job_id = $1;

-- name: UpdateJobStatusFailed :exec
UPDATE jobs SET status = $2, end_time = $3, comments = $4 , retry_count = $5 where job_id = $1;