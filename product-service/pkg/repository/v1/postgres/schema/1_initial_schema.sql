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

CREATE TABLE IF NOT EXISTS products (
    swidtag VARCHAR NOT NULL PRIMARY KEY,
    product_name VARCHAR NOT NULL DEFAULT '',
    product_version VARCHAR NOT NULL DEFAULT '',
    product_edition VARCHAR NOT NULL DEFAULT '',
    product_category VARCHAR NOT NULL DEFAULT '',
    product_editor VARCHAR NOT NULL DEFAULT '',
    scope VARCHAR NOT NULL ,
    option_of VARCHAR NOT NULL DEFAULT '',
    cost INTEGER NOT NULL DEFAULT 0,
    aggregation_id INTEGER NOT NULL DEFAULT 0,
    aggregation_name VARCHAR NOT NULL DEFAULT '',
    created_on TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR NOT NULL,
    updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by VARCHAR
);

CREATE TABLE IF NOT EXISTS products_equipments (
    swidtag VARCHAR NOT NULL REFERENCES products(swidtag),
    equipment_id VARCHAR NOT NULL,
    num_of_users INTEGER,
    PRIMARY KEY (swidtag,equipment_id)
);

CREATE TABLE IF NOT EXISTS products_applications (
    swidtag VARCHAR NOT NULL REFERENCES products(swidtag),
    application_id VARCHAR NOT NULL,
    PRIMARY KEY (swidtag,application_id)
);

-- For testing
insert into products(swidtag,product_name,product_version,product_edition,product_category,
 product_editor,scope,aggregation_id,aggregation_name,created_by,updated_by)
 values ('p1','prod1','0.1.1','0.1.1','test','oracle','France',1,'agg1','admin','admin');
insert into products(swidtag,product_name,product_version,product_edition,product_category,
 product_editor,scope,option_of,created_by,updated_by)
 values ('p2','prod2','0.1.2','0.1.2','test','oracle','France','p1','admin','admin');
insert into products_applications(swidtag,application_id)
 VALUES ('p1','a1');
 insert into products_applications(swidtag,application_id)
 VALUES ('p2','a1');
INSERT INTO products_equipments(swidtag,equipment_id)
 values ('p1','e1');


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE products_equipments;
DROP TABLE products_applications;
DROP TABLE products;
