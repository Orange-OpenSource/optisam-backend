package rest

const (
	locations = `select distinct p.location as name,count(*),SUM(COUNT(p.location)) OVER() AS total_count from product_catalog p
group by p.location`

	licensings = `select distinct p.licensing as name,count(*),SUM(COUNT(p.licensing)) OVER() AS total_count from product_catalog p 
group by p.licensing
`
	recommendationType = `select distinct p.recommendation as name,count(*),SUM(COUNT(p.recommendation)) OVER() AS total_count from product_catalog p 
group by p.recommendation`
	scopes = `	
select
distinct(j.scope) as name,
count(j.product) as count,
SUM(COUNT(j.product)) OVER() AS total_count
from
(
    Select
        scope,
        product_name as product
    from 
        products
        inner join product_catalog on products.product_name = product_catalog.name AND products.product_editor = product_catalog.editor_name
    UNION
    select
        scope,
        product_name as product     
    from
        acqrights 
        inner join product_catalog on acqrights.product_name = product_catalog.name AND acqrights.product_editor = product_catalog.editor_name
) as j
group by j.scope
`

	vendors = `
select
    (foo) :: text as name,
    count(distinct(j.id)),
    sum(count(distinct(j.id))) OVER() AS totalRecords
from
    (
        select
            json_array_elements_text(support_vendors :: json) as foo,id
        from
            product_catalog
        WHERE
            support_vendors :: TEXT <> 'null' 
    ) as j
where
    foo :: text <> ''
group by
    (foo) :: text
order by
    COUNT(*) desc
    `
	countryCode = `select distinct e.country_code as code,count(*),SUM(COUNT(e.country_code)) OVER() AS total_count from editor_catalog e where e.country_code !='' 
group by e.country_code`
	groupContract = `select distinct e.group_contract as group_contract,count(*),SUM(COUNT(e.group_contract)) OVER() AS total_count from editor_catalog e 
group by e.group_contract`
	editorScopes = `
   select
    distinct(j.scope) as name,
    count(j.editor) as count,
    SUM(COUNT(j.editor)) OVER() AS total_count
from
    (
        Select
            scope,
            product_editor as editor
        from 
            products
            inner join editor_catalog on products.product_editor = editor_catalog.name
        UNION 
        select
            scope,
            product_editor as editor     
        from
            acqrights
            inner join editor_catalog on acqrights.product_editor = editor_catalog.name
    ) as j
group by j.scope`

	auditYears = `
select
    (foo ->> 'year')  as name,
    count(distinct(j.id)),
    sum(count(distinct(j.id))) OVER() AS totalRecords
from
    (
        select
            json_array_elements(audits :: json) as foo,id
        from
            editor_catalog
        WHERE
            audits :: TEXT <> 'null'
    ) as j
where
        foo ->> 'year' <> 'null' and cast(foo ->> 'year' as INTEGER) > 1970
group by
    (foo ->> 'year')`

	getEditor = `-- name: GetEditor :one
SELECT editor_catalog.id, editor_catalog.name, editor_catalog.general_information, editor_catalog.partner_managers, editor_catalog.audits, editor_catalog.vendors, editor_catalog.created_on, editor_catalog.updated_on,editor_catalog.country_code,editor_catalog.address, editor_catalog.group_contract,editor_catalog.global_account_manager,editor_catalog.sourcers,COUNT(product_catalog.id), (Select json_agg(t.scope) as a from (
    (Select scope from products where product_editor = editor_catalog.name )
    UNION
    (select scope from acqrights where product_editor = editor_catalog.name)
) t) from editor_catalog
LEFT JOIN product_catalog ON editor_catalog.id = product_catalog.editorID
where editor_catalog.id = $1
GROUP BY editor_catalog.id
`

	// const listEditors = `-- name: GetEditor :many
	// SELECT count(*) OVER() AS totalRecords, editor_catalog.id, editor_catalog.name, editor_catalog.general_information, editor_catalog.partner_managers, editor_catalog.audits, editor_catalog.vendors, editor_catalog.created_on, editor_catalog.updated_on, COUNT(product_catalog.id),(Select json_agg(t.scope) as a from (
	//     (Select scope from products where product_editor = editor_catalog.name )
	//     UNION
	//     (select scope from acqrights where product_editor = editor_catalog.name)
	// ) t) from editor_catalog
	// LEFT JOIN product_catalog ON editor_catalog.id = product_catalog.editorID
	// where
	// (CASE WHEN $5::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($6::TEXT) || '%' ELSE TRUE END)
	// GROUP BY editor_catalog.id
	// ORDER BY
	//   CASE WHEN $1::bool THEN editor_catalog.created_on END asc,
	//   CASE WHEN $2::bool THEN editor_catalog.created_on END desc,
	//   CASE WHEN $7::bool THEN editor_catalog.name END asc,
	//   CASE WHEN $8::bool THEN editor_catalog.name END desc,
	//   CASE WHEN $9::bool THEN COUNT(product_catalog.id) END asc,
	//   CASE WHEN $10::bool THEN COUNT(product_catalog.id) END desc
	// LIMIT $3 OFFSET $4
	// `

	listEditors = `select * ,(Select json_agg(t.scope) as a from (
    (Select scope from products where product_editor = editor.name )
    UNION
       (select scope from acqrights where product_editor = editor.name)
) t) from (
SELECT count(*) OVER() AS totalRecords, editor_catalog.id, editor_catalog.name, editor_catalog.general_information, editor_catalog.partner_managers, editor_catalog.audits, editor_catalog.vendors, editor_catalog.created_on, editor_catalog.updated_on,editor_catalog.country_code,editor_catalog.address,editor_catalog.group_contract,editor_catalog.global_account_manager,editor_catalog.sourcers, COUNT(product_catalog.id) as pcount from editor_catalog
LEFT JOIN product_catalog ON editor_catalog.id = product_catalog.editorID 
where
(CASE WHEN $5::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($6::TEXT) || '%' ELSE TRUE END)
AND
 (CASE WHEN $11::BOOL THEN group_contract = $12 ELSE TRUE END)
AND
 (CASE WHEN $15::BOOL THEN country_code = ANY($16) ELSE TRUE END)
AND
 (CASE WHEN $17::BOOL THEN 
      (CASE 
  WHEN audits::TEXT <> 'null' THEN 
    EXISTS (
      SELECT 1
      FROM json_array_elements(coalesce(editor_catalog.audits::json ,'[]'::json)) arr
      WHERE arr ->> 'year' <> 'null'
      AND arr ->> 'year' = ANY($18)
    )
  ELSE FALSE
END)
    ELSE TRUE END)
  GROUP BY editor_catalog.id 
    ORDER BY 
CASE WHEN $1::bool THEN editor_catalog.created_on END asc,
CASE WHEN $2::bool THEN editor_catalog.created_on END desc,
CASE WHEN $7::bool THEN editor_catalog.name END asc,
CASE WHEN $8::bool THEN editor_catalog.name END desc,
CASE WHEN $9::bool THEN COUNT(product_catalog.id) END asc,
CASE WHEN $10::bool THEN COUNT(product_catalog.id) END desc,
CASE WHEN $13::bool THEN editor_catalog.group_contract END asc,
CASE WHEN $14::bool THEN editor_catalog.group_contract END desc
LIMIT $3 OFFSET $4) as editor
`
	innerScopeEditor = `
select * ,(Select json_agg(t.scope) as a from (
    (Select scope from products where product_editor = editor.name )
    UNION
       (select scope from acqrights where product_editor = editor.name)
) t) from (
SELECT count(*) OVER() AS totalRecords, editor_catalog.id, editor_catalog.name, editor_catalog.general_information, editor_catalog.partner_managers, editor_catalog.audits, editor_catalog.vendors, editor_catalog.created_on, editor_catalog.updated_on,editor_catalog.country_code,editor_catalog.address,editor_catalog.group_contract,editor_catalog.global_account_manager,editor_catalog.sourcers, COUNT(product_catalog.id) as pcount from editor_catalog
LEFT JOIN product_catalog ON editor_catalog.id = product_catalog.editorID 
INNER JOIN (
	SELECT * FROM (
	SELECT product_editor FROM products WHERE scope = Any($19)
	UNION
	SELECT product_editor FROM acqrights WHERE scope = Any($19)
	GROUP BY product_editor
	) as parkData
	) AS pp ON pp.product_editor= editor_catalog.name
where
(CASE WHEN $5::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($6::TEXT) || '%' ELSE TRUE END)
AND
 (CASE WHEN $11::BOOL THEN group_contract = $12 ELSE TRUE END)
AND
 (CASE WHEN $15::BOOL THEN country_code = ANY($16) ELSE TRUE END)
AND
 (CASE WHEN $17::BOOL THEN 
      (CASE 
  WHEN audits::TEXT <> 'null' THEN 
    EXISTS (
      SELECT 1
      FROM json_array_elements(coalesce(editor_catalog.audits::json ,'[]'::json)) arr
      WHERE arr ->> 'year' <> 'null'
      AND arr ->> 'year' = ANY($18)
    )
  ELSE FALSE
END)
    ELSE TRUE END)
GROUP BY editor_catalog.id
order by 
CASE WHEN $1::bool THEN editor_catalog.created_on END asc,
CASE WHEN $2::bool THEN editor_catalog.created_on END desc,
CASE WHEN $7::bool THEN editor_catalog.name END asc,
CASE WHEN $8::bool THEN editor_catalog.name END desc,
CASE WHEN $9::bool THEN COUNT(product_catalog.id) END asc,
CASE WHEN $10::bool THEN COUNT(product_catalog.id) END desc,
CASE WHEN $13::bool THEN editor_catalog.group_contract END asc,
CASE WHEN $14::bool THEN editor_catalog.group_contract END desc
LIMIT $3 OFFSET $4) as editor
`
	listEditorNames = `-- name: GetEditorNames :many
SELECT id, name from editor_catalog
where
(CASE WHEN $1::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($2::TEXT) || '%' ELSE TRUE END)
GROUP BY editor_catalog.id
ORDER BY
  CASE WHEN $3::bool THEN editor_catalog.name END asc,
  CASE WHEN $4::bool THEN editor_catalog.name END desc
LIMIT $5 OFFSET $6
`
	listEditorNamesAll = `-- name: GetEditorNames :many
SELECT id, name from editor_catalog
where
(CASE WHEN $1::BOOL THEN lower(editor_catalog.name) LIKE '%' || lower($2::TEXT) || '%' ELSE TRUE END)
GROUP BY editor_catalog.id
ORDER BY
  CASE WHEN $3::bool THEN editor_catalog.name END asc,
  CASE WHEN $4::bool THEN editor_catalog.name END desc
`
)
