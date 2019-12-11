CREATE TABLE IF NOT EXISTS roles (
  user_role VARCHAR PRIMARY KEY   
);

INSERT INTO roles(user_role)
VALUES
('Admin'),
('SuperAdmin'),
('User');

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;

CREATE EXTENSION pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  username CITEXT PRIMARY KEY,
  first_name VARCHAR,
  last_name VARCHAR,
  role VARCHAR REFERENCES roles (user_role),
  password VARCHAR NOT NULL,
  locale VARCHAR,
  cont_failed_login SMALLINT NOT NULL DEFAULT 0,
  created_on TIMESTAMP DEFAULT NOW() ,
  last_login  TIMESTAMP
);



DELETE FROM users ;

INSERT INTO users(username,first_name,last_name,password,locale,role)
VALUES 
('admin@test.com','super','admin',crypt('admin', gen_salt('md5')),'en','SuperAdmin');

CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    fully_qualified_name ltree,
    scopes TEXT [],
    parent_id INTEGER REFERENCES groups (id),
    created_by CITEXT REFERENCES users (username),
    created_on TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS fully_qualified_name_gist_idx ON groups USING gist(fully_qualified_name);

DELETE FROM groups ;

INSERT INTO groups(name, fully_qualified_name, created_by, scopes)
VALUES ('ROOT', 'ROOT', 'admin@test.com', ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France', 'Ivory Coast' ]);

CREATE TABLE IF NOT EXISTS group_ownership (
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE, 
    user_id CITEXT REFERENCES  users(username) ON DELETE CASCADE,
    created_on TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

DELETE FROM group_ownership;

INSERT INTO group_ownership(group_id,user_id) VALUES(1,'admin@test.com');
