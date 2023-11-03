-- +migrate Up
Grant ALL ON ALL TABLES
IN SCHEMA "public"
TO optisam_app_user;

Grant ALL ON ALL SEQUENCES
IN SCHEMA "public"
TO optisam_app_user;
-- +migrate Down
REVOKE ALL ON ALL SEQUENCES
IN SCHEMA "public" from optisam_app_user;
REVOKE ALL ON ALL TABLES
IN SCHEMA "public" from optisam_app_user;