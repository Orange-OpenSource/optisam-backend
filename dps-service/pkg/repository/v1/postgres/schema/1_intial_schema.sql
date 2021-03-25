-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TYPE job_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED', 'RETRY', 'RUNNING');

CREATE TABLE jobs (
  job_id SERIAL NOT NULL PRIMARY KEY,
  type VARCHAR NOT NULL,
  status job_status NOT NULL DEFAULT 'PENDING',
  data JSONB NOT NULL,
  comments VARCHAR,
  start_time TIMESTAMP,
  end_time TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  retry_count INTEGER DEFAULT 0
);

CREATE TYPE upload_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED', 'INPROGRESS');

CREATE TYPE data_type AS ENUM ('DATA','METADATA','GLOBALDATA');

CREATE TABLE IF NOT EXISTS uploaded_data_files  (
    upload_id SERIAL NOT NULL,
    scope VARCHAR NOT NULL,
    data_type data_type,
    file_name VARCHAR NOT NULL,
    status upload_status NOT NULL DEFAULT 'PENDING',
    uploaded_by VARCHAR NOT NULL,
    uploaded_on TIMESTAMP NOT NULL DEFAULT NOW(),
    total_records INTEGER NOT NULL DEFAULT 0,
    success_records INTEGER NOT NULL DEFAULT 0,
    failed_records INTEGER NOT NULL DEFAULT 0,
    comments varchar DEFAULT '',
    PRIMARY KEY(upload_id,file_name)
);
-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TYPE job_status;
DROP TYPE upload_status;
DROP TABLE jobs;
DROP TABLE uploaded_data_files;