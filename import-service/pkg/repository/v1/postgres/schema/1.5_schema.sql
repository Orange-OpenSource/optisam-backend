-- +migrate Up
--create types
-- +migrate StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'request_status') THEN
        CREATE TYPE request_status AS ENUM
        (
            'PENDING', 'SUCCESS', 'PARTIAL'
        );
    END IF;
    --more types here...
END$$;
-- +migrate StatementEnd


CREATE TABLE IF NOT EXISTS nominative_user_requests (
  request_id SERIAL NOT NULL PRIMARY KEY,
  upload_id VARCHAR NOT NULL,
  scope VARCHAR NOT NULL,
  swidtag VARCHAR,
  status VARCHAR NOT NULL DEFAULT 'PENDING',
  product_name VARCHAR,
  product_version VARCHAR,
  aggregation_id VARCHAR,
  aggregation_name VARCHAR,
  editor VARCHAR,
  file_name VARCHAR,
  file_location VARCHAR,
  sheet_name VARCHAR,
  postgres_success boolean DEFAULT FALSE,
  dgraph_success boolean DEFAULT FALSE,
  total_dgraph_batches INTEGER DEFAULT 0,
  dgraph_completed_batches INTEGER DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_by VARCHAR
);

CREATE TABLE IF NOT EXISTS nominative_user_requests_details (
  id SERIAL NOT NULL PRIMARY KEY,
  request_id INT,
  record_succeed JSONB NULL,
  record_failed JSONB NULL,
  headers JSONB,
  host VARCHAR,
  remote_addr VARCHAR,
   CONSTRAINT fk_nominative_user_requests_details
      FOREIGN KEY(request_id) 
	  REFERENCES nominative_user_requests(request_id)
);
-- +migrate Down
DROP TABLE nominative_user_requests_details;
DROP TABLE nominative_user_requests;
DROP TYPE request_status;


