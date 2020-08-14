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

CREATE TABLE IF NOT EXISTS acqrights (
    sku VARCHAR NOT NULL PRIMARY KEY,
    swidtag VARCHAR NOT NULL,
    product_name VARCHAR NOT NULL DEFAULT '',
    product_editor VARCHAR NOT NULL DEFAULT '',
    entity VARCHAR NOT NULL DEFAULT '',
    scope VARCHAR NOT NULL,
    metric VARCHAR NOT NULL,
    num_licenses_acquired INTEGER NOT NULL DEFAULT 0,
    num_licences_maintainance INTEGER NOT NULL DEFAULT 0,
    avg_unit_price REAL NOT NULL DEFAULT 0,
    avg_maintenance_unit_price REAL NOT NULL DEFAULT 0,
    total_purchase_cost REAL NOT NULL DEFAULT 0,
    total_maintenance_cost REAL NOT NULL DEFAULT 0,
    total_cost REAL NOT NULL DEFAULT 0,
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR NOT NULL,
    updated_on TIMESTAMP,
    updated_by VARCHAR 
);

CREATE TABLE IF NOT EXISTS aggregations (
    aggregation_id SERIAL PRIMARY KEY,
    aggregation_name VARCHAR NOT NULL,
    aggregation_metric VARCHAR NOT NULL,
    aggregation_scope VARCHAR NOT NULL,
    products TEXT[] NOT NULL,
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR NOT NULL,
    updated_on TIMESTAMP,
    updated_by VARCHAR 
);

-- For testing

insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,
    num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_maintenance_cost,
    total_cost, created_by, updated_by ) values ('p1_s1','p1','prod1','oracle','Orange','France','test_ops',10,
    0,10,5,100,0,100,'admin','admin');

insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,
    num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_maintenance_cost,
    total_cost, created_by, updated_by ) values ('p2_s1','p2','prod2','oracle','Orange','France','test_ops',20,
    1,5,5,100,5,100,'admin','admin');

insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,
    num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_maintenance_cost,
    total_cost, created_by, updated_by ) values ('p1_s2','p1','prod1','oracle','Orange','France','test_nup',20,
    1,5,5,100,5,100,'admin','admin');

insert into aggregations(aggregation_id, aggregation_name, aggregation_metric,scope, products, created_by, updated_by ) 
    values (1,'agg1','test_ops','France',ARRAY['p1','p2'],'admin','admin');


insert into aggregations(aggregation_id, aggregation_name, aggregation_metric,scope, products, created_by, updated_by ) 
    values (2,'agg2','test_nup','France',ARRAY['p1'],'admin','admin');


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE job;
DROP TABLE acqrights;
DROP TABLE aggregations;


