
// Application DB
CREATE TABLE IF NOT EXISTS applications_equipments (
    application_id VARCHAR NOT NULL,
    equipment_Id VARCHAR NOT NULL,
    scope VARCHAR NOT NULL,
    -- SCOPE BASED CHANGE
    PRIMARY KEY (application_id,equipment_Id,scope));
ALTER TABLE applications ADD application_environment VARCHAR NOT NULL DEFAULT '';

// Account DB
INSERT INTO users(username,first_name,last_name,password,locale,role)
VALUES 
('service@test.com','super','admin','$2a$11$su8WpIWDzAoOhrvsm2U83OXW8JDs36BJNGVhJgnUIOyZW6DolRJSK','en','SuperAdmin'),
('anjali.katariya@orange.com','super','admin','$2a$11$su8WpIWDzAoOhrvsm2U83OXW8JDs36BJNGVhJgnUIOyZW6DolRJSK','en','SuperAdmin');

// Account DB 
INSERT INTO groups(name, fully_qualified_name, created_by)
VALUES 
('ROOT', 'ROOT', 'service@test.com'),
('ROOT', 'ROOT', 'anjali.katariya@orange.com');

// Account DB
INSERT INTO group_ownership(group_id,user_id) 
VALUES
(853,'service@test.com'),
(854,'anjali.katariya@orange.com');
