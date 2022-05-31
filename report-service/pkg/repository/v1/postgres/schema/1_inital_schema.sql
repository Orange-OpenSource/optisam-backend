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
  retry_count INTEGER DEFAULT 0,
  meta_data JSONB NOT NULL
);

CREATE TYPE report_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED', 'RUNNING');

CREATE TABLE IF NOT EXISTS report_type (
report_type_id SERIAL NOT NULL PRIMARY KEY,
report_type_name VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS report (
report_id SERIAL NOT NULL PRIMARY KEY,
report_type_id INTEGER NOT NULL REFERENCES report_type(report_type_id),
scope VARCHAR NOT NULL,
report_metadata JSONB NOT NULL,
report_data JSON,
report_status report_status NOT NULL DEFAULT 'PENDING',
created_by VARCHAR NOT NULL,
created_on TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO report_type (report_type_id,report_type_name) VALUES (1,'Compliance');
INSERT INTO report_type (report_type_id,report_type_name) VALUES (2,'ProductEquipments');

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE jobs;
DROP TABLE report;
DROP TABLE report_type;
