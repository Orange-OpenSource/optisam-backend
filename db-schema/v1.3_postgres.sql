/ / Product DB
CREATE TABLE
    IF NOT EXISTS shared_licenses(
        sku VARCHAR NOT NULL,
        scope VARCHAR NOT NULL,
        sharing_scope VARCHAR NOT NULL,
        shared_licences INTEGER NOT NULL DEFAULT 0,
        recieved_licences INTEGER NOT NULL DEFAULT 0,
        primary key (sku, scope, sharing_scope)
    );

/*
 - Concurrent User Table
 - Author : Ranveer Singh
 - Date : 27-09-2022
 - Story : 3402 
 */

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

CREATE TYPE product_type AS ENUM ('ONPREMISE', 'SAAS');

Alter table products
ADD
    COLUMN product_type product_type NOT NULL DEFAULT 'ONPREMISE';

CREATE TABLE
    IF NOT EXISTS nominative_user (
        user_id SERIAL NOT NULL PRIMARY KEY,
        scope VARCHAR NOT NULL,
        swidtag VARCHAR,
        aggregations_id INTEGER,
        activation_date TIMESTAMP DEFAULT NULL,
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

/ / Account DB
CREATE TABLE
    IF NOT EXISTS scopes_expenditure (
        id SERIAL PRIMARY KEY,
        scope_code VARCHAR NOT NULL REFERENCES scopes (scope_code),
        expenses FLOAT NOT NULL,
        expenses_year INTEGER NOT NULL,
        created_on TIMESTAMP DEFAULT NOW(),
        created_by VARCHAR REFERENCES users (username),
        updated_on TIMESTAMP,
        updated_by VARCHAR REFERENCES users (username),
        UNIQUE(scope_code, expenses_year)
    );



// Report DB

//OPTISAM-4393

INSERT INTO report_type (report_type_id,report_type_name) VALUES (3,'Expenses by Editor');

//Account DB - scopes_expenditure 
alter table scopes_expenditure 
drop CONSTRAINT scopes_expenditure_scope_code_fkey;

alter table scopes_expenditure
Add CONSTRAINT scopes_expenditure_scope_code_fkey FOREIGN Key(scope_code)
REFERENCES scopes (scope_code) ON DELETE CASCADE;
