-- +migrate Up
alter table jobs add COLUMN IF NOT EXISTS ppid VARCHAR;
-- +migrate Down
alter table jobs drop COLUMN ppid;