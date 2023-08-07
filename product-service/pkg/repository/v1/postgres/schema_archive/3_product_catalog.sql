-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

Alter Table editor_catalog ADD COLUMN country_code VARCHAR default '';
Alter Table editor_catalog ADD COLUMN address VARCHAR default '';

Alter table editor_catalog ADD COLUMN group_contract BOOLEAN DEFAULT FALSE;
CREATE TYPE product_catalog_recommendation AS ENUM ('NONE', 'AUTHORIZED', 'BLACKLISTED', 'RECOMMENDED');
alter table
    product_catalog drop column recommendation;
alter table
    product_catalog
add
    column recommendation product_catalog_recommendation default 'NONE';

CREATE TYPE product_catalog_licensing AS ENUM ('NONE', 'CLOSEDSOURCE', 'OPENSOURCE');
alter table
    product_catalog
add
    column licensing product_catalog_licensing default 'NONE';


Alter Table editor_catalog  ADD COLUMN global_account_manager JSONB default '[]' :: jsonb;
Alter Table editor_catalog  ADD COLUMN sourcers JSONB default '[]' :: jsonb;
    
-- onetime for updating existing data
-- UPDATE
--     product_catalog
-- SET
--     licensing =(
--         CASE
--             WHEN is_opensource = true
--             and is_closesource = false THEN 'OPENSOURCE' :: product_catalog_licensing
--             WHEN is_opensource = false
--             and is_closesource = true THEN 'CLOSEDSOURCE' :: product_catalog_licensing
--             ELSE 'NONE' :: product_catalog_licensing
--         END
--     );

--  strictly onetime
do $$
 DECLARE idx varchar;
result JSON;
BEGIN FOR idx IN SELECT id FROM editor_catalog WHERE audits :: TEXT <> 'null' 
    LOOP 
    result := ( 
        SELECT
            array_to_json(array_agg(t))
        FROM
            (
                SELECT
                    temp.obj -> 'entity' as entity,
                    temp.obj -> 'date' as date,
                    date_part(
                        'year',
                        to_timestamp((temp.obj -> 'date' ->> 'seconds') :: numeric) AT TIME ZONE 'IST'
                    ) as year
                FROM
                    (
                        SELECT
                            audits :: jsonb as array_of_objects
                        from
                            editor_catalog
                        WHERE
                            id = idx
                    ) data,
                    LATERAL jsonb_array_elements(data.array_of_objects) AS temp(obj)
            ) as t
    );
--update editor_catalog set audits = result::jsonb where id = idx;
    RAISE NOTICE 'Result for id %: %',
    idx,
    result;
END LOOP;
END;
$$

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
Alter table
    editor_catalog
DELETE
    COLUMN group_contract;
