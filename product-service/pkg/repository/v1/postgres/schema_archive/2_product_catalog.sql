-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied


-- ENUM 
CREATE TYPE location_type AS ENUM ('NONE','On Premise','SAAS','Both');
CREATE TYPE opensource_type AS ENUM ('NONE', 'COMMERCIAL', 'COMMUNITY', 'BOTH');

-- Tables 

CREATE TABLE IF NOT EXISTS editor_catalog ( 
  id VARCHAR PRIMARY KEY , 
  name VARCHAR NOT NULL, 
  general_information VARCHAR, 
  partner_managers JSONB, 
  audits JSONB, 
  vendors JSONB, 
  created_on TIMESTAMP NOT NULL,  
  updated_on TIMESTAMP NOT NULL, 
  source VARCHAR
  );
CREATE UNIQUE INDEX editor_name_unique_idx on editor_catalog (LOWER(name));  


CREATE TABLE IF NOT EXISTS product_catalog (
  id VARCHAR PRIMARY KEY,
  name VARCHAR NOT NULL,
  editorID VARCHAR NOT NULL REFERENCES editor_catalog(id) ON DELETE CASCADE,
  genearl_information VARCHAR,
  contract_tips VARCHAR,
  support_vendors JSONB,
  metrics JSONB,
  is_opensource BOOLEAN DEFAULT FALSE,
  licences_opensource VARCHAR,
  is_closesource BOOLEAN DEFAULT FALSE,
  licenses_closesource JSONB,
  location location_type NOT NULL,
  created_on TIMESTAMP NOT NULL,
  updated_on TIMESTAMP NOT NULL,
  recommendation VARCHAR,
  useful_links JSONB,
  swid_tag_product VARCHAR,
  source VARCHAR,
  editor_name VARCHAR NOT NULL,
  opensource_type opensource_type NOT NULL
);

CREATE UNIQUE INDEX product_name_unique_idx on product_catalog (LOWER(name),LOWER(editor_name));  

CREATE TABLE IF NOT EXISTS version_catalog (
 id  VARCHAR PRIMARY KEY,
 swid_tag_system VARCHAR NOT NULL,
 p_id VARCHAR NOT NULL REFERENCES product_catalog(id) ON DELETE CASCADE,
 name VARCHAR NOT NULL,
 end_of_life TIMESTAMP,
 end_of_support TIMESTAMP,
 recommendation VARCHAR,
 swid_tag_version VARCHAR,
 source VARCHAR
);

CREATE UNIQUE INDEX veresion_catalog_name_unique_idx on version_catalog (LOWER(name),p_id);

CREATE TABLE IF NOT EXISTS upload_file_logs(
  upload_id SERIAL NOT NULL,
  file_name VARCHAR NOT NULL DEFAULT '',
  uploaded_on TIMESTAMP NOT NULL DEFAULT NOW(),
  PRIMARY KEY(upload_id),
  message VARCHAR
);
-- Alter table
--   editor_catalog
-- ADD
--   COLUMN group_contract BOOLEAN DEFAULT FALSE;
-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE product_catalog;
DROP TABLE version_catalog;
DROP TABLE editor_catalog;
Drop Table upload_file_logs
-- Alter table
--   editor_catalog DELETE COLUMN group_contract;