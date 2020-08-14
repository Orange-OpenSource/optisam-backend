-- docker run --name optisam -p 5432:5432 -e POSTGRES_DB=optisam -e  POSTGRES_USER=optisam -e POSTGRES_PASSWORD=optisam postgres
-- docker cp 1_configure.sql optisam:/
-- docker exec -it optisam psql -d optisam -U optisam -w -f 1_configure.sql


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
  updated_on TIMESTAMP NOT NULL DEFAULT now()
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

DELETE FROM config_master;
DELETE FROM config_metadata;
DELETE FROM config_data;

