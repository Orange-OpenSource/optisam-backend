-- steps to generate code:
-- add schema changes in migatoin files in product service folder
-- cd to optisam-backend
-- run  $ docker run --rm -v /$(pwd):/src -w //src/catalog-service/pkg/repository/v1/postgres kjconroy/sqlc:1.6.0 generate

-- name: InsertProductCatalog :exec
INSERT INTO product_catalog (id,name,editorID, genearl_information,contract_tips,support_vendors,metrics,licences_opensource,location,created_on,updated_on,recommendation,useful_links,swid_tag_product,editor_name,opensource_type,licensing)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17);

-- name: InsertVersionCatalog :exec
INSERT INTO version_catalog (id,p_id,name,end_of_life,end_of_support,recommendation,swid_tag_version,swid_tag_system)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8);

-- name: GetVersionCatalogByPrductID :many
SELECT * from version_catalog 
WHERE p_id = @id;

-- name: GetProductCatalogByPrductID :one
SELECT id, name, editorid, genearl_information, contract_tips, support_vendors, metrics, licences_opensource, location, created_on, updated_on, recommendation, useful_links,swid_tag_product,opensource_type,editor_name,licensing 
 from product_catalog 
WHERE id = $1;

-- name: InsertEditorCatalog :exec
INSERT INTO editor_catalog (id,name, general_information,partner_managers,audits,vendors,created_on,updated_on,source,country_code,address,group_contract,global_account_manager,sourcers)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14);

-- name: GetEditorCatalog :one
SELECT * from editor_catalog WHERE id = @id;

-- name: DeleteProductCatalog :exec
DELETE FROM product_catalog
WHERE id = $1;

-- name: DeleteEditorCatalog :exec
DELETE FROM editor_catalog
WHERE id = $1;

-- name: UpdateProductCatalog :exec
UPDATE product_catalog SET 
name=$1,editorID=$2, genearl_information=$3,contract_tips=$4,support_vendors=$5,metrics=$6,licences_opensource=$7,
location=$8,updated_on=$9,recommendation=$10,useful_links=$11,swid_tag_product=$12,editor_name=$13,opensource_type=$14,licensing=$15
where id =$16;
-- name: UpdateEditorCatalog :exec
UPDATE editor_catalog SET general_information=$1, partner_managers=$2, audits=$3, vendors=$4, updated_on=$5, name=$7, country_code=$8, address=$9, group_contract= $10, global_account_manager= $11 ,sourcers= $12 where id=$6;

-- name: DeleteVersionCatalog :exec
Delete from version_catalog 
WHERE id = @id;

-- name: UpdateVersionCatalog :exec
UPDATE version_catalog SET 
name=$1,end_of_life=$2, end_of_support=$3,recommendation=$4,swid_tag_version=$5,swid_tag_system=$6
where id = $7;

-- name: GetVersionCatalogBySwidTag :one
SELECT * from version_catalog 
WHERE swid_tag_version = @swid_tag_version;

-- name: GetProductCatalogBySwidTag :one
SELECT * from product_catalog 
WHERE swid_tag_product = @swid_tag_product;

-- name: UpsertEditorCatalog :one
INSERT INTO editor_catalog (id,name,created_on,updated_on) values ($1,$2,$3,$4) ON CONFLICT (LOWER(name)) DO Update SET updated_on =$4 returning id,name;

-- name: UpsertProductCatalog :one
INSERT INTO product_catalog (id,name,editorID,editor_name,is_closesource,is_opensource,genearl_information,location,created_on,updated_on,opensource_type) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) on CONFLICT(LOWER(name),LOWER(editor_name)) Do Update SET updated_on =$9 returning id;

-- name: GetEditorCatalogByName :one
SELECT * from editor_catalog WHERE name = @name;

-- name: GetProductCatalogByEditorId :one
SELECT * from product_catalog 
WHERE editorID = @editorID AND name = @name;

-- name: UpsertVersionCatalog :one
INSERT INTO version_catalog (id,p_id,name,end_of_life,end_of_support,swid_tag_system)
VALUES ($1,$2,$3,$4,$5,$6) on CONFLICT(LOWER(name),p_id) Do Update SET end_of_life =$4,end_of_support=$5 returning id;


-- name: GetProductsByEditorID :many
SELECT * from product_catalog 
WHERE editorID = @editor_id;

-- name: GetProductsNamesByEditorID :many
SELECT id,name from product_catalog 
WHERE editorID = @editor_id;

-- name: UpdateProductEditor :exec
UPDATE product_catalog SET 
updated_on=$1,editor_name=$2
where id =$3;

-- name: UpdateVersionForEditor :exec
UPDATE version_catalog SET swid_tag_system=$1
where id =$2;

-- name: GetUploadFileLogs :many
select * from upload_file_logs order by upload_id desc limit 5 ; 

-- name: CreateUploadFileLog :exec
insert into upload_file_logs (file_name,message)values($1,$2); 

-- name: GetEditorCatalogName :one
SELECT id,name from editor_catalog WHERE id = @id;

-- name: UpdateEditorNameForProductCatalog :exec
update product_catalog set editor_name = $1 where editorid = $2;

-- name: UpdateVersionsSysSwidatagsForEditor :exec
update version_catalog set swid_tag_system = case 
when (name = '')
then
REPLACE(CONCAT((
    select
        name
    from
        product_catalog
    where
        id = p_id
),'_',(
    select
        name
    from
        editor_catalog
    where
        editor_catalog.id = $1
)),' ','_')
else
REPLACE(CONCAT((
    select
        name
    from
        product_catalog
    where
        id = p_id
),
'_',
(
    select
        name
    from
        editor_catalog
    where
        editor_catalog.id = $1
), '_',name),' ','_') 
end
where p_id in (select id from product_catalog where product_catalog.editorID = $1);