INSERT INTO 
groups(name, fully_qualified_name, created_by, scopes,parent_id)
VALUES ('FRANCE', 'ROOT.FRANCE', 'admin@test.com',
ARRAY [ 'Orange', 'Guinea Conakry', 'Group', 'France' ],1),
('INDIA', 'ROOT.INDIA', 'admin@test.com',
ARRAY [ 'Orange', 'Guinea Conakry', 'Group' ],1),
('GUR', 'ROOT.INDIA.GUR', 'admin@test.com',
ARRAY [ 'Orange', 'Guinea Conakry' ],3);

INSERT INTO
group_ownership(user_id,group_id)
VALUES ('user1@test.com',2),
('user2@test.com',3);


SELECT * FROM 
groups
WHERE fully_qualified_name <@ 
(SELECT fully_qualified_name 
FROM groups
INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
WHERE group_ownership.user_id = 'admin@test.com'
);