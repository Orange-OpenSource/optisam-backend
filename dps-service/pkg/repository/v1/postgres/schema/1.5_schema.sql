-- +migrate Up
alter table jobs add COLUMN IF NOT EXISTS ppid VARCHAR;
-- SQL in section 'Up' is executed when this migration is applied
-- alter type upload_status add value IF NOT EXISTS 'CANCELLED';
ALTER TABLE uploaded_data_files
    ALTER COLUMN status TYPE VARCHAR(255);
ALTER TABLE deletion_audit
    ALTER COLUMN status TYPE VARCHAR(255);
DROP TYPE IF EXISTS upload_status CASCADE;
CREATE TYPE "public"."upload_status" AS ENUM ('COMPLETED', 'FAILED', 'INPROGRESS', 'PARTIAL', 'PENDING', 'PROCESSED', 'SUCCESS', 'UPLOADED','CANCELLED');
ALTER TABLE uploaded_data_files
    ALTER COLUMN status TYPE upload_status
    USING (status:: upload_status);
ALTER TABLE deletion_audit
    ALTER COLUMN status TYPE upload_status
    USING (status:: upload_status);
ALTER TABLE deletion_audit ALTER status SET DEFAULT 'INPROGRESS';
ALTER TABLE uploaded_data_files ALTER status SET DEFAULT 'PENDING';

-- +migrate Down
alter table jobs drop COLUMN ppid;