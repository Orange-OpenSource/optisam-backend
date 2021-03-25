-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TYPE audit_status AS ENUM ('DELETED', 'UPDATED');

CREATE TABLE IF NOT EXISTS users_audit (
  id SERIAL,
  username VARCHAR NOT NULL,
  first_name VARCHAR NOT NULL,
  last_name VARCHAR NOT NULL,
  role VARCHAR NOT NULL REFERENCES roles (user_role),
  locale VARCHAR NOT NULL,
  cont_failed_login SMALLINT NOT NULL DEFAULT 0,
  created_on TIMESTAMP NOT NULL,
  last_login  TIMESTAMP,
  operation audit_status,
  updated_by VARCHAR NOT NULL,
  updated_on TIMESTAMP DEFAULT NOW(),
  PRIMARY KEY(id)
);

-- +migrate Down
-- SQL in section 'Up' is executed when this migration is applied
DELETE FROM users_audit;