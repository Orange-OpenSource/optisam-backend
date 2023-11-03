-- +migrate Up

ALter table nominative_user_file_uploaded_details add COLUMN IF NOT EXISTS
        ppid VARCHAR;
alter table jobs add COLUMN IF NOT EXISTS ppid VARCHAR;
ALTER TABLE acqrights ADD COLUMN IF NOT EXISTS support_numbers text[] DEFAULT array[]::varchar[];
update acqrights set support_numbers = ARRAY[support_number];
ALTER TABLE aggregated_rights ADD COLUMN IF NOT EXISTS support_numbers text[] DEFAULT array[]::varchar[];
update aggregated_rights set support_numbers = ARRAY[support_number];

-- +migrate Down
alter table jobs drop COLUMN ppid;
ALter table nominative_user_file_uploaded_details drop COLUMN
        ppid VARCHAR;
ALTER TABLE acqrights DROP COLUMN IF EXISTS support_numbers;
