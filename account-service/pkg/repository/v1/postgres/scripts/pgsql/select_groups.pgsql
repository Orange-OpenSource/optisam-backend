SELECT 
	id,
	name,
	fully_qualified_name,
	parent_id,
	scopes,
	(SELECT COUNT(*) FROM group_ownership WHERE group_ownership.group_id=groups.id) AS total_users,
	(SELECT COUNT(*) FROM groups e where groups.id=e.parent_id) AS total_groups  
	FROM groups 
	WHERE user_id=user1@user.com;