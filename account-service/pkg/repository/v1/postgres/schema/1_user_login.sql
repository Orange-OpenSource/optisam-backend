-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS roles (
  user_role VARCHAR PRIMARY KEY   
);

INSERT INTO roles(user_role)
VALUES
('Admin'),
('SuperAdmin'),
('User');

-- Create extension with root user in db
select control_extension('create','citext');
select control_extension('create','ltree');

CREATE TABLE IF NOT EXISTS users (
  username VARCHAR PRIMARY KEY,
  first_name VARCHAR NOT NULL,
  last_name VARCHAR,
  role VARCHAR REFERENCES roles (user_role),
  password VARCHAR NOT NULL,
  locale VARCHAR NOT NULL,
  cont_failed_login SMALLINT NOT NULL DEFAULT 0,
  created_on TIMESTAMP DEFAULT NOW() ,
  last_login TIMESTAMP,
  first_login BOOLEAN DEFAULT FALSE,
  profile_pic BYTEA
);

INSERT INTO users(username,first_name,last_name,password,locale,role)
VALUES 
('admin@test.com','super','admin','$2a$11$su8WpIWDzAoOhrvsm2U83OXW8JDs36BJNGVhJgnUIOyZW6DolRJSK','en','SuperAdmin');

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    fully_qualified_name ltree,
    scopes TEXT [],
    parent_id INTEGER REFERENCES groups (id) ON DELETE CASCADE,
    created_by VARCHAR REFERENCES users (username),
    created_on TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS fully_qualified_name_gist_idx ON groups USING gist(fully_qualified_name);

INSERT INTO groups(name, fully_qualified_name, created_by)
VALUES ('ROOT', 'ROOT', 'admin@test.com');

CREATE TABLE IF NOT EXISTS group_ownership (
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE, 
    user_id VARCHAR REFERENCES  users(username) ON DELETE CASCADE,
    created_on TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

INSERT INTO group_ownership(group_id,user_id) VALUES(1,'admin@test.com');

create type scope_types as enum('GENERIC','SPECIFIC');

CREATE TABLE IF NOT EXISTS scopes (
  scope_code VARCHAR PRIMARY KEY NOT NULL,
  scope_name VARCHAR NOT NULL,
  scope_type scope_types NOT NULL DEFAULT 'GENERIC',
  created_on TIMESTAMP DEFAULT NOW(),
  created_by VARCHAR REFERENCES users (username)
);

CREATE OR REPLACE FUNCTION correct_group_hierarchy()
  RETURNS trigger AS
$$
BEGIN
   DELETE FROM group_ownership
   Where group_id IN (
   SELECT group_id
   FROM group_ownership
   INNER JOIN groups ON groups.id  = group_ownership.group_id
   WHERE user_id = New.user_id
   AND   group_id != NEW.group_id
   AND fully_qualified_name <@
   (SELECT fully_qualified_name 
	FROM groups
   where id = new.group_id
	)) AND user_id = new.user_id;
RETURN NEW;
END;
$$
LANGUAGE 'plpgsql';

CREATE TRIGGER insert_group_ownership_correct_group_hierarchy
AFTER INSERT ON group_ownership
FOR EACH ROW
EXECUTE PROCEDURE correct_group_hierarchy();
