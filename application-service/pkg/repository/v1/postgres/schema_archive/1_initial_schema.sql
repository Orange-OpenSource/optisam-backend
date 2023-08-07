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

CREATE TABLE IF NOT EXISTS applications (
    application_id VARCHAR NOT NULL ,
    application_name VARCHAR NOT NULL,
    application_version VARCHAR NOT NULL,
    application_owner VARCHAR NOT NULL,
    application_environment VARCHAR NOT NULL,
    application_domain VARCHAR NOT NULL,
    scope VARCHAR NOT NULL,
    obsolescence_risk VARCHAR,
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY(application_id,scope)
);

CREATE INDEX scope_index ON applications (scope);

CREATE TABLE IF NOT EXISTS applications_equipments (
    application_id VARCHAR NOT NULL,
    equipment_Id VARCHAR NOT NULL,
    scope VARCHAR NOT NULL,
    -- SCOPE BASED CHANGE
    PRIMARY KEY (application_id,equipment_Id,scope)
);

CREATE TABLE IF NOT EXISTS applications_instances (
    application_id VARCHAR NOT NULL,
    instance_id VARCHAR NOT NULL,
    instance_environment VARCHAR NOT NULL,
    products TEXT [],
    equipments TEXT [],
    scope VARCHAR NOT NULL,
    -- SCOPE BASED CHANGE
    PRIMARY KEY (instance_id,scope)
);

--  Meta config tables

CREATE TABLE IF NOT EXISTS domain_criticity_meta (
    domain_critic_id SERIAL NOT NULL PRIMARY KEY,
    domain_critic_name VARCHAR NOT NULL
);

INSERT INTO domain_criticity_meta(domain_critic_name) VALUES ('Critical'),('Non Critical'),('Neutral');

CREATE TABLE IF NOT EXISTS maintenance_level_meta (
    maintenance_level_id SERIAL NOT NULL PRIMARY KEY,
    maintenance_level_name VARCHAR NOT NULL
);

INSERT INTO maintenance_level_meta(maintenance_level_name) VALUES ('Level 1'),('Level 2'),('Level 3'),('Level 4');

CREATE TABLE IF NOT EXISTS risk_meta (
    risk_id SERIAL NOT NULL PRIMARY KEY,
    risk_name VARCHAR NOT NULL
);

INSERT INTO risk_meta(risk_name) VALUES ('Low'),('Medium'),('High');

-- Obsolescence tables

CREATE TABLE IF NOT EXISTS domain_criticity (
    critic_id SERIAL NOT NULL PRIMARY KEY,
    scope VARCHAR NOT NULL,
    domain_critic_id INTEGER NOT NULL REFERENCES domain_criticity_meta(domain_critic_id),
    domains TEXT [] NOT NULL,
    created_by VARCHAR NOT NULL,
    created_on TIMESTAMP DEFAULT NOW(),
    UNIQUE (scope, domain_critic_id)
);

CREATE TABLE IF NOT EXISTS maintenance_time_criticity (
    maintenance_critic_id SERIAL NOT NULL PRIMARY KEY,
    scope VARCHAR NOT NULL,
    level_id INTEGER NOT NULL REFERENCES maintenance_level_meta(maintenance_level_id),
    start_month INTEGER not null,
    end_month INTEGER not null,
    created_by VARCHAR NOT NULL,
    created_on TIMESTAMP DEFAULT NOW(),
    UNIQUE (scope, level_id)
);

CREATE TABLE IF NOT EXISTS risk_matrix (
    configuration_id SERIAL NOT NULL PRIMARY KEY,
    scope VARCHAR NOT NULL,
    created_by VARCHAR NOT NULL,
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (scope)
);


CREATE TABLE IF NOT EXISTS risk_matrix_config (
    configuration_id INTEGER NOT NULL REFERENCES risk_matrix(configuration_id) on DELETE CASCADE,
    domain_critic_id INTEGER NOT NULL REFERENCES domain_criticity_meta(domain_critic_id),
    maintenance_level_id INTEGER NOT NULL  REFERENCES maintenance_level_meta(maintenance_level_id),
    risk_id  INTEGER NOT NULL REFERENCES risk_meta(risk_id),
    UNIQUE (configuration_id,domain_critic_id,maintenance_level_id)
);

-- For testing 
-- insert into applications(application_id,application_name,application_version,application_owner,scope)
-- VALUES ('a1','optisam','0.1.1','Orange','France');
-- insert into applications_instances (application_id,instance_id,instance_environment,products,equipments,scope)
-- VALUES ('a1','a1_i1','Production',ARRAY['p1','p2'],ARRAY['e1'],'France');
-- insert into applications_instances (application_id,instance_id,instance_environment,products,equipments,scope)
-- VALUES ('a1','a1_i2','Qualif',ARRAY['p1','p3'],ARRAY['e1','e2'],'France');

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE jobs;
DROP TABLE applications;
DROP TABLE applications_instances;
DROP TABLE products_applications;
DROP TABLE applications_products;