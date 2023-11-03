-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION do_something()
returns void AS $$
BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'account_status') THEN
CREATE TYPE account_status AS ENUM ('Active', 'Inactive');
alter table users add column if not EXISTS account_status account_status default 'Inactive';

-- for existing users
update users set account_status='Active';

END IF;
END;
$$ language plpgsql;
select do_something();
-- +migrate StatementEnd

-- +migrate Down
-- SQL in section 'Up' is executed when this migration is applied
alter table
    users
drop 
    column account_status;
DROP TYPE account_status ;
