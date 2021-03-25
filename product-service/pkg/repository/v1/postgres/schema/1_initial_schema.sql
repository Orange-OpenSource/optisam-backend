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

CREATE TABLE IF NOT EXISTS products (
    swidtag VARCHAR NOT NULL,
    product_name VARCHAR NOT NULL DEFAULT '',
    product_version VARCHAR NOT NULL DEFAULT '',
    product_edition VARCHAR NOT NULL DEFAULT '',
    product_category VARCHAR NOT NULL DEFAULT '',
    product_editor VARCHAR NOT NULL DEFAULT '',
    scope VARCHAR NOT NULL ,
    option_of VARCHAR NOT NULL DEFAULT '',
    aggregation_id INTEGER NOT NULL DEFAULT 0,
    aggregation_name VARCHAR NOT NULL DEFAULT '',
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR NOT NULL,
    updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by VARCHAR,
    PRIMARY KEY(swidtag, scope)
);

CREATE TABLE IF NOT EXISTS acqrights (
    sku VARCHAR NOT NULL,
    swidtag VARCHAR NOT NULL,
    product_name VARCHAR NOT NULL,
    product_editor VARCHAR NOT NULL,
    entity VARCHAR NOT NULL DEFAULT '',
    scope VARCHAR NOT NULL,
    metric VARCHAR NOT NULL,
    num_licenses_acquired INTEGER NOT NULL DEFAULT 0,
    num_licences_computed INTEGER NOT NULL DEFAULT 0,
    num_licences_maintainance INTEGER NOT NULL DEFAULT 0,
    avg_unit_price NUMERIC(15,2) NOT NULL DEFAULT 0,
    avg_maintenance_unit_price NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_purchase_cost NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_computed_cost NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_maintenance_cost NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_cost NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR NOT NULL,
    updated_on TIMESTAMP,
    updated_by VARCHAR ,
    start_of_maintenance TIMESTAMP DEFAULT NULL,
    end_of_maintenance TIMESTAMP DEFAULT NULL,
    version VARCHAR NOT NULL,
    PRIMARY KEY(sku, scope)
);

CREATE TABLE IF NOT EXISTS aggregations (
    aggregation_id SERIAL NOT NULL,
    aggregation_name VARCHAR NOT NULL,
    aggregation_metric VARCHAR NOT NULL,
    aggregation_scope VARCHAR NOT NULL,
    products TEXT[] NOT NULL,
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR NOT NULL,
    updated_on TIMESTAMP,
    updated_by VARCHAR,
   PRIMARY KEY(aggregation_id),
    UNIQUE (aggregation_name, aggregation_scope)
);

CREATE TABLE IF NOT EXISTS products_equipments (
    swidtag VARCHAR NOT NULL ,
    equipment_id VARCHAR NOT NULL,
    num_of_users INTEGER,
    scope VARCHAR NOT NULL,
    FOREIGN KEY (swidtag, scope) REFERENCES products ON DELETE CASCADE,
    PRIMARY KEY (swidtag,equipment_id,scope)
);

CREATE TABLE IF NOT EXISTS products_applications (
    swidtag VARCHAR NOT NULL ,
    application_id VARCHAR NOT NULL,
    scope VARCHAR NOT NULL,
    FOREIGN KEY (swidtag, scope) REFERENCES products ON DELETE CASCADE,
    PRIMARY KEY (swidtag,application_id,scope)
);

-- For testing
-- insert into products(swidtag,product_name,product_version,product_edition,product_category,
--  product_editor,scope,aggregation_id,aggregation_name,created_by,updated_by)
--  values ('p1','prod1','0.1.1','0.1.1','test','oracle','TST',1,'agg1','admin','admin');
-- insert into products(swidtag,product_name,product_version,product_edition,product_category,
--  product_editor,scope,option_of,created_by,updated_by)
--  values ('p2','prod2','0.1.2','0.1.2','test','oracle','TST','p1','admin','admin');
-- insert into products_applications(swidtag,application_id)
--  VALUES ('p1','a1');
--  insert into products_applications(swidtag,application_id)
--  VALUES ('p2','a1');
-- INSERT INTO products_equipments(swidtag,equipment_id)
--  values ('p1','e1');
-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s1','p1','prod1','oracle','Orange','France','test_ops',1000,900, 
--     0,10,5,10000,9000,0,10000,'admin','admin');

--     insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s2','p2','prod2','oracle','Orange','France','test_ops',1000,950 
--     ,0,10,5,10000,9500,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s3','p3','prod3','oracle','Orange','France','test_ops',1000,980 
--     ,0,10,5,10000,9800,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s4','p4','prod4','oracle','Orange','France','test_ops',1000,990 
--     ,0,10,5,10000,9900,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s5','p1','prod1','oracle','Orange','France','test_ops',1000,999 
--     ,0,10,5,10000,9990,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s6','p6','prod1','oracle','Orange','France','test_ops',1000,1000 
--     ,0,10,5,10000,10000,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s7','p7','prod1','oracle','Orange','France','test_ops',1000,1001 
--     ,0,10,5,10000,10010,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s8','p8','prod1','oracle','Orange','France','test_ops',1000,1000 
--     ,0,10,5,10000,10000,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s9','p9','prod1','oracle','Orange','France','test_ops',1000,10000 
--     ,0,10,5,10000,10000,0,10000,'admin','admin');

-- insert into acqrights(sku,swidtag ,product_name,product_editor,entity,scope ,metric ,num_licenses_acquired,num_licences_computed,
--     num_licences_maintainance ,avg_unit_price ,avg_maintenance_unit_price ,total_purchase_cost ,total_computed_cost,total_maintenance_cost,
--     total_cost, created_by, updated_by ) values ('p1_s10','p10','prod1','oracle','Orange','France','test_ops',1000,2000 
--     ,0,10,5,10000,20000,0,10000,'admin','admin');


-- insert into aggregations(aggregation_id, aggregation_name, aggregation_metric,scope, products, created_by, updated_by ) 
--     values (1,'agg1','test_ops','France',ARRAY['p1','p2'],'admin','admin');


-- insert into aggregations(aggregation_id, aggregation_name, aggregation_metric,scope, products, created_by, updated_by ) 
--     values (2,'agg2','test_nup','France',ARRAY['p1'],'admin','admin');



-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE products_equipments;
DROP TABLE products_applications;
DROP TABLE products;
DROP TABLE acqrights;
DROP TABLE aggregations;
