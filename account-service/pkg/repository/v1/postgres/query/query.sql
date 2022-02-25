-- name: InsertUserAudit :exec
INSERT INTO users_audit(
  username,first_name,last_name,role,locale,cont_failed_login,created_on,last_login,operation,updated_by)
  VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);

-- name: DeleteUser :exec
DELETE FROM users
WHERE username = @user_id;

-- name: GetGroupsByScope :many
Select id,scopes from groups where $1 = ANY(scopes::TEXT[]);

-- name: DeleteGroupById :exec
Delete from groups where id = $1;

-- name: UpdateScopesInGroup :exec
Update groups set scopes = $2 where id = $1;

-- name: DeleteScope :exec
Delete from scopes where scope_code = $1;


