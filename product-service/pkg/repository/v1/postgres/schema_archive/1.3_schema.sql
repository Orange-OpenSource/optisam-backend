-- +migrate Up

-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE
    IF NOT EXISTS nominative_user (
        user_id SERIAL NOT NULL PRIMARY KEY,
        scope VARCHAR NOT NULL,
        swidtag VARCHAR,
        aggregations_id INTEGER,
        activation_date TIMESTAMP,
        user_email VARCHAR NOT NULL,
        user_name VARCHAR,
        first_name varchar,
        profile VARCHAR,
        product_editor VARCHAR,
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        created_by VARCHAR NOT NULL,
        updated_by VARCHAR,
        FOREIGN KEY (swidtag, scope) REFERENCES products ON DELETE CASCADE,
        FOREIGN KEY (aggregations_id) REFERENCES aggregations(id) ON DELETE CASCADE,
        UNIQUE(
            user_email,
            swidtag,
            scope,
            profile
        ),
        UNIQUE(
            user_email,
            scope,
            profile,
            aggregations_id
        )
    );

CREATE TABLE
    IF NOT EXISTS product_concurrent_user (
        id SERIAL NOT NULL,
        is_aggregations BOOLEAN DEFAULT FALSE,
        aggregation_id INT DEFAULT 0,
        swidtag VARCHAR NULL,
        number_of_users INTEGER,
        profile_user VARCHAR NULL,
        team VARCHAR NULL,
        scope VARCHAR NOT NULL,
        purchase_date DATE NOT NULL,
        created_on TIMESTAMP NOT NULL DEFAULT NOW(),
        created_by VARCHAR NOT NULL,
        updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_by VARCHAR,
        PRIMARY KEY(id),
        UNIQUE (
            aggregation_id,
            scope,
            purchase_date
        ),
        UNIQUE (swidtag, scope, purchase_date),
        FOREIGN KEY (aggregation_id) REFERENCES aggregations(id) ON DELETE CASCADE,
        FOREIGN KEY (swidtag, scope) REFERENCES products ON DELETE CASCADE
    );

CREATE TABLE
    IF NOT EXISTS shared_licenses(
        sku VARCHAR NOT NULL,
        scope VARCHAR NOT NULL,
        sharing_scope VARCHAR NOT NULL,
        shared_licences INTEGER NOT NULL DEFAULT 0,
        recieved_licences INTEGER NOT NULL DEFAULT 0,
        primary key (sku, scope, sharing_scope)
    );

alter table nominative_user
ALTER COLUMN activation_date
SET default null;

CREATE TYPE file_status AS ENUM ('PARTIAL','SUCCESS','FAILED');


CREATE TABLE
    IF NOT EXISTS nominative_user_file_uploaded_details (
        id SERIAL NOT NULL PRIMARY KEY,
        scope VARCHAR NOT NULL,
        swidtag VARCHAR,
        aggregations_id INTEGER,
        product_editor VARCHAR,
        uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
        uploaded_by VARCHAR NOT NULL,
        nominative_users_details JSONB,
        record_succeed INTEGER,
        record_failed INTEGER,
        file_name VARCHAR,
        sheet_name VARCHAR,
        file_status file_status NOT NULL DEFAULT 'FAILED',
        FOREIGN KEY (swidtag, scope) REFERENCES products ON DELETE CASCADE,
        FOREIGN KEY (aggregations_id) REFERENCES aggregations(id) ON DELETE CASCADE
    );

    ALter table nominative_user_file_uploaded_details add COLUMN
        upload_id VARCHAR NOT NULL DEFAULT '';
