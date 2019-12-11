
SELECT 
id,
name,
fully_qualified_name,
parent_id,
scopes
FROM groups
WHERE fully_qualified_name <@ 
(SELECT fully_qualified_name 
FROM groups
INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
WHERE group_ownership.user_id = 'user2@test.com'
);