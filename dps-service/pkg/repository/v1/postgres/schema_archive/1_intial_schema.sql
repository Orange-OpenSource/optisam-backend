-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TYPE job_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED', 'RETRY', 'RUNNING');

create type deletion_type as enum('ACQRIGHTS','INVENTORY_PARK', 'WHOLE_INVENTORY');

CREATE TYPE upload_status AS ENUM ('COMPLETED', 'FAILED',  'INPROGRESS','PARTIAL' ,'PENDING','PROCESSED','SUCCESS','UPLOADED');

CREATE TYPE data_type AS ENUM ('DATA','METADATA','GLOBALDATA');

CREATE TYPE scope_types AS ENUM ('GENERIC','SPECIFIC');



create table deletion_audit (
  id SERIAL NOT NULL,
  scope varchar NOT NULL,
  deletion_type deletion_type NOT NULL,
  status upload_status DEFAULT 'INPROGRESS' NOT NULL,
  reason varchar DEFAULT '' ,
  created_by varchar NOT NULL,
  created_on TIMESTAMP  Default now() NOT NULL,
  updated_on TIMESTAMP,
  PRIMARY KEY (id,scope,deletion_type)
);

CREATE TABLE jobs (
  job_id SERIAL NOT NULL PRIMARY KEY,
  type VARCHAR NOT NULL,
  status job_status NOT NULL DEFAULT 'PENDING',
  data JSONB NOT NULL,
  comments VARCHAR,
  start_time TIMESTAMP,
  end_time TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  retry_count INTEGER DEFAULT 0,
  meta_data JSONB NOT NULL

);

CREATE Index job_status_indx on jobs(status);

-- add any new status in alphabatical order
CREATE TABLE IF NOT EXISTS uploaded_data_files  (
    upload_id SERIAL NOT NULL,
    gid INTEGER NOT NULL DEFAULT 0,
    scope VARCHAR NOT NULL,
    data_type data_type,
    file_name VARCHAR NOT NULL,
    status upload_status NOT NULL DEFAULT 'PENDING',
    uploaded_by VARCHAR NOT NULL,
    uploaded_on TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_on TIMESTAMP   ,
    total_records INTEGER NOT NULL DEFAULT 0,
    success_records INTEGER NOT NULL DEFAULT 0,
    failed_records INTEGER NOT NULL DEFAULT 0,
    comments varchar DEFAULT '',
    scope_type  scope_types DEFAULT 'GENERIC',
    analysis_id VARCHAR DEFAULT '',
    PRIMARY KEY(upload_id,file_name)
);

CREATE TABLE IF NOT EXISTS core_factor_references(
  id INTEGER NOT NULL,
  manufacturer VARCHAR NOT NULL DEFAULT '',
  model VARCHAR NOT NULL DEFAULT '',
  core_factor VARCHAR NOT NULL DEFAULT '',
  PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS core_factor_logs(
  upload_id SERIAL NOT NULL,
  file_name VARCHAR NOT NULL DEFAULT '',
  uploaded_on TIMESTAMP NOT NULL DEFAULT NOW(),
  PRIMARY KEY(upload_id)
);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TYPE job_status;
DROP TYPE upload_status;
DROP TABLE jobs;
DROP TABLE uploaded_data_files;