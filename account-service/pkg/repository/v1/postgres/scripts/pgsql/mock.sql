INSERT INTO users(username,first_name,last_name,password,locale,role)
VALUES 
('admin1@test.com','super1','admin1',crypt('admin1', gen_salt('md5')),'en','Admin'),
('admin2@test.com','super2','admin2',crypt('admin2', gen_salt('md5')),'en','Admin'),
('admin3@test.com','super3','admin3',crypt('admin3', gen_salt('md5')),'en','Admin'),
('admin4@test.com','super4','admin4',crypt('admin3', gen_salt('md5')),'en','Admin'),
('admin5@test.com','super5','admin5',crypt('admin5', gen_salt('md5')),'en','Admin'),
('admin6@test.com','super6','admin6',crypt('admin6', gen_salt('md5')),'en','Admin'),
('admin7@test.com','super7','admin7',crypt('admin7', gen_salt('md5')),'en','Admin');

INSERT INTO groups(name, fully_qualified_name, created_by, scopes)
VALUES 
('a', 'ROOT.a', 'admin@test.com', ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France', 'Ivory Coast' ]),
('b', 'ROOT.b', 'admin@test.com', ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France', 'Ivory Coast' ]),
('c', 'ROOT.c', 'admin@test.com', ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France', 'Ivory Coast' ]),
('d', 'ROOT.b.d', 'admin@test.com', ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France', 'Ivory Coast' ]),
('e', 'ROOT.b.e', 'admin@test.com', ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France', 'Ivory Coast' ]),
('f', 'ROOT.c.f', 'admin@test.com', ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France', 'Ivory Coast' ]);

INSERT INTO group_ownership(group_id,user_id)
VALUES
(2,'admin1@test.com'),
(3,'admin2@test.com'),
(4,'admin3@test.com'),
(5,'admin4@test.com'),
(6,'admin5@test.com'),
(7,'admin6@test.com'),
(7,'admin7@test.com');


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
	))) AND user_id = new.user_id;
RETURN NEW;
END;
$$
LANGUAGE 'plpgsql';

CREATE TRIGGER insert_group_ownership_correct_group_hierarchy
AFTER INSERT ON group_ownership
FOR EACH ROW
EXECUTE PROCEDURE correct_group_hierarchy();


DELETE FROM group_ownership
   Where group_id IN (
   SELECT group_id
   FROM group_ownership
   INNER JOIN groups ON groups.id  = group_ownership.group_id
   WHERE user_id = 'admin7@test.com'
   AND group_id != 4
   AND fully_qualified_name <@
   (SELECT ARRAY(SELECT fully_qualified_name 
	FROM groups
	INNER JOIN group_ownership ON groups.id  = 4 limit 1
	))
   )  AND user_id = 'admin7@test.com';