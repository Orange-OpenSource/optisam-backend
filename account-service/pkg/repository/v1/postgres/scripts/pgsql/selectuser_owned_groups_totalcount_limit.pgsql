
SELECT 
id,
name,
fully_qualified_name,
parent_id,
scopes,
count(*) OVER() AS full_count
FROM groups
WHERE fully_qualified_name <@ 
(SELECT fully_qualified_name 
FROM groups
INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
WHERE group_ownership.user_id = 'admin@test.com'
)LIMIT 2;