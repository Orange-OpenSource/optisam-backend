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
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS applications (
    application_id VARCHAR NOT NULL PRIMARY KEY,
    application_name VARCHAR NOT NULL,
    application_version VARCHAR NOT NULL,
    application_owner VARCHAR NOT NULL,
    scope VARCHAR NOT NULL,
    created_on TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS applications_instances (
    application_id VARCHAR NOT NULL,
    instance_id VARCHAR NOT NULL,
    instance_environment VARCHAR NOT NULL,
    products TEXT [],
    equipments TEXT [],
    scope VARCHAR NOT NULL,
    PRIMARY KEY (instance_id)
);

-- For testing 
insert into applications(application_id,application_name,application_version,application_owner,scope)
VALUES ('a1','optisam','0.1.1','Orange','France');
insert into applications_instances (application_id,instance_id,instance_environment,products,equipments,scope)
VALUES ('a1','a1_i1','Production',ARRAY['p1','p2'],ARRAY['e1'],'France');
insert into applications_instances (application_id,instance_id,instance_environment,products,equipments,scope)
VALUES ('a1','a1_i2','Qualif',ARRAY['p1','p3'],ARRAY['e1','e2'],'France');

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE jobs;
DROP TABLE applications;
DROP TABLE applications_instances;
DROP TABLE products_applications;
DROP TABLE applications_products;