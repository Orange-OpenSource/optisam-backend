-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS status (
  id SERIAL PRIMARY KEY,
  text VARCHAR
);

INSERT INTO status(id,text) VALUES(1,'ACTIVE');
INSERT INTO status(id,text) VALUES(2,'INACTIVE');

CREATE TABLE IF NOT EXISTS config_master (
  id SERIAL PRIMARY KEY  NOT NULL,
  name VARCHAR NOT NULL,
  equipment_type VARCHAR NOT NULL,
  status INTEGER NOT NULL REFERENCES status(id),
  created_by VARCHAR NOT NULL,
  created_on TIMESTAMP NOT NULL,
  updated_by VARCHAR NOT NULL,
  updated_on TIMESTAMP NOT NULL DEFAULT now(),
  scope VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS config_metadata (
  id SERIAL PRIMARY KEY NOT NULL , 
  config_id INTEGER NOT NULL REFERENCES config_master(id) ON DELETE CASCADE,
  equipment_type VARCHAR NOT NULL,
  attribute_name VARCHAR NOT NULL,
  config_filename VARCHAR NOT NULL
);


CREATE TABLE IF NOT EXISTS config_data (
  metadata_id INTEGER  NOT NULL REFERENCES config_metadata(id) ON DELETE CASCADE,
  attribute_value VARCHAR NOT NULL,
  json_data jsonb NOT NULL
);


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE config_master;

DELETE FROM status;

DROP TABLE status;

DELETE FROM config_metadata;
DROP TABLE config_metadata;

DELETE FROM config_data;
DROP TABLE config_data;