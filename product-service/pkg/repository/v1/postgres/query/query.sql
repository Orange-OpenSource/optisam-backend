-- name: EquipmentProducts :many

SELECT * from products_equipments WHERE equipment_id = $1;

-- name: ListEditors :many

SELECT
    DISTINCT (p.product_editor) AS product_editor
FROM products p
WHERE
    p.scope = ANY(@scope:: TEXT [])
    AND LENGTH(p.product_editor) > 0
UNION
SELECT
    DISTINCT (ec.name) AS product_editor
FROM editor_catalog AS ec
WHERE LENGTH(ec.name) > 0;

-- name: ListEditorsScope :many

SELECT
    DISTINCT ON (p.product_editor) p.product_editor
FROM products p
WHERE
    p.scope = ANY(@scope:: TEXT [])
    AND LENGTH(p.product_editor) > 0;

-- name: ListProductsView :many

SELECT
    count(*) OVER() AS totalRecords,
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_category,
    p.product_editor,
    p.product_edition,
    pc.swid_tag_product as product_swid_tag,
    pc.id as product_id,
    v.swid_tag_version as version_swid_tag,
    ec.id as editor_id,
    COALESCE(pa.num_of_applications, 0):: INTEGER as num_of_applications,
    COALESCE(pe.num_of_equipments, 0):: INTEGER as num_of_equipments,
    COALESCE(acq.total_cost, 0):: FLOAT as cost,
    COALESCE(nom_users.nominative_users, 0):: INTEGER as nominative_users,
    COALESCE(
        conc_users.concurrent_users,
        0
    ):: INTEGER as concurrent_users,
    p.product_type:: text,
    COALESCE(pe.equipment_users, 0):: INTEGER as equipment_users
FROM products p
    LEFT JOIN (
        SELECT
            swidtag,
            count(application_id) as num_of_applications
        FROM
            products_applications
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) pa ON p.swidtag = pa.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            count(equipment_id) as num_of_equipments,
            sum(num_of_users) as equipment_users
        FROM
            products_equipments
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) pe ON p.swidtag = pe.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            sum(total_cost) as total_cost
        FROM acqrights
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) acq ON p.swidtag = acq.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            count(user_email) as nominative_users
        FROM nominative_user
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) nom_users ON p.swidtag = nom_users.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            sum(number_of_users) as concurrent_users
        FROM
            product_concurrent_user
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) conc_users ON p.swidtag = conc_users.swidtag
    Left JOIN product_catalog pc ON p.product_name = pc.name
    AND p.product_editor = pc.editor_name
    Left JOIN version_catalog v ON pc.id = v.p_id
    AND v.name = p.product_version
    LEFT JOIN editor_catalog ec ON ec.name = p.product_editor
WHERE
    p.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @lk_swidtag:: bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_swidtag:: bool THEN lower(p.swidtag) = lower(@swidtag)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_editor:: bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_editor:: bool THEN lower(p.product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_type:: bool THEN lower(p.product_type:: text) = lower(@product_type)
            ELSE TRUE
        END
    )
GROUP BY
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_category,
    p.product_editor,
    p.product_edition,
    pa.num_of_applications,
    pe.num_of_equipments,
    pe.equipment_users,
    acq.total_cost,
    pc.swid_tag_product,
    pc.id,
    v.swid_tag_version,
    ec.id,
    nom_users.nominative_users,
    conc_users.concurrent_users,
    p.product_type
ORDER BY
    CASE
        WHEN @swidtag_asc:: bool THEN p.swidtag
    END asc,
    CASE
        WHEN @swidtag_desc:: bool THEN p.swidtag
    END desc,
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @product_edition_asc:: bool THEN p.product_edition
    END asc,
    CASE
        WHEN @product_edition_desc:: bool THEN p.product_edition
    END desc,
    CASE
        WHEN @product_category_asc:: bool THEN p.product_category
    END asc,
    CASE
        WHEN @product_category_desc:: bool THEN p.product_category
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN p.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN p.product_editor
    END desc,
    CASE
        WHEN @num_of_applications_asc:: bool THEN num_of_applications
    END asc,
    CASE
        WHEN @num_of_applications_desc:: bool THEN num_of_applications
    END desc,
    CASE
        WHEN @num_of_equipments_asc:: bool THEN num_of_equipments
    END asc,
    CASE
        WHEN @num_of_equipments_desc:: bool THEN num_of_equipments
    END desc,
    CASE
        WHEN @cost_asc:: bool THEN acq.total_cost
    END asc,
    CASE
        WHEN @cost_desc:: bool THEN acq.total_cost
    END desc,
    CASE
        WHEN @product_type_asc:: bool THEN p.product_type
    END asc,
    CASE
        WHEN @product_type_desc:: bool THEN p.product_type
    END desc,
    CASE
        WHEN @users_asc:: bool  THEN concurrent_users
    END desc,
        CASE
        WHEN @users_desc:: bool  THEN concurrent_users
    END desc,
    CASE 
        WHEN @users_asc:: bool THEN nominative_users
    END desc,
    CASE
        WHEN @users_desc:: bool THEN nominative_users
    END desc,
     CASE
        WHEN @users_asc:: bool THEN equipment_users
    END desc,
    CASE
        WHEN @users_desc:: bool THEN equipment_users
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: ListProductsViewRedirectedApplication :many

SELECT
    count(*) OVER() AS totalRecords,
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_category,
    p.product_editor,
    p.product_edition,
    COALESCE(pa.num_of_applications, 0):: INTEGER as num_of_applications,
    COALESCE(pe.num_of_equipments, 0):: INTEGER as num_of_equipments,
    pe.equipment_ids,
    COALESCE(acq.total_cost, 0):: FLOAT as cost
FROM products p
    INNER JOIN (
        SELECT
            swidtag,
            count(application_id) as num_of_applications
        FROM
            products_applications
        WHERE
            scope = ANY(@scope:: TEXT [])
            AND (
                CASE
                    WHEN @is_application_id:: bool THEN application_id = @application_id
                    ELSE TRUE
                END
            )
        GROUP BY
            swidtag
    ) pa ON p.swidtag = pa.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            count(equipment_id) as num_of_equipments,
            ARRAY_AGG(equipment_id):: TEXT [] as equipment_ids
        FROM
            products_equipments
        WHERE
            scope = ANY(@scope:: TEXT [])
            AND (
                CASE
                    WHEN @is_equipment_id:: bool THEN equipment_id = ANY(@equipment_ids:: TEXT [])
                    ELSE TRUE
                END
            )
        GROUP BY
            swidtag
    ) pe ON p.swidtag = pe.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            sum(total_cost) as total_cost
        FROM acqrights
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) acq ON p.swidtag = acq.swidtag
WHERE
    p.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @lk_swidtag:: bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_swidtag:: bool THEN lower(p.swidtag) = lower(@swidtag)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_editor:: bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_editor:: bool THEN lower(p.product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
GROUP BY
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_category,
    p.product_editor,
    p.product_edition,
    pa.num_of_applications,
    pe.num_of_equipments,
    pe.equipment_ids,
    acq.total_cost
ORDER BY
    CASE
        WHEN @swidtag_asc:: bool THEN p.swidtag
    END asc,
    CASE
        WHEN @swidtag_desc:: bool THEN p.swidtag
    END desc,
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @product_edition_asc:: bool THEN p.product_edition
    END asc,
    CASE
        WHEN @product_edition_desc:: bool THEN p.product_edition
    END desc,
    CASE
        WHEN @product_category_asc:: bool THEN p.product_category
    END asc,
    CASE
        WHEN @product_category_desc:: bool THEN p.product_category
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN p.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN p.product_editor
    END desc,
    CASE
        WHEN @num_of_applications_asc:: bool THEN num_of_applications
    END asc,
    CASE
        WHEN @num_of_applications_desc:: bool THEN num_of_applications
    END desc,
    CASE
        WHEN @num_of_equipments_asc:: bool THEN num_of_equipments
    END asc,
    CASE
        WHEN @num_of_equipments_desc:: bool THEN num_of_equipments
    END desc,
    CASE
        WHEN @cost_asc:: bool THEN acq.total_cost
    END asc,
    CASE
        WHEN @cost_desc:: bool THEN acq.total_cost
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: ListProductsViewRedirectedEquipment :many

SELECT
    count(*) OVER() AS totalRecords,
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_category,
    p.product_editor,
    p.product_edition,
    pe.allocated_metric,
    COALESCE(pa.num_of_applications, 0):: INTEGER as num_of_applications,
    COALESCE(pe.equipment_users, 0):: INTEGER as equipment_users,
    COALESCE(pe.num_of_equipments, 0):: INTEGER as num_of_equipments,
    COALESCE(acq.total_cost, 0):: FLOAT as cost
FROM products p
    LEFT JOIN (
        SELECT
            swidtag,
            count(application_id) as num_of_applications
        FROM
            products_applications
        WHERE
            scope = ANY(@scope:: TEXT [])
            AND (
                CASE
                    WHEN @is_application_id:: bool THEN application_id = @application_id
                    ELSE TRUE
                END
            )
        GROUP BY
            swidtag
    ) pa ON p.swidtag = pa.swidtag
    INNER JOIN (
        SELECT
            swidtag,
            count(equipment_id) as num_of_equipments,
            sum(num_of_users) as equipment_users,
            allocated_metric
        FROM
            products_equipments
        WHERE
            scope = ANY(@scope:: TEXT [])
            AND (
                CASE
                    WHEN @is_equipment_id:: bool THEN equipment_id = @equipment_id
                    ELSE TRUE
                END
            )
        GROUP BY
            swidtag,
            allocated_metric
    ) pe ON p.swidtag = pe.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            sum(total_cost) as total_cost
        FROM acqrights
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) acq ON p.swidtag = acq.swidtag
WHERE
    p.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @lk_swidtag:: bool THEN lower(p.swidtag) LIKE '%' || lower(@swidtag:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_swidtag:: bool THEN lower(p.swidtag) = lower(@swidtag)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_editor:: bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_editor:: bool THEN lower(p.product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
GROUP BY
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_category,
    p.product_editor,
    p.product_edition,
    pa.num_of_applications,
    pe.equipment_users,
    pe.allocated_metric,
    pe.num_of_equipments,
    acq.total_cost
ORDER BY
    CASE
        WHEN @swidtag_asc:: bool THEN p.swidtag
    END asc,
    CASE
        WHEN @swidtag_desc:: bool THEN p.swidtag
    END desc,
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @product_edition_asc:: bool THEN p.product_edition
    END asc,
    CASE
        WHEN @product_edition_desc:: bool THEN p.product_edition
    END desc,
    CASE
        WHEN @product_category_asc:: bool THEN p.product_category
    END asc,
    CASE
        WHEN @product_category_desc:: bool THEN p.product_category
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN p.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN p.product_editor
    END desc,
    CASE
        WHEN @num_of_applications_asc:: bool THEN num_of_applications
    END asc,
    CASE
        WHEN @num_of_applications_desc:: bool THEN num_of_applications
    END desc,
    CASE
        WHEN @num_of_equipments_asc:: bool THEN num_of_equipments
    END asc,
    CASE
        WHEN @num_of_equipments_desc:: bool THEN num_of_equipments
    END desc,
    CASE
        WHEN @cost_asc:: bool THEN acq.total_cost
    END asc,
    CASE
        WHEN @cost_desc:: bool THEN acq.total_cost
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: ListProductsByApplication :many

Select
    count(*) OVER() AS totalRecords,
    p.swidtag,
    p.product_name,
    p.product_version,
    p.product_editor,
    COALESCE(acq.total_cost, 0):: FLOAT as total_cost,
    COALESCE(pe.num_of_equipments, 0):: INTEGER as num_of_equipments
from products p
    LEFT JOIN (
        SELECT
            swidtag,
            sum(total_cost) as total_cost
        FROM acqrights
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) acq ON p.swidtag = acq.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            count(equipment_id) as num_of_equipments
        FROM
            products_equipments
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY
            swidtag
    ) pe ON p.swidtag = pe.swidtag
where
    scope = ANY(@scope:: TEXT [])
    AND p.swidtag = ANY(@swidtag:: TEXT [])
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_editor:: bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_editor:: bool THEN lower(p.product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
ORDER BY
    CASE
        WHEN @swidtag_asc:: bool THEN p.swidtag
    END asc,
    CASE
        WHEN @swidtag_desc:: bool THEN p.swidtag
    END desc,
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @product_edition_asc:: bool THEN p.product_edition
    END asc,
    CASE
        WHEN @product_edition_desc:: bool THEN p.product_edition
    END desc,
    CASE
        WHEN @product_category_asc:: bool THEN p.product_category
    END asc,
    CASE
        WHEN @product_category_desc:: bool THEN p.product_category
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN p.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN p.product_editor
    END desc,
    CASE
        WHEN @total_cost_asc:: bool THEN acq.total_cost
    END asc,
    CASE
        WHEN @total_cost_desc:: bool THEN acq.total_cost
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: GetProductInformation :one

SELECT
    p.swidtag,
    p.product_name,
    p.product_editor,
    p.product_version,
    acq.metrics,
    COALESCE(papp.num_of_applications, 0):: INTEGER as num_of_applications,
    COALESCE(peq.num_of_equipments, 0):: INTEGER as num_of_equipments,
    p.product_type
FROM products p
    LEFT JOIN (
        SELECT
            pa.swidtag,
            count(pa.application_id) as num_of_applications
        FROM
            products_applications pa
        WHERE pa.scope = @scope
        GROUP BY
            pa.swidtag
    ) papp ON p.swidtag = papp.swidtag
    LEFT JOIN (
        SELECT
            pe.swidtag,
            count(pe.equipment_id) as num_of_equipments
        FROM
            products_equipments pe
        WHERE pe.scope = @scope
        GROUP BY
            pe.swidtag
    ) peq ON p.swidtag = peq.swidtag
    LEFT JOIN (
        SELECT
            ac.swidtag,
            ARRAY_AGG(DISTINCT acmetrics):: TEXT [] as metrics
        FROM
            acqrights ac,
            unnest(
                string_to_array(ac.metric, ',')
            ) as acmetrics
        WHERE ac.scope = @scope
        GROUP BY
            ac.swidtag
    ) acq ON p.swidtag = acq.swidtag
WHERE
    p.swidtag = @swidtag
    AND p.scope = @scope
GROUP BY
    p.swidtag,
    p.product_name,
    p.product_editor,
    p.product_version,
    acq.metrics,
    papp.num_of_applications,
    peq.num_of_equipments,
    p.product_type;

-- name: GetProductInformationFromAcqright :one
SELECT ac.swidtag,
       ac.product_name,
       ac.product_editor,
       ac.version,
       p.swid_tag_product as product_swid_tag,
       v.swid_tag_version as version_swid_tag,
       ARRAY_AGG(DISTINCT acmetrics)::TEXT[] as metrics
FROM acqrights ac
Left JOIN product_catalog p ON ac.product_name = p.name AND ac.product_editor = p.editor_name
Left JOIN version_catalog v ON p.id = v.p_id  AND v.name = ac.version,
unnest(string_to_array(ac.metric,',')) as acmetrics
WHERE ac.scope = @scope
    AND ac.swidtag = @swidtag
GROUP BY ac.swidtag,
         ac.product_name,
         ac.product_editor,
         ac.version,
         p.swid_tag_product,
         v.swid_tag_version;

-- name: GetProductOptions :many

SELECT
    p.swidtag,
    p.product_name,
    p.product_edition,
    p.product_editor,
    p.product_version
FROM products p
WHERE
    p.option_of = @swidtag
    AND p.scope = @scope;

-- name: AggregatedRightDetails :one

SELECT
    agg.aggregation_name,
    agg.product_editor,
    agg.products as product_names,
    agg.swidtags as product_swidtags,
    COALESCE( (
            SELECT y.metrics
            FROM (
                    SELECT
                        ARRAY_AGG(DISTINCT armetrics):: TEXT [] as metrics
                    FROM
                        aggregated_rights ar,
                        unnest(
                            string_to_array(ar.metric, ',')
                        ) as armetrics
                    WHERE
                        ar.scope = @scope
                        AND ar.aggregation_id = @id
                ) y
        ),
        '{}'
    ):: TEXT [] as metrics,
    COALESCE( (
            SELECT
                sum(y.num_of_applications)
            FROM (
                    SELECT
                        count(DISTINCT pa.application_id) as num_of_applications
                    FROM
                        products_applications pa
                    WHERE
                        pa.scope = @scope
                        AND pa.swidtag = ANY(agg.swidtags)
                ) y
        ),
        0
    ):: INTEGER as num_of_applications,
    COALESCE( (
            SELECT
                sum(z.num_of_equipments)
            FROM (
                    SELECT
                        count(DISTINCT pe.equipment_id) as num_of_equipments
                    FROM
                        products_equipments pe
                    WHERE
                        pe.scope = @scope
                        AND pe.swidtag = ANY(agg.swidtags)
                ) z
        ),
        0
    ):: INTEGER as num_of_equipments,
    COALESCE( (
            SELECT
                ARRAY_AGG(DISTINCT x.version)
            FROM (
                    SELECT
                        acq.version as version
                    FROM
                        acqrights acq
                    WHERE
                        acq.swidtag = ANY(agg.swidtags)
                        AND acq.scope = @scope
                    UNION
                    SELECT
                        prd.product_version as version
                    FROM
                        products prd
                    WHERE
                        prd.swidtag = ANY(agg.swidtags)
                        AND prd.scope = @scope
                ) x
        ),
        '{}'
    ):: TEXT [] as product_versions
FROM aggregations agg
WHERE scope = @scope AND id = @id
GROUP BY
    agg.aggregation_name,
    agg.product_editor,
    agg.products,
    agg.swidtags;

-- -- name: ProductAggregationChildOptions :many

-- SELECT p.swidtag,p.product_name,p.product_edition,p.product_editor,p.product_version

-- FROM products p

-- WHERE p.option_of in (

-- SELECT p.swidtag

-- FROM products p

-- WHERE

--   p.aggregation_id = @aggregation_id

--   AND p.scope = ANY(@scope::TEXT[]))

-- AND p.scope = ANY(@scope::TEXT[]) ;

-- name: UpsertProduct :exec

INSERT INTO
    products (
        swidtag,
        product_name,
        product_version,
        product_edition,
        product_category,
        product_editor,
        scope,
        option_of,
        created_on,
        created_by,
        product_type
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $13
    ) ON CONFLICT (swidtag, scope)
DO
UPDATE
SET
    product_name = $2,
    product_version = $3,
    product_edition = $4,
    product_category = $5,
    product_editor = $6,
    option_of = $8,
    updated_on = $11,
    updated_by = $12;

-- name: UpsertProductPartial :exec

INSERT INTO
    products (swidtag, scope, created_by)
VALUES ($1, $2, $3) ON CONFLICT (swidtag, scope)
DO NOTHING;

-- name: DeleteProductApplications :exec

DELETE FROM
    products_applications
WHERE
    swidtag = @product_id
    and application_id = ANY(@application_id:: TEXT [])
    and scope = @scope;

-- name: DeleteProductEquipments :exec

DELETE FROM
    products_equipments
WHERE
    swidtag = @product_id
    and equipment_id = ANY(@equipment_id:: TEXT [])
    and scope = @scope;

-- name: GetProductsByEditor :many

SELECT
    p.swidtag swidtag,
    p.product_name product_name,
    p.product_version product_version
FROM products p
    LEFT JOIN (
        SELECT swidtag
        FROM
            products_equipments
        WHERE
            scope = ANY(@scopes:: TEXT [])
        GROUP BY
            swidtag
    ) pe ON p.swidtag = pe.swidtag
WHERE
    p.product_editor = @product_editor
    and p.scope = ANY(@scopes:: TEXT [])
UNION
select
    vc.swid_tag_system swidtag,
    pc.name product_name,
    vc.name product_version
from product_catalog pc
    left join version_catalog vc on pc.id = vc.p_id
    left join editor_catalog ec on pc.editorid = ec.id
WHERE ec.name = @product_editor;

-- name: GetProductsByEditorScope :many

SELECT
    p.swidtag,
    p.product_name,
    p.product_version
FROM products p
    LEFT JOIN (
        SELECT swidtag
        FROM
            products_equipments
        WHERE
            scope = ANY(@scopes:: TEXT [])
        GROUP BY
            swidtag
    ) pe ON p.swidtag = pe.swidtag
WHERE
    p.product_editor = @product_editor
    and p.scope = ANY(@scopes:: TEXT []);

-- name: UpsertAcqRights :exec

INSERT INTO
    acqrights (
        sku,
        swidtag,
        product_name,
        product_editor,
        scope,
        metric,
        num_licenses_acquired,
        avg_unit_price,
        avg_maintenance_unit_price,
        total_purchase_cost,
        total_maintenance_cost,
        total_cost,
        created_by,
        start_of_maintenance,
        end_of_maintenance,
        num_licences_maintainance,
        version,
        comment,
        last_purchased_order,
        support_number,
        maintenance_provider,
        ordering_date,
        corporate_sourcing_contract,
        software_provider,
        file_name,
        file_data,
        repartition
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19,
        $20,
        $21,
        $22,
        $23,
        $24,
        $25,
        $26,
        $27
    ) ON CONFLICT (sku, scope)
DO
UPDATE
SET
    swidtag = $2,
    product_name = $3,
    product_editor = $4,
    scope = $5,
    metric = $6,
    num_licenses_acquired = $7,
    avg_unit_price = $8,
    avg_maintenance_unit_price = $9,
    total_purchase_cost = $10,
    total_maintenance_cost = $11,
    total_cost = $12,
    updated_by = $13,
    start_of_maintenance = $14,
    end_of_maintenance = $15,
    num_licences_maintainance = $16,
    version = $17,
    comment = $18,
    last_purchased_order = $19,
    support_number = $20,
    maintenance_provider = $21,
    ordering_date = $22,
    corporate_sourcing_contract = $23,
    software_provider = $24,
    file_name = $25,
    file_data = $26,
    repartition = $27;

-- name: ListAcqRightsIndividual :many
SELECT count(*) OVER() AS totalRecords,a.sku,a.swidtag,a.product_name,a.product_editor,a.metric,a.num_licenses_acquired,a.num_licences_maintainance,a.avg_unit_price,a.avg_maintenance_unit_price,a.total_purchase_cost,a.total_maintenance_cost,a.total_cost ,a.start_of_maintenance, a.end_of_maintenance , a.version, a.comment, a.ordering_date, a.software_provider, a.corporate_sourcing_contract, a.last_purchased_order, a.support_number, a.maintenance_provider, a.file_name, a.repartition, p.swid_tag_product as product_swid_tag,
       v.swid_tag_version as version_swid_tag,ec.id as editor_id,p.id as product_id
FROM 
acqrights a
Left JOIN product_catalog p ON a.product_name = p.name AND a.product_editor = p.editor_name
Left JOIN version_catalog v ON p.id = v.p_id  AND v.name = a.version
LEFT JOIN editor_catalog as ec ON ec.name = a.product_editor
WHERE 
  a.scope = ANY(@scope::TEXT[])
  AND (CASE WHEN @lk_swidtag::bool THEN lower(a.swidtag) LIKE '%' || lower(@swidtag::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_swidtag::bool THEN lower(a.swidtag) = lower(@swidtag) ELSE TRUE END)
  AND (CASE WHEN @lk_product_name::bool THEN lower(a.product_name) LIKE '%' || lower(@product_name::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_name::bool THEN lower(a.product_name) = lower(@product_name) ELSE TRUE END)
  AND (CASE WHEN @lk_product_editor::bool THEN lower(a.product_editor) LIKE '%' || lower(@product_editor::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_product_editor::bool THEN lower(a.product_editor) = lower(@product_editor) ELSE TRUE END)
  AND (CASE WHEN @lk_sku::bool THEN lower(a.sku) LIKE '%' || lower(@sku::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_sku::bool THEN lower(a.sku) = lower(@sku) ELSE TRUE END)
  AND (CASE WHEN @is_ordering_date::bool THEN a.ordering_date <= @ordering_date ELSE TRUE END)
  AND (CASE WHEN @lk_software_provider::bool THEN lower(a.software_provider) LIKE '%' || lower(@software_provider::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_software_provider::bool THEN lower(a.software_provider) = lower(@software_provider) ELSE TRUE END)
  AND (CASE WHEN @lk_metric::bool THEN lower(a.metric) LIKE '%' || lower(@metric::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_metric::bool THEN lower(a.metric) = lower(@metric) ELSE TRUE END)
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN a.swidtag END asc,
  CASE WHEN @swidtag_desc::bool THEN a.swidtag END desc,
  CASE WHEN @product_name_asc::bool THEN a.product_name END asc,
  CASE WHEN @product_name_desc::bool THEN a.product_name END desc,
  CASE WHEN @product_editor_asc::bool THEN a.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN a.product_editor END desc,
  CASE WHEN @sku_asc::bool THEN a.sku END asc,
  CASE WHEN @sku_desc::bool THEN a.sku END desc,
  CASE WHEN @metric_asc::bool THEN a.metric END asc,
  CASE WHEN @metric_desc::bool THEN a.metric END desc,
  CASE WHEN @num_licenses_acquired_asc::bool THEN a.num_licenses_acquired END asc,
  CASE WHEN @num_licenses_acquired_desc::bool THEN a.num_licenses_acquired END desc,
  CASE WHEN @num_licences_maintainance_asc::bool THEN a.num_licences_maintainance END asc,
  CASE WHEN @num_licences_maintainance_desc::bool THEN a.num_licences_maintainance END desc,
  CASE WHEN @avg_unit_price_asc::bool THEN a.avg_unit_price END asc,
  CASE WHEN @avg_unit_price_desc::bool THEN a.avg_unit_price END desc,  
  CASE WHEN @avg_maintenance_unit_price_asc::bool THEN a.avg_maintenance_unit_price END asc,
  CASE WHEN @avg_maintenance_unit_price_desc::bool THEN a.avg_maintenance_unit_price END desc,
  CASE WHEN @total_purchase_cost_asc::bool THEN a.total_purchase_cost END asc,
  CASE WHEN @total_purchase_cost_desc::bool THEN a.total_purchase_cost END desc,
  CASE WHEN @total_maintenance_cost_asc::bool THEN a.total_maintenance_cost END asc,
  CASE WHEN @total_maintenance_cost_desc::bool THEN a.total_maintenance_cost END desc,
  CASE WHEN @total_cost_asc::bool THEN a.total_cost END asc,
  CASE WHEN @total_cost_desc::bool THEN a.total_cost END desc,
  CASE WHEN @start_of_maintenance_asc::bool THEN a.start_of_maintenance END asc,
  CASE WHEN @start_of_maintenance_desc::bool THEN a.start_of_maintenance END desc,
  CASE WHEN @end_of_maintenance_asc::bool THEN a.end_of_maintenance END asc,
  CASE WHEN @end_of_maintenance_desc::bool THEN a.end_of_maintenance END desc
  LIMIT @page_size OFFSET @page_num;

SELECT
    count(*) OVER() AS totalRecords,
    a.sku,
    a.swidtag,
    a.product_name,
    a.product_editor,
    a.metric,
    a.num_licenses_acquired,
    a.num_licences_maintainance,
    a.avg_unit_price,
    a.avg_maintenance_unit_price,
    a.total_purchase_cost,
    a.total_maintenance_cost,
    a.total_cost,
    a.start_of_maintenance,
    a.end_of_maintenance,
    a.version,
    a.comment,
    a.ordering_date,
    a.software_provider,
    a.corporate_sourcing_contract,
    a.last_purchased_order,
    a.support_number,
    a.maintenance_provider,
    a.file_name,
    a.repartition
FROM acqrights a
WHERE
    a.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @lk_swidtag:: bool THEN lower(a.swidtag) LIKE '%' || lower(@swidtag:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_swidtag:: bool THEN lower(a.swidtag) = lower(@swidtag)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(a.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(a.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_editor:: bool THEN lower(a.product_editor) LIKE '%' || lower(@product_editor:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_editor:: bool THEN lower(a.product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_sku:: bool THEN lower(a.sku) LIKE '%' || lower(@sku:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_sku:: bool THEN lower(a.sku) = lower(@sku)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_ordering_date:: bool THEN a.ordering_date <= @ordering_date
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_software_provider:: bool THEN lower(a.software_provider) LIKE '%' || lower(@software_provider:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_software_provider:: bool THEN lower(a.software_provider) = lower(@software_provider)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_metric:: bool THEN lower(a.metric) LIKE '%' || lower(@metric:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_metric:: bool THEN lower(a.metric) = lower(@metric)
            ELSE TRUE
        END
    )
ORDER BY
    CASE
        WHEN @swidtag_asc:: bool THEN a.swidtag
    END asc,
    CASE
        WHEN @swidtag_desc:: bool THEN a.swidtag
    END desc,
    CASE
        WHEN @product_name_asc:: bool THEN a.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN a.product_name
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN a.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN a.product_editor
    END desc,
    CASE
        WHEN @sku_asc:: bool THEN a.sku
    END asc,
    CASE
        WHEN @sku_desc:: bool THEN a.sku
    END desc,
    CASE
        WHEN @metric_asc:: bool THEN a.metric
    END asc,
    CASE
        WHEN @metric_desc:: bool THEN a.metric
    END desc,
    CASE
        WHEN @num_licenses_acquired_asc:: bool THEN a.num_licenses_acquired
    END asc,
    CASE
        WHEN @num_licenses_acquired_desc:: bool THEN a.num_licenses_acquired
    END desc,
    CASE
        WHEN @num_licences_maintainance_asc:: bool THEN a.num_licences_maintainance
    END asc,
    CASE
        WHEN @num_licences_maintainance_desc:: bool THEN a.num_licences_maintainance
    END desc,
    CASE
        WHEN @avg_unit_price_asc:: bool THEN a.avg_unit_price
    END asc,
    CASE
        WHEN @avg_unit_price_desc:: bool THEN a.avg_unit_price
    END desc,
    CASE
        WHEN @avg_maintenance_unit_price_asc:: bool THEN a.avg_maintenance_unit_price
    END asc,
    CASE
        WHEN @avg_maintenance_unit_price_desc:: bool THEN a.avg_maintenance_unit_price
    END desc,
    CASE
        WHEN @total_purchase_cost_asc:: bool THEN a.total_purchase_cost
    END asc,
    CASE
        WHEN @total_purchase_cost_desc:: bool THEN a.total_purchase_cost
    END desc,
    CASE
        WHEN @total_maintenance_cost_asc:: bool THEN a.total_maintenance_cost
    END asc,
    CASE
        WHEN @total_maintenance_cost_desc:: bool THEN a.total_maintenance_cost
    END desc,
    CASE
        WHEN @total_cost_asc:: bool THEN a.total_cost
    END asc,
    CASE
        WHEN @total_cost_desc:: bool THEN a.total_cost
    END desc,
    CASE
        WHEN @start_of_maintenance_asc:: bool THEN a.start_of_maintenance
    END asc,
    CASE
        WHEN @start_of_maintenance_desc:: bool THEN a.start_of_maintenance
    END desc,
    CASE
        WHEN @end_of_maintenance_asc:: bool THEN a.end_of_maintenance
    END asc,
    CASE
        WHEN @end_of_maintenance_desc:: bool THEN a.end_of_maintenance
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: ListProductAggregation :many

Select
    DISTINCT agg.aggregation_name,
    agg.product_editor,
    ec.id as editor_id,
    agg.swidtags,
    COALESCE( (
            SELECT
                sum(y.num_of_applications)
            FROM (
                    SELECT
                        count(DISTINCT pa.application_id) as num_of_applications
                    FROM
                        products_applications pa
                    WHERE
                        pa.scope = @scope
                        AND pa.swidtag = ANY(agg.swidtags)
                ) y
        ),
        0
    ):: INTEGER as num_of_applications,
    COALESCE( (
            SELECT
                sum(z.num_of_equipments)
            FROM (
                    SELECT
                        count(DISTINCT pe.equipment_id) as num_of_equipments
                    FROM
                        products_equipments pe
                    WHERE
                        pe.scope = @scope
                        AND pe.swidtag = ANY(agg.swidtags)
                ) z
        ),
        0
    ):: INTEGER as num_of_equipments,
    COALESCE( (
            SELECT
                sum(z.num_of_users)
            FROM (
                    SELECT
                        sum(DISTINCT pe.num_of_users) as num_of_users
                    FROM
                        products_equipments pe
                    WHERE
                        pe.scope = @scope
                        AND pe.swidtag = ANY(agg.swidtags)
                ) z
        ),
        0
    ):: INTEGER as num_of_users,
    agg.id,
    COALESCE(ar.total_cost, 0):: NUMERIC(15, 2) as total_cost,
    COALESCE(nu.nom_users, 0):: INTEGER as nominative_users,
    COALESCE(cu.con_users, 0):: INTEGER as concurrent_users
from aggregations agg
    LEFT JOIN (
        select
            p.product_name as name,
            p.product_version as version,
            p.swidtag as p_id
        from products p
        where
            p.scope = @scope
    ) p on p_id = ANY(agg.swidtags:: TEXT [])
    LEFT JOIN (
        SELECT
            a.aggregation_id,
            sum(a.total_cost):: NUMERIC(15, 2) as total_cost
        FROM
            aggregated_rights a
        WHERE a.scope = @scope
        GROUP BY
            a.aggregation_id
    ) ar ON agg.id = ar.aggregation_id
    LEFT JOIN (
        select
            count(nu.user_email) as nom_users,
            aggregations_id as agg_id
        from
            nominative_user nu
        where nu.scope = @scope
        GROUP BY
            agg_id
    ) nu on agg.id = nu.agg_id
    LEFT JOIN (
        select
            sum(number_of_users) as con_users,
            aggregation_id as agg_id
        from
            product_concurrent_user cu
        where cu.scope = @scope
        GROUP BY
            agg_id
    ) cu on agg.id = cu.agg_id
    LEFT JOIN editor_catalog as ec on ec.name = agg.product_editor
WHERE
    agg.scope = @scope
GROUP BY
    agg.aggregation_name,
    agg.product_editor,
    agg.swidtags,
    agg.id,
    ar.total_cost,
    nu.nom_users,
    cu.con_users,
    ec.id
LIMIT
    @page_size OFFSET @page_num;

-- name: ListAcqRightsAggregation :many

SELECT
    count(*) OVER() AS totalRecords,
    a.sku,
    a.aggregation_id,
    a.metric,
    a.ordering_date,
    a.corporate_sourcing_contract,
    a.software_provider,
    a.scope,
    a.num_licenses_acquired,
    a.num_licences_computed,
    a.num_licences_maintenance,
    a.avg_unit_price,
    a.avg_maintenance_unit_price,
    a.total_purchase_cost,
    a.total_computed_cost,
    a.total_maintenance_cost,
    a.total_cost,
    a.start_of_maintenance,
    a.end_of_maintenance,
    a.last_purchased_order,
    a.support_number,
    a.maintenance_provider,
    a.comment,
    a.created_on,
    a.created_by,
    a.updated_on,
    a.updated_by,
    a.file_name,
    a.repartition,
    agg.aggregation_name,
    agg.product_editor,
    agg.products,
    agg.swidtags,
    ec.id as editor_id,
    agg.mapping
from
    aggregated_rights a
    LEFT JOIN (
        SELECT
            ag.id,
            ag.aggregation_name,
            ag.scope,
            ag.product_editor,
            ag.products,
            ag.swidtags,
            (
                Select
                    coalesce(
                        (
                            SELECT
                                array_to_json(array_agg(row_to_json(mapping)))
                            from
                                (
                                    SELECT
                                        products.product_name,
                                        products.product_version
                                    FROM
                                        products
                                    where
                                        swidtag in (
                                            SELECT
                                                UNNEST (ag.swidtags)
                                        )
                                    UNION
                                    select
                                        pc.name product_name,
                                        vc.name product_version
                                    from
                                        product_catalog pc
                                        join version_catalog vc on pc.id = vc.p_id
                                    where
                                        vc.swid_tag_system in (
                                            SELECT
                                                UNNEST (ag.swidtags)
                                        )
                                    UNION
                                    SELECT
                                        acqrights.product_name,
                                        acqrights.version
                                    FROM
                                        acqrights
                                    where
                                        swidtag in (
                                            SELECT
                                                UNNEST (ag.swidtags)
                                        )
                                ) as mapping
                        ),
                        '[]' :: json
                    ) 
            ) as mapping
        FROM
            aggregations ag
        WHERE
            ag.scope = @scope
        GROUP BY
            ag.id
    ) agg ON agg.id = a.aggregation_id
    LEFT JOIN editor_catalog as ec on ec.name = agg.product_editor
WHERE a.scope = @scope
  AND (CASE WHEN @is_agg_name::bool THEN lower(agg.aggregation_name) = lower(@aggregation_name) ELSE TRUE END)
  AND (CASE WHEN @lk_agg_name::bool THEN lower(agg.aggregation_name) LIKE '%' || lower(@aggregation_name) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_editor::bool THEN lower(agg.product_editor) = lower(@product_editor) ELSE TRUE END)
  AND (CASE WHEN @lk_editor::bool THEN lower(agg.product_editor) LIKE '%' || lower(@product_editor) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_sku::bool THEN lower(a.sku) = lower(@sku) ELSE TRUE END)
  AND (CASE WHEN @lk_sku::bool THEN lower(a.sku) LIKE '%' || lower(@sku::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_metric::bool THEN lower(@metric) = any(string_to_array(lower(a.metric) , ',') :: text[]) ELSE TRUE END)
  AND (CASE WHEN @lk_metric::bool THEN lower(a.metric) LIKE '%' || lower(@metric::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_ordering_date::bool THEN a.ordering_date <= @ordering_date ELSE TRUE END)
  AND (CASE WHEN @lk_software_provider::bool THEN lower(a.software_provider) LIKE '%' || lower(@software_provider::TEXT) || '%' ELSE TRUE END)
  AND (CASE WHEN @is_software_provider::bool THEN lower(a.software_provider) = lower(@software_provider) ELSE TRUE END)
  ORDER BY
  CASE WHEN @swidtag_asc::bool THEN ARRAY_Length(agg.swidtags, 1) END asc,
  CASE WHEN @swidtag_desc::bool THEN ARRAY_Length(agg.swidtags, 1) END desc,
  CASE WHEN @agg_name_asc::bool THEN agg.aggregation_name END asc,
  CASE WHEN @agg_name_desc::bool THEN agg.aggregation_name END desc,
  CASE WHEN @product_editor_asc::bool THEN agg.product_editor END asc,
  CASE WHEN @product_editor_desc::bool THEN agg.product_editor END desc,
  CASE WHEN @sku_asc::bool THEN a.sku END asc,
  CASE WHEN @sku_desc::bool THEN a.sku END desc,
  CASE WHEN @metric_asc::bool THEN a.metric END asc,
  CASE WHEN @metric_desc::bool THEN a.metric END desc,
  CASE WHEN @num_licenses_acquired_asc::bool THEN a.num_licenses_acquired END asc,
  CASE WHEN @num_licenses_acquired_desc::bool THEN a.num_licenses_acquired END desc,
  CASE WHEN @num_licences_maintenance_asc::bool THEN a.num_licences_maintenance END asc,
  CASE WHEN @num_licences_maintenance_desc::bool THEN a.num_licences_maintenance END desc,
  CASE WHEN @avg_unit_price_asc::bool THEN a.avg_unit_price END asc,
  CASE WHEN @avg_unit_price_desc::bool THEN a.avg_unit_price END desc,  
  CASE WHEN @avg_maintenance_unit_price_asc::bool THEN a.avg_maintenance_unit_price END asc,
  CASE WHEN @avg_maintenance_unit_price_desc::bool THEN a.avg_maintenance_unit_price END desc,
  CASE WHEN @total_purchase_cost_asc::bool THEN a.total_purchase_cost END asc,
  CASE WHEN @total_purchase_cost_desc::bool THEN a.total_purchase_cost END desc,
  CASE WHEN @total_maintenance_cost_asc::bool THEN a.total_maintenance_cost END asc,
  CASE WHEN @total_maintenance_cost_desc::bool THEN a.total_maintenance_cost END desc,
  CASE WHEN @total_cost_asc::bool THEN a.total_cost END asc,
  CASE WHEN @total_cost_desc::bool THEN a.total_cost END desc,
  CASE WHEN @start_of_maintenance_asc::bool THEN a.start_of_maintenance END asc,
  CASE WHEN @start_of_maintenance_desc::bool THEN a.start_of_maintenance END desc,
  CASE WHEN @end_of_maintenance_asc::bool THEN a.end_of_maintenance END asc,
  CASE WHEN @end_of_maintenance_desc::bool THEN a.end_of_maintenance END desc,
  CASE WHEN @license_under_maintenance_asc::bool THEN age(end_of_maintenance, start_of_maintenance) END desc,
  CASE WHEN @license_under_maintenance_desc::bool  THEN age(end_of_maintenance, start_of_maintenance) END asc
LIMIT @page_size OFFSET @page_num;


-- name: ListProductsAggregationIndividual :many

SELECT
    *,
    COALESCE(pa.num_of_applications, 0):: INTEGER as num_of_applications,
    COALESCE(pe.num_of_equipments, 0):: INTEGER as num_of_equipments
FROM products p
    LEFT JOIN (
        SELECT
            swidtag,
            count(application_id) as num_of_applications
        FROM
            products_applications
        WHERE scope = p.scope
        GROUP BY
            swidtag
    ) pa ON p.swidtag = pa.swidtag
    LEFT JOIN (
        SELECT
            swidtag,
            count(equipment_id) as num_of_equipments
        FROM
            products_equipments
        WHERE scope = p.scope
        GROUP BY
            swidtag
    ) pe ON p.swidtag = pe.swidtag
    LEFT JOIN (
        SELECT *
        FROM aggregations
        WHERE
            scope = p.scope
    ) ar ON p.swidtag = ar.swidtag
WHERE
    p.aggregation_name = @aggregation_name
    and p.scope = ANY(@scope:: TEXT []);

-- name: InsertAggregation :one

INSERT INTO
    aggregations (
        aggregation_name,
        scope,
        product_editor,
        products,
        swidtags,
        created_by
    )
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: UpsertAggregatedRights :exec

INSERT INTO
    aggregated_rights (
        sku,
        scope,
        aggregation_id,
        metric,
        ordering_date,
        corporate_sourcing_contract,
        software_provider,
        num_licenses_acquired,
        avg_unit_price,
        avg_maintenance_unit_price,
        total_purchase_cost,
        total_maintenance_cost,
        total_cost,
        created_by,
        start_of_maintenance,
        end_of_maintenance,
        last_purchased_order,
        support_number,
        maintenance_provider,
        num_licences_maintenance,
        comment,
        file_name,
        file_data,
        repartition
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19,
        $20,
        $21,
        $22,
        $23,
        $24
    ) ON CONFLICT (sku, scope)
DO
UPDATE
SET
    aggregation_id = $3,
    metric = $4,
    ordering_date = $5,
    corporate_sourcing_contract = $6,
    software_provider = $7,
    num_licenses_acquired = $8,
    avg_unit_price = $9,
    avg_maintenance_unit_price = $10,
    total_purchase_cost = $11,
    total_maintenance_cost = $12,
    total_cost = $13,
    updated_by = $14,
    start_of_maintenance = $15,
    end_of_maintenance = $16,
    last_purchased_order = $17,
    support_number = $18,
    maintenance_provider = $19,
    num_licences_maintenance = $20,
    comment = $21,
    file_name = $22,
    file_data = $23,
    repartition = $24;

-- name: UpdateAggregation :exec

UPDATE aggregations
SET
    aggregation_name = @aggregation_name,
    product_editor = @product_editor,
    products = @product_names,
    swidtags = @swidtags,
    updated_by = @updated_by
WHERE id = @id AND scope = @scope;

-- name: DeleteAggregation :exec

DELETE FROM aggregations WHERE id = @id AND scope = @scope;

-- name: ListAggregations :many
SELECT
    count(aggregations.id) OVER() AS totalRecords,
    aggregations.id,
    aggregation_name,
    product_editor,
    products,
    swidtags,
    scope,
    ec.id as editor_id, (
        Select coalesce( (
                    SELECT
                        array_to_json(
                            array_agg(row_to_json(mapping))
                        )
                    from (
                            SELECT
                                products.product_name,
                                products.product_version
                            FROM
                                products
                            where
                                swidtag in (
                                    SELECT
                                        UNNEST (swidtags)
                                )
                            UNION
                            select
                                pc.name product_name,
                                vc.name product_version
                            from
                                product_catalog pc
                                join version_catalog vc on pc.id = vc.p_id
                            where
                                vc.swid_tag_system in (
                                    SELECT
                                        UNNEST (swidtags)
                                )
                            UNION
                            SELECT
                                acqrights.product_name,
                                acqrights.version
                            FROM
                                acqrights
                            where
                                swidtag in (
                                    SELECT
                                        UNNEST (swidtags)
                                )
                        ) as mapping
                ),
                '[]':: json
            )
    )
FROM
    aggregations
    LEFT join editor_catalog as ec on ec.name = aggregations.product_editor
WHERE
    aggregations.scope = @scope
    AND (
        CASE
            WHEN @is_agg_name :: bool THEN lower(aggregation_name) = lower(@aggregation_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @ls_agg_name :: bool THEN lower(aggregation_name) LIKE '%' || lower(@aggregation_name) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_editor :: bool THEN lower(product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_editor :: bool THEN lower(product_editor) LIKE '%' || lower(@product_editor) || '%'
            ELSE TRUE
        END
    )
ORDER BY
    CASE
        WHEN @agg_name_asc :: bool THEN aggregation_name
    END asc,
    CASE
        WHEN @agg_name_desc :: bool THEN aggregation_name
    END desc,
    CASE
        WHEN @product_editor_asc :: bool THEN product_editor
    END asc,
    CASE
        WHEN @product_editor_desc :: bool THEN product_editor
    END desc
LIMIT
    @page_size OFFSET @page_num;

-- name: ListProductsForAggregation :many
SELECT acq.swidtag, acq.product_name, acq.product_editor ,acq.version as product_version
FROM acqrights acq 
WHERE acq.scope = @scope 
AND acq.swidtag NOT IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope) 
AND acq.product_editor = @editor
UNION
SELECT prd.swidtag, prd.product_name, prd.product_editor,prd.product_version
FROM products prd 
WHERE prd.scope = @scope 
AND prd.swidtag NOT IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope)
AND prd.product_editor = @editor
UNION
select vc.swid_tag_system swidtag,pc.name product_name,ec.name as product_editor,vc.name product_version from product_catalog pc 
join version_catalog vc on pc.id = vc.p_id 
left join editor_catalog as ec on pc.editorid = ec.id
WHERE ec.name = @editor
AND vc.swid_tag_system NOT IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope)
;

-- name: ListEditorsForAggregation :many

SELECT
    DISTINCT acq.product_editor
FROM acqrights acq
WHERE
    acq.scope =  ANY(@scope:: TEXT [])
    AND acq.product_editor <> ''
UNION
SELECT DISTINCT prd.product_editor 
FROM products prd 
WHERE prd.scope =ANY(@scope:: TEXT []) AND prd.product_editor <> '';

-- name: ListDeployedAndAcquiredEditors :many

SELECT
    DISTINCT acq.product_editor
FROM acqrights acq
WHERE
    acq.scope = @scope
    AND acq.product_editor <> ''
INTERSECT
SELECT
    DISTINCT prd.product_editor
FROM products prd
WHERE
    prd.scope = @scope
    AND prd.product_editor <> '';

-- name: GetAcqRightsByEditor :many

SELECT
    acq.sku,
    acq.swidtag,
    acq.metric,
    acq.avg_unit_price:: FLOAT AS avg_unit_price,
    acq.num_licenses_acquired
FROM acqrights acq
WHERE
    acq.product_editor = @product_editor
    AND acq.scope = @scope;

-- name: GetAggregationByEditor :many

SELECT
    agg.aggregation_name,
    array_to_string(agg.swidtags, ',') as swidtags,
    COALESCE(agr.sku, '') as sku,
    COALESCE(agr.metric, '') as metric,
    COALESCE(agr.avg_unit_price, 0):: FLOAT AS avg_unit_price,
    COALESCE(agr.num_licenses_acquired, 0)
FROM aggregations agg
    LEFT JOIN (
        SELECT
            a.aggregation_id,
            a.sku,
            a.metric,
            a.avg_unit_price,
            a.num_licenses_acquired
        FROM
            aggregated_rights a
        WHERE
            a.scope = @scope
    ) agr ON agr.aggregation_id = agg.id
WHERE
    agg.scope = @scope
    AND agg.product_editor = @product_editor;

-- name: ListMetricsForAggregation :many

SELECT DISTINCT acq.metric FROM acqrights acq WHERE acq.scope = $1;

-- name: GetLicensesCost :one

SELECT
    COALESCE(SUM(total_cost), 0):: Numeric(15, 2) as total_cost,
    COALESCE(
        SUM(total_maintenance_cost),
        0
    ):: Numeric(15, 2) as total_maintenance_cost
FROM (
        SELECT
            SUM(total_cost):: Numeric(15, 2) as total_cost,
            SUM(total_maintenance_cost):: Numeric(15, 2) as total_maintenance_cost
        FROM acqrights
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY scope
        UNION ALL
        SELECT
            SUM(total_cost):: Numeric(15, 2) as total_cost,
            SUM(total_maintenance_cost):: Numeric(15, 2) as total_maintenance_cost
        FROM
            aggregated_rights
        WHERE
            scope = ANY(@scope:: TEXT [])
        GROUP BY scope
    ) a;

-- name: GetAcqRightsCost :one

SELECT
    SUM(total_cost):: Numeric(15, 2) as total_cost,
    SUM(total_maintenance_cost):: Numeric(15, 2) as total_maintenance_cost
from acqrights
WHERE
    scope = ANY(@scope:: TEXT [])
GROUP BY scope;

-- name: GetApplicationsByProductID :many

SELECT application_id
from products_applications
WHERE
    scope = @scope
    AND swidtag = @swidtag;

-- name: GetProductCount :many

SELECT
    COUNT(pa.swidtag) as num_of_products,
    pa.application_id
FROM products_applications pa
WHERE pa.scope = @scope
GROUP BY pa.application_id;

-- name: GetProductsByApplicationID :many

SELECT swidtag
from products_applications
WHERE
    scope = @scope
    AND application_id = @application_id;

-- name: ProductsPerMetric :many

SELECT
    x.metric as metric,
    SUM(x.composition) as composition
FROM(
        SELECT
            acq.metric,
            count(acq.swidtag) as composition
        FROM acqrights acq
        Where
            acq.scope = @scope
        GROUP BY acq.metric
        UNION ALL
        SELECT
            agr.metric,
            count(agr.aggregation_id) as composition
        FROM
            aggregated_rights agr
        Where
            agr.scope = @scope
        GROUP BY agr.metric
    ) x
GROUP BY x.metric;

-- name: CounterFeitedProductsLicences :many

SELECT
    swidtags as swid_tags,
    product_names as product_names,
    aggregation_name as aggregation_name,
    num_computed_licences:: Numeric(15, 2) as num_computed_licences,
    num_acquired_licences:: Numeric(15, 2) as num_acquired_licences, (delta_number):: Numeric(15, 2) as delta
FROM
    overall_computed_licences
WHERE
    scope = @scope
    AND editor = @editor
    AND cost_optimization = FALSE
GROUP BY
    swidtags,
    product_names,
    aggregation_name,
    num_acquired_licences,
    num_computed_licences,
    delta_number
HAVING (delta_number) < 0
ORDER BY delta ASC
LIMIT 5;

-- name: CounterFeitedProductsCosts :many

SELECT
    swidtags as swid_tags,
    product_names as product_names,
    aggregation_name as aggregation_name,
    computed_cost:: Numeric(15, 2) as computed_cost,
    purchase_cost:: Numeric(15, 2) as purchase_cost, (delta_cost):: Numeric(15, 2) as delta_cost
FROM
    overall_computed_licences
WHERE
    scope = @scope
    AND editor = @editor
    AND cost_optimization = FALSE
GROUP BY
    swidtags,
    product_names,
    aggregation_name,
    purchase_cost,
    computed_cost,
    delta_cost
HAVING (delta_cost) < 0
ORDER BY delta_cost ASC
LIMIT 5;

-- name: OverDeployedProductsLicences :many

SELECT
    swidtags as swid_tags,
    product_names as product_names,
    aggregation_name as aggregation_name,
    num_computed_licences:: Numeric(15, 2) as num_computed_licences,
    num_acquired_licences:: Numeric(15, 2) as num_acquired_licences, (delta_number):: Numeric(15, 2) as delta
FROM
    overall_computed_licences
WHERE
    scope = @scope
    AND editor = @editor
    AND cost_optimization = FALSE
GROUP BY
    swidtags,
    product_names,
    aggregation_name,
    num_acquired_licences,
    num_computed_licences,
    delta_number
HAVING (delta_number) > 0
ORDER BY delta ASC
LIMIT 5;

-- name: OverDeployedProductsCosts :many

SELECT
    swidtags as swid_tags,
    product_names as product_names,
    aggregation_name as aggregation_name,
    computed_cost:: Numeric(15, 2) as computed_cost,
    purchase_cost:: Numeric(15, 2) as purchase_cost, (delta_cost):: Numeric(15, 2) as delta_cost
FROM
    overall_computed_licences
WHERE
    scope = @scope
    AND editor = @editor
    AND cost_optimization = FALSE
GROUP BY
    swidtags,
    product_names,
    aggregation_name,
    purchase_cost,
    computed_cost,
    delta_cost
HAVING (delta_cost) > 0
ORDER BY delta_cost ASC
LIMIT 5;

-- name: ListAcqrightsProducts :many

SELECT DISTINCT swidtag,scope FROM acqrights;

-- name: ListAcqrightsProductsByScope :many

SELECT DISTINCT swidtag, scope
FROM acqrights
where
    acqrights.scope = $1
    and swidtag not in (
        select
            unnest(aggregations.swidtags)
        from aggregations
            inner join aggregated_rights on aggregations.id = aggregated_rights.aggregation_id and aggregations.scope = $1
    );

-- name: AddComputedLicenses :exec

UPDATE acqrights
SET
    num_licences_computed = @computedLicenses,
    total_computed_cost = @computedCost
WHERE sku = @sku AND scope = @scope;

-- name: CounterfeitPercent :one

select
    coalesce(sum(num_acquired_licences), 0):: Numeric(15, 2) as acq,
    coalesce(abs(sum(delta_number)), 0):: Numeric(15, 2) as delta_rights
from
    overall_computed_licences ocl
where
    ocl.scope = @scope
    AND ocl.delta_number < 0;

-- name: OverdeployPercent :one

select
    coalesce(sum(num_acquired_licences), 0):: Numeric(15, 2) as acq,
    coalesce(abs(sum(delta_number)), 0):: Numeric(15, 2) as delta_rights
from
    overall_computed_licences ocl
where
    ocl.scope = @scope
    AND ocl.delta_number >= 0;

-- name: UpsertProductApplications :exec

Insert into
    products_applications (swidtag, application_id, scope)
Values ($1, $2, $3) ON CONFLICT (swidtag, application_id, scope)
Do NOTHING;

-- name: UpsertProductEquipments :exec

Insert into
    products_equipments (
        swidtag,
        equipment_id,
        num_of_users,
        scope,
        allocated_metric
    )
Values ($1, $2, $3, $4, $5) ON CONFLICT (swidtag, equipment_id, scope)
Do
Update
set
    num_of_users = $3,
    allocated_metric = $5;

-- name: ProductsNotDeployed :many
SELECT DISTINCT(acqrights.swidtag), acqrights.product_name, acqrights.product_editor,ec.id, version FROM acqrights 
LEFT JOIN products as p on p.product_name=acqrights.product_name AND p.product_editor=acqrights.product_editor 
Left Join editor_catalog as ec on ec.name = acqrights.product_editor
WHERE acqrights.swidtag NOT IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND acqrights.scope = @scope AND p.product_type='ONPREMISE' ;

-- name: ProductsNotAcquired :many
SELECT swidtag, product_name, product_editor, product_version,ec.id  FROM products
Left Join editor_catalog as ec on ec.name = products.product_editor
WHERE products.swidtag NOT IN (SELECT swidtag FROM acqrights WHERE acqrights.scope = @scope UNION SELECT unnest(swidtags) as swidtags from aggregations INNER JOIN aggregated_rights ON aggregations.id = aggregated_rights.aggregation_id where aggregations.scope = @scope)
AND products.product_name NOT IN (  SELECT product_name FROM acqrights WHERE acqrights.scope =  @scope)
AND products.swidtag IN (SELECT swidtag FROM products_equipments WHERE products_equipments.scope = @scope)
AND products.scope = @scope;

-- name: DeleteProductsByScope :exec

DELETE FROM products WHERE scope = @scope;

-- name: DeleteAcqrightsByScope :exec

DELETE FROM acqrights WHERE scope = @scope;

-- name: DeleteSharedDataByScope :exec

DELETE FROM shared_licenses WHERE scope = @scope;

-- name: DeleteAggregationByScope :exec

DELETE FROM aggregations WHERE scope = @scope;

-- name: DeleteAggregatedRightsByScope :exec

DELETE FROM aggregated_rights WHERE scope = @scope;

-- name: GetAggregationByName :one

SELECT *
FROM aggregations
WHERE
    aggregation_name = @aggregation_name
    AND scope = @scope;

-- name: GetAggregatedRightBySKU :one

SELECT
    sku,
    aggregation_id,
    metric,
    ordering_date,
    corporate_sourcing_contract,
    software_provider,
    scope,
    num_licenses_acquired,
    num_licences_computed,
    num_licences_maintenance,
    avg_unit_price,
    avg_maintenance_unit_price,
    total_purchase_cost,
    total_computed_cost,
    total_maintenance_cost,
    total_cost,
    start_of_maintenance,
    end_of_maintenance,
    last_purchased_order,
    support_number,
    maintenance_provider,
    comment,
    file_name
FROM aggregated_rights
WHERE sku = @sku AND scope = @scope;

-- name: GetAggregatedRightsFileDataBySKU :one

SELECT file_data
FROM aggregated_rights
WHERE sku = @sku and scope = @scope;

-- name: GetAcqRightBySKU :one

SELECT
    sku,
    swidtag,
    product_name,
    product_editor,
    metric,
    num_licenses_acquired,
    num_licences_maintainance,
    avg_unit_price,
    avg_maintenance_unit_price,
    total_purchase_cost,
    total_maintenance_cost,
    total_cost,
    start_of_maintenance,
    end_of_maintenance,
    version,
    comment,
    ordering_date,
    software_provider,
    corporate_sourcing_contract,
    last_purchased_order,
    support_number,
    maintenance_provider,
    file_name,
    file_data,
    repartition
FROM acqrights
WHERE
    sku = @acqright_sku
    and scope = @scope;

-- name: GetAcqRightFileDataBySKU :one

SELECT file_data
FROM acqrights
WHERE
    sku = @acqright_sku
    and scope = @scope;

-- name: GetUnitPriceBySku :one

Select
    ac.sku,
    ac.avg_unit_price
from acqrights ac
where
    ac.scope = @scope
    and ac.sku = @sku
union
SELECT
    ag.sku,
    ag.avg_unit_price
from aggregated_rights ag
where
    ag.scope = @scope
    and ag.sku = @sku;

-- name: UpsertDashboardUpdates :exec

Insert into
    dashboard_audit (
        updated_at,
        next_update_at,
        updated_by,
        scope
    )
values ($1, $2, $3, $4) on CONFLICT (scope)
Do
update
set
    updated_at = $1,
    next_update_at = $2,
    updated_by = $3;

-- name: GetDashboardUpdates :one

select
    updated_at at time zone $2:: varchar as updated_at,
    next_update_at at time zone $2:: varchar as next_update_at
from dashboard_audit
where scope = $1;

-- name: DeleteAcqrightBySKU :exec

DELETE FROM acqrights WHERE scope = @scope AND sku = @sku;

-- name: DeleteAggregatedRightBySKU :exec

DELETE FROM aggregated_rights WHERE scope = @scope AND sku = @sku;

-- name: GetEquipmentsBySwidtag :one

SELECT
    ARRAY_AGG(DISTINCT(equipment_id)):: TEXT [] as equipments
from products_equipments
WHERE
    scope = @scope
    and swidtag = @swidtag;

-- name: GetAcqBySwidtags :many

Select
    sku,
    swidtag,
    product_name,
    product_editor,
    metric,
    num_licenses_acquired,
    num_licences_maintainance,
    avg_unit_price,
    avg_maintenance_unit_price,
    total_purchase_cost,
    total_maintenance_cost,
    total_cost,
    start_of_maintenance,
    end_of_maintenance,
    version,
    comment,
    ordering_date,
    software_provider,
    corporate_sourcing_contract,
    last_purchased_order,
    support_number,
    maintenance_provider,
    file_name
from acqrights
where
    swidtag = ANY(@swidtag:: TEXT [])
    and scope = @scope
    AND (
        CASE
            WHEN @is_metric:: bool THEN lower(metric) = lower(@metric)
            ELSE TRUE
        END
    );

-- name: GetAcqBySwidtag :one

Select *
from acqrights
where
    swidtag = @swidtag
    and scope = @scope;

-- name: GetIndividualProductDetailByAggregation :many

Select
    agg.aggregation_name,
    coalesce(num_of_applications, 0) as num_of_applications,
    coalesce(num_of_equipments, 0) as num_of_equipments,
    agg.product_editor,
    COALESCE(pname,'')::TEXT as name,
    COALESCE(pversion,'')::TEXT as version,
    prod_id,
    COALESCE(ar.total_cost,0)::NUMERIC(15,2) as total_cost
from
	aggregations agg
	LEFT JOIN (
		select 
			p.product_name as pname,
			p.product_version as pversion,
			p.swidtag as prod_id
			from products p 
		    where 
			p.scope =  @scope
        UNION
		select
			pc.name as pname,
			v.name as pversion,
			v.swid_tag_system as prod_id
			from product_catalog pc
			left join version_catalog v on pc.id =v.p_id
	) p on prod_id  = ANY(agg.swidtags::TEXT[])
    LEFT JOIN (
        SELECT
            pa.swidtag,
            count(application_id) as num_of_applications
        FROM
            products_applications pa
        WHERE pa.scope = @scope
        GROUP BY
            pa.swidtag
    ) pa ON prod_id = pa.swidtag
    LEFT JOIN (
        SELECT
            pe.swidtag,
            count(equipment_id) as num_of_equipments
        FROM
            products_equipments pe
        WHERE pe.scope = @scope
        GROUP BY
            pe.swidtag
    ) pe ON prod_id = pe.swidtag
    LEFT JOIN (
        SELECT
            a.aggregation_id,
            SUM(a.total_cost):: Numeric(15, 2) as total_cost
        FROM
            aggregated_rights a
        WHERE a.scope = @scope
        GROUP BY
            a.aggregation_id
    ) ar ON ar.aggregation_id = agg.id
WHERE
    agg.scope = @scope
    and agg.aggregation_name = @aggregation_name;

-- name: ListSelectedProductsForAggregration :many
SELECT acq.swidtag, acq.product_name, acq.product_editor,acq.version as product_version
FROM acqrights acq 
WHERE acq.scope = @scope 
AND acq.swidtag IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope and agg.id = @id) 
UNION
SELECT prd.swidtag, prd.product_name, prd.product_editor, prd.product_version
FROM products prd 
WHERE prd.scope = @scope 
AND prd.swidtag IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope and agg.id = @id)
UNION
select vc.swid_tag_system swidtag,pc.name product_name,ec.name as product_editor,vc.name product_version from product_catalog pc 
join version_catalog vc on pc.id = vc.p_id 
left join editor_catalog ec on pc.editorid = ec.id
WHERE ec.name =  @editor
AND vc.swid_tag_system  IN (SELECT UNNEST(agg.swidtags) from aggregations agg where agg.scope = @scope and agg.id = @id);


-- name: GetAggregationByID :one

SELECT * FROM aggregations WHERE id = @id AND scope = @scope;

-- name: GetTotalCounterfietAmount :one

select
    coalesce(sum(ocl.delta_cost), 0.0):: FLOAT as counterfiet_amount
from
    overall_computed_licences ocl
where
    ocl.scope = @scope
    AND ocl.delta_cost < 0
    AND cost_optimization = FALSE;

-- name: GetScopeCounterfietAmountEditor :many
SELECT coalesce(sum(ocl.delta_cost), 0.0) :: FLOAT AS cost, ocl.scope
FROM overall_computed_licences ocl
WHERE
 ocl.delta_cost < 0
 AND cost_optimization = FALSE
 AND editor = $2
 AND scope = ANY($1::text[])
GROUP BY
 ocl.scope;

-- name: GetScopeTotalAmountEditor :many
SELECT
    editorExpenseByScopeData.scope,
    coalesce(SUM(total_cost), 0.0):: FLOAT as cost
FROM (
        SELECT
            scope,
            total_cost
        FROM acqrights
        WHERE
            acqrights.scope = ANY($1::text[]) and acqrights.product_editor=$2 
        UNION ALL
        SELECT
            a.scope,
            ar.total_cost
        FROM aggregations as a
            INNER JOIN aggregated_rights as ar ON a.id = ar.aggregation_id
        WHERE
            a.scope = ANY($1::text[]) and
            a.product_editor = $2
    ) as editorExpenseByScopeData
GROUP BY editorExpenseByScopeData.scope;

-- name: GetScopeUnderUsageCostEditor :many
SELECT coalesce(sum(ocl.delta_cost), 0.0) :: FLOAT AS cost, ocl.scope
FROM overall_computed_licences ocl
WHERE
 ocl.delta_cost > 0
 AND cost_optimization = FALSE
 AND editor = $2
 AND scope = ANY($1::text[])
GROUP BY
 ocl.scope;

-- name: GetTotalUnderusageAmount :one

select
    coalesce(sum(ocl.delta_cost), 0.0):: FLOAT as underusage_amount
from
    overall_computed_licences ocl
where
    ocl.scope = @scope
    AND ocl.delta_cost > 0
    AND cost_optimization = FALSE;

-- name: GetTotalDeltaCost :one

select
    coalesce(sum(ocl.delta_cost), 0.0):: FLOAT
FROM
    overall_computed_licences ocl
where
    cost_optimization = TRUE
    AND ocl.scope = @scope;

-- name: ListAggregationNameByScope :many

SELECT aggregation_name
from aggregations
    inner join aggregated_rights on aggregations.id = aggregated_rights.aggregation_id
where aggregations.scope = $1;

-- name: ListAggregationNameWithScope :many

SELECT
    aggregation_name,
    aggregations.scope
from aggregations
    inner join aggregated_rights on aggregations.id = aggregated_rights.aggregation_id;

-- name: AddComputedLicensesToAggregation :exec

UPDATE aggregated_rights
SET
    num_licences_computed = @computedLicenses,
    total_computed_cost = @computedCost
WHERE sku = @sku AND scope = @scope;

-- name: GetAcqRightMetricsBySwidtag :many

SELECT sku, metric
FROM acqrights
WHERE
    scope = @scope
    AND swidtag = @swidtag;

-- name: GetAggRightMetricsByAggregationId :many

SELECT sku, metric
FROM aggregated_rights
WHERE
    scope = @scope
    AND aggregation_id = @agg_id;

-- name: GetMetricsBySku :one

Select sku, metric
from acqrights acq
where
    acq.sku = @sku
    AND acq.scope = @scope
UNION
select sku, metric
from aggregated_rights agg
where
    agg.sku = @sku
    AND agg.scope = @scope;

-- name: GetIndividualProductForAggregationCount :one
SELECT count(*) 
FROM (
    Select p.product_name from products p WHERE p.scope = @scope 
                                                AND p.swidtag = ANY(@swidtags::TEXT[])
    UNION
    select v.name from version_catalog v WHERE v.swid_tag_system = ANY(@swidtags::TEXT[]) AND v.name != '') t1;

-- name: InsertOverAllComputedLicences :exec

INSERT INTO
    overall_computed_licences (
        sku,
        swidtags,
        scope,
        product_names,
        aggregation_name,
        metrics,
        num_computed_licences,
        num_acquired_licences,
        total_cost,
        purchase_cost,
        computed_cost,
        delta_number,
        cost_optimization,
        delta_cost,
        avg_unit_price,
        computed_details,
        metic_not_defined,
        not_deployed,
        editor
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19
    ) ON CONFLICT (
        sku,
        swidtags,
        scope
    )
DO
UPDATE
SET
    product_names = $4,
    aggregation_name = $5,
    metrics = $6,
    num_computed_licences = $7,
    num_acquired_licences = $8,
    total_cost = $9,
    purchase_cost = $10,
    computed_cost = $11,
    delta_number = $12,
    cost_optimization = $13,
    delta_cost = $14,
    computed_details = $16,
    metic_not_defined = $17,
    not_deployed = $18;

-- name: DeleteOverallComputedLicensesByScope :exec

DELETE FROM overall_computed_licences WHERE scope = @scope;

-- name: DropAllocatedMetricFromEquipment :exec

UPDATE products_equipments
SET
    allocated_metric = @allocated_metric
WHERE
    swidtag = @swidtag
    AND scope = @scope
    AND equipment_id = @equipment_id;

-- name: TotalCostOfEachScope :many
    
SELECT
    COALESCE(SUM(total_cost), 0)::NUMERIC(15, 2) AS total_cost,
    a.scope
FROM (
    SELECT
        SUM(total_cost)::NUMERIC(15, 2) AS total_cost,
        scope
    FROM acqrights
    WHERE
        scope = ANY(@scope::TEXT[])
    GROUP BY scope
    UNION ALL
    SELECT
        SUM(total_cost)::NUMERIC(15, 2) AS total_cost,
        scope
    FROM aggregated_rights
    WHERE
        scope = ANY(@scope::TEXT[])
    GROUP BY scope
) a
GROUP BY a.scope;

-- name: GetOverallLicencesByProduct :many

SELECT
    scope,
    COALESCE(SUM(num_computed_licences),0) :: Numeric(15, 2) as computed_licences,
    COALESCE(SUM(num_acquired_licences),0) :: Numeric(15, 2) as acquired_licences
FROM
    overall_computed_licences
WHERE
    product_names = @product_name
    AND editor = @editor
    AND scope = ANY(@scope :: TEXT [])
GROUP BY
    scope;

-- name: GetOverallCostByProduct :many

SELECT
    scope,
    COALESCE(SUM(total_cost),0) :: Numeric(15, 2) as total_cost,
    COALESCE(SUM(
        CASE
            When delta_cost < 0 Then delta_cost
        End
    ),0) :: Numeric(15, 2) as counterfeiting_cost,
    COALESCE(SUM(
        CASE
            When delta_cost > 0 Then delta_cost
        End
    ),0) :: Numeric(15, 2) as underusage_cost
FROM
    overall_computed_licences
WHERE
    product_names = @product_name
    AND editor = @editor
    AND scope = ANY(@scope :: TEXT [])
GROUP BY
    scope;

-- name: GetTotalCostByProduct :many
SELECT
    editorExpenseByScopeData.scope,
    COALESCE(SUM(total_cost), 0.0)::FLOAT AS total_cost
FROM (
    SELECT
        scope,
        total_cost
    FROM acqrights
    WHERE
        acqrights.scope = ANY(@scope :: TEXT []) AND
        acqrights.product_editor = @editor AND acqrights.product_name = @product_name
        group by scope,total_cost
    UNION ALL
    SELECT
        a.scope,
        ar.total_cost
    FROM aggregations AS a
    INNER JOIN aggregated_rights AS ar ON a.id = ar.aggregation_id
    WHERE
        a.scope = ANY(@scope :: TEXT []) AND
        a.product_editor = @editor AND @product_name = ANY(a.products)
        group by a.scope,total_cost
) AS editorExpenseByScopeData
GROUP BY editorExpenseByScopeData.scope;

-- name: GetProductListByEditor :many

SELECT
    DISTINCT p.product_name
FROM
    products as p
where
    p.product_editor = @editor
    AND p.scope = ANY(@scope :: TEXT [])
UNION
SELECT
    acq.product_name
FROM
    acqrights as acq
where
    acq.product_editor = @editor
    AND acq.scope = ANY(@scope :: TEXT []);

-- name: GetAvailableAcqLicenses :one

SELECT
    COALESCE(num_licenses_acquired, 0):: INTEGER as acquired_licences
FROM acqrights
WHERE sku = @sku AND scope = @scope
GROUP BY
    num_licenses_acquired;

-- name: GetAvailableAggLicenses :one

SELECT
    COALESCE(num_licenses_acquired, 0):: INTEGER as acquired_licences
FROM aggregated_rights
WHERE sku = @sku AND scope = @scope
GROUP BY
    num_licenses_acquired;

-- name: GetSharedLicenses :many

Select *
FROM shared_licenses
where sku = @sku AND scope = @scope
Group By
    sku,
    scope,
    sharing_scope
HAVING
    shared_licences > 0
    OR recieved_licences > 0;

-- name: GetTotalSharedLicenses :one

Select
    COALESCE(SUM(shared_licences), 0):: INTEGER as total_shared_licences,
    COALESCE(SUM(recieved_licences), 0):: INTEGER as total_recieved_licences
FROM shared_licenses
where sku = @sku AND scope = @scope;

-- name: UpsertSharedLicenses :exec

INSERT INTO
    shared_licenses (
        sku,
        scope,
        sharing_scope,
        shared_licences
    )
VALUES ($1, $2, $3, $4) ON CONFLICT (sku, scope, sharing_scope)
Do
Update
set shared_licences = $4;

-- name: UpsertRecievedLicenses :exec

INSERT INTO
    shared_licenses (
        sku,
        scope,
        sharing_scope,
        recieved_licences
    )
VALUES ($1, $2, $3, $4) ON CONFLICT (sku, scope, sharing_scope)
Do
Update
set recieved_licences = $4;

-- name: GetSharedData :many

Select * from shared_licenses where scope = @scope;

-- name: DeleteSharedLicences :exec

Delete from shared_licenses where sku = @sku AND scope = @scope;

-- name: UpsertProductNominativeUser :exec

INSERT INTO
    nominative_user (
        scope,
        swidtag,
        aggregations_id,
        activation_date,
        user_email,
        user_name,
        first_name,
        profile,
        product_editor,
        created_by,
        updated_by,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13
    ) ON CONFLICT (
        swidtag,
        scope,
        user_email,
        profile
    )
DO
UPDATE
SET
    activation_date = $4,
    user_name = $6,
    first_name = $7,
    updated_by = $11,
    updated_at = $13;

-- name: UpsertAggrigationNominativeUser :exec

INSERT INTO
    nominative_user (
        scope,
        swidtag,
        aggregations_id,
        activation_date,
        user_email,
        user_name,
        first_name,
        profile,
        product_editor,
        created_by,
        updated_by,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13
    ) ON CONFLICT (
        aggregations_id,
        scope,
        user_email,
        profile
    )
DO
UPDATE
SET
    activation_date = $4,
    user_name = $6,
    first_name = $7,
    updated_by = $11,
    updated_at = $13;

-- name: ListNominativeUsersProducts :many

SELECT
    count(*) OVER() AS totalRecords,
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    p.product_version,
    p.product_name
FROM nominative_user nU
    INNER JOIN products p on nu.swidtag = p.swidtag and nu.scope = ANY(@scope:: TEXT []) AND p.scope = ANY(@scope:: TEXT [])
where
    nu.swidtag != ''
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_version:: bool THEN lower(p.product_version) LIKE '%' || lower(@product_version:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_version:: bool THEN lower(p.product_version) = lower(@product_version)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_name:: bool THEN lower(nu.user_name) LIKE '%' || lower(@user_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_name:: bool THEN lower(nu.user_name) = lower(@user_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_first_name:: bool THEN lower(nu.first_name) LIKE '%' || lower(@first_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_first_name:: bool THEN lower(nu.first_name) = lower(@first_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_email:: bool THEN lower(nu.user_email) LIKE '%' || lower(@user_email:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_email:: bool THEN lower(nu.user_email) = lower(@user_email)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_profile:: bool THEN lower(nu.profile) LIKE '%' || lower(@profile:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_profile:: bool THEN lower(nu.profile) = lower(@profile)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_activation_date:: bool THEN date(nu.activation_date):: text = @activation_date:: text
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_editor:: bool THEN lower(nu.product_editor) LIKE '%' || lower(@product_editor:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_editor:: bool THEN lower(nu.product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
GROUP BY
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    p.product_version,
    p.product_name
ORDER BY
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @user_name_asc:: bool THEN nu.user_name
    END asc,
    CASE
        WHEN @user_name_desc:: bool THEN nu.user_name
    END desc,
    CASE
        WHEN @first_name_asc:: bool THEN nu.first_name
    END asc,
    CASE
        WHEN @first_name_desc:: bool THEN nu.first_name
    END desc,
    CASE
        WHEN @user_email_asc:: bool THEN nu.user_email
    END asc,
    CASE
        WHEN @user_email_desc:: bool THEN nu.user_email
    END desc,
    CASE
        WHEN @profile_asc:: bool THEN nu.profile
    END asc,
    CASE
        WHEN @profile_desc:: bool THEN nu.profile
    END desc,
    CASE
        WHEN @activation_date_asc:: bool THEN nu.activation_date
    END asc,
    CASE
        WHEN @activation_date_desc:: bool THEN nu.activation_date
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN nu.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN nu.product_editor
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: ListNominativeUsersAggregation :many

SELECT
    count(*) OVER() AS totalRecords,
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    nu.aggregations_id,
    agg.aggregation_name
FROM nominative_user nU
    LEFT JOIN aggregations agg on nu.aggregations_id = agg.id
WHERE
    nu.aggregations_id != 0
    AND nu.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @lk_aggregation_name:: bool THEN lower(agg.aggregation_name) LIKE '%' || lower(@aggregation_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_aggregation_name:: bool THEN lower(agg.aggregation_name) = lower(@aggregation_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_name:: bool THEN lower(nu.user_name) LIKE '%' || lower(@user_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_name:: bool THEN lower(nu.user_name) = lower(@user_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_first_name:: bool THEN lower(nu.first_name) LIKE '%' || lower(@first_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_first_name:: bool THEN lower(nu.first_name) = lower(@first_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_email:: bool THEN lower(nu.user_email) LIKE '%' || lower(@user_email:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_email:: bool THEN lower(nu.user_email) = lower(@user_email)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_profile:: bool THEN lower(nu.profile) LIKE '%' || lower(@profile:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_profile:: bool THEN lower(nu.profile) = lower(@profile)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_activation_date:: bool THEN date(nu.activation_date):: text = @activation_date:: text
            ELSE TRUE
        END
    )
GROUP BY
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    nu.aggregations_id,
    agg.aggregation_name
ORDER BY
    CASE
        WHEN @aggregation_name_asc:: bool THEN agg.aggregation_name
    END asc,
    CASE
        WHEN @aggregation_name_desc:: bool THEN agg.aggregation_name
    END desc,
    CASE
        WHEN @user_name_asc:: bool THEN nu.user_name
    END asc,
    CASE
        WHEN @user_name_desc:: bool THEN nu.user_name
    END desc,
    CASE
        WHEN @first_name_asc:: bool THEN nu.first_name
    END asc,
    CASE
        WHEN @first_name_desc:: bool THEN nu.first_name
    END desc,
    CASE
        WHEN @user_email_asc:: bool THEN nu.user_email
    END asc,
    CASE
        WHEN @user_email_desc:: bool THEN nu.user_email
    END desc,
    CASE
        WHEN @profile_asc:: bool THEN nu.profile
    END asc,
    CASE
        WHEN @profile_desc:: bool THEN nu.profile
    END desc,
    CASE
        WHEN @activation_date_asc:: bool THEN nu.activation_date
    END asc,
    CASE
        WHEN @activation_date_desc:: bool THEN nu.activation_date
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: UpsertConcurrentUser :exec

INSERT INTO
    product_concurrent_user (
        is_aggregations,
        aggregation_id,
        swidtag,
        number_of_users,
        profile_user,
        team,
        scope,
        purchase_date,
        created_by,
        updated_by,
        created_on,
        updated_on
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    ) ON CONFLICT (swidtag, scope, purchase_date)
DO
UPDATE
SET
    number_of_users = $4,
    profile_user = $5,
    team = $6,
    updated_by = $10,
    updated_on = $12;

-- name: UpsertAggregationConcurrentUser :exec

INSERT INTO
    product_concurrent_user (
        is_aggregations,
        aggregation_id,
        swidtag,
        number_of_users,
        profile_user,
        team,
        scope,
        purchase_date,
        created_by,
        updated_by,
        created_on,
        updated_on
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    ) ON CONFLICT (
        aggregation_id,
        scope,
        purchase_date
    )
DO
UPDATE
SET
    number_of_users = $4,
    profile_user = $5,
    team = $6,
    updated_by = $10,
    updated_on = $12;

-- name: GetConcurrentUserByID :one

SELECT *
FROM product_concurrent_user
WHERE scope = @scope AND id = @id;

-- name: ListConcurrentUsers :many

SELECT
    count(*) OVER() AS totalRecords,
    pcu.id,
    pcu.swidtag,
    pcu.purchase_date,
    pcu.number_of_users,
    pcu.profile_user,
    pcu.team,
    pcu.updated_on,
    pcu.created_on,
    pcu.created_by,
    pcu.updated_by,
    pcu.aggregation_id,
    agg.aggregation_name,
    p.product_version,
    p.product_name,
    p.product_editor,
    pcu.is_aggregations
FROM
    product_concurrent_user pcu
    LEFT JOIN aggregations agg on pcu.aggregation_id = agg.id
    LEFT JOIN products p on pcu.swidtag = p.swidtag
    AND p.scope = ANY(@scope:: TEXT [])
WHERE
    pcu.scope = ANY(@scope:: TEXT [])   
    AND pcu.is_aggregations = @is_aggregations:: bool
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_aggregation_name:: bool THEN lower(agg.aggregation_name) LIKE '%' || lower(@aggregation_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_aggregation_name:: bool THEN lower(agg.aggregation_name) = lower(@aggregation_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_version:: bool THEN lower(p.product_version) LIKE '%' || lower(@product_version:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_version:: bool THEN lower(p.product_version) = lower(@product_version)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_profile_user:: bool THEN lower(pcu.profile_user) LIKE '%' || lower(@profile_user:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_profile_user:: bool THEN lower(pcu.profile_user) = lower(@profile_user)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_team:: bool THEN lower(pcu.team) LIKE '%' || lower(@team:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_team:: bool THEN lower(pcu.team) = lower(@team)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_number_of_users:: bool THEN lower(pcu.number_of_users) LIKE '%' || lower(@number_of_users:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_number_of_users:: bool THEN lower(pcu.number_of_users) = lower(@number_of_users)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_purchase_date:: bool THEN DATE(pcu.updated_on) = DATE(@purchase_date)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_editor_name:: bool THEN lower(p.product_editor) LIKE '%' || lower(@product_editor:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_editor_name:: bool THEN lower(p.product_editor) = lower(@product_editor)
            ELSE TRUE
        END
    )
GROUP BY
    pcu.id,
    pcu.swidtag,
    pcu.purchase_date,
    pcu.number_of_users,
    pcu.profile_user,
    pcu.team,
    pcu.updated_on,
    pcu.created_on,
    pcu.created_by,
    pcu.updated_by,
    pcu.aggregation_id,
    agg.aggregation_name,
    p.product_version,
    p.product_name,
    p.product_editor
ORDER BY
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @aggregation_name_asc:: bool THEN agg.aggregation_name
    END asc,
    CASE
        WHEN @aggregation_name_desc:: bool THEN agg.aggregation_name
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @profile_user_asc:: bool THEN pcu.profile_user
    END asc,
    CASE
        WHEN @profile_user_desc:: bool THEN pcu.profile_user
    END desc,
    CASE
        WHEN @team_asc:: bool THEN pcu.team
    END asc,
    CASE
        WHEN @team_desc:: bool THEN pcu.team
    END desc,
    CASE
        WHEN @number_of_users_asc:: bool THEN pcu.number_of_users
    END asc,
    CASE
        WHEN @number_of_users_desc:: bool THEN pcu.number_of_users
    END desc,
    CASE
        WHEN @purchase_date_asc:: bool THEN pcu.updated_on
    END asc,
    CASE
        WHEN @purchase_date_desc:: bool THEN pcu.updated_on
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN p.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN p.product_editor
    END desc
LIMIT @page_size
OFFSET @page_num;

-- name: DeletConcurrentUserByID :exec

DELETE FROM
    product_concurrent_user
WHERE scope = @scope AND id = @id;

-- name: GetNominativeUserByID :one

SELECT * FROM nominative_user WHERE scope = @scope AND user_id = @id;

-- name: DeleteNominativeUserByID :exec

DELETE FROM nominative_user WHERE scope = @scope AND user_id = @id;

-- name: GetConcurrentUsersByMonth :many

SELECT
    CONCAT(
        TO_CHAR(purchase_date, 'Month'),
        EXTRACT(
            YEAR
            FROM
                purchase_date
        )
    ) as purchaseMonthYear,
    SUM(number_of_users) AS totalConUsers
FROM product_concurrent_user
WHERE
    scope = @scope
    AND (
        CASE
            WHEN @is_purchase_start_date:: bool THEN purchase_date >= @start_date
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_purchase_end_date:: bool THEN purchase_date <= @end_date
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_swidtag:: bool THEN swidtag = @swidtag
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_aggregation_id:: bool THEN aggregation_id = @aggregation_id
            ELSE TRUE
        END
    )
GROUP BY purchaseMonthYear;

-- name: GetConcurrentUsersByDay :many

SELECT
    purchase_date,
    SUM(number_of_users) AS totalConUsers
FROM product_concurrent_user
WHERE
    scope = @scope
    AND (
        CASE
            WHEN @is_purchase_start_date:: bool THEN purchase_date >= @start_date
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_purchase_end_date:: bool THEN purchase_date <= @end_date
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_swidtag:: bool THEN swidtag = @swidtag
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_aggregation_id:: bool THEN aggregation_id = @aggregation_id
            ELSE TRUE
        END
    )
GROUP BY purchase_date;

-- name: ExportConcurrentUsers :many

SELECT
    count(*) OVER() AS totalRecords,
    pcu.id,
    pcu.swidtag,
    pcu.purchase_date,
    pcu.number_of_users,
    pcu.profile_user,
    pcu.team,
    pcu.updated_on,
    pcu.created_on,
    pcu.created_by,
    pcu.updated_by,
    pcu.aggregation_id,
    agg.aggregation_name,
    p.product_version,
    p.product_name,
    pcu.is_aggregations
FROM
    product_concurrent_user pcu
    LEFT JOIN aggregations agg on pcu.aggregation_id = agg.id
    LEFT JOIN products p on pcu.swidtag = p.swidtag
    AND p.scope = ANY(@scope:: TEXT [])
WHERE
    pcu.scope = ANY(@scope:: TEXT [])
    AND pcu.is_aggregations = @is_aggregations:: bool
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_aggregation_name:: bool THEN lower(agg.aggregation_name) LIKE '%' || lower(@aggregation_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_aggregation_name:: bool THEN lower(agg.aggregation_name) = lower(@aggregation_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_version:: bool THEN lower(p.product_version) LIKE '%' || lower(@product_version:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_version:: bool THEN lower(p.product_version) = lower(@product_version)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_profile_user:: bool THEN lower(pcu.profile_user) LIKE '%' || lower(@profile_user:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_profile_user:: bool THEN lower(pcu.profile_user) = lower(@profile_user)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_team:: bool THEN lower(pcu.team) LIKE '%' || lower(@team:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_team:: bool THEN lower(pcu.team) = lower(@team)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_number_of_users:: bool THEN lower(pcu.number_of_users) LIKE '%' || lower(@number_of_users:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_number_of_users:: bool THEN lower(pcu.number_of_users) = lower(@number_of_users)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_purchase_date:: bool THEN pcu.purchase_date <= @purchase_date
            ELSE TRUE
        END
    )
GROUP BY
    pcu.id,
    pcu.swidtag,
    pcu.purchase_date,
    pcu.number_of_users,
    pcu.profile_user,
    pcu.team,
    pcu.updated_on,
    pcu.created_on,
    pcu.created_by,
    pcu.updated_by,
    pcu.aggregation_id,
    agg.aggregation_name,
    p.product_version,
    p.product_name
ORDER BY
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @aggregation_name_asc:: bool THEN agg.aggregation_name
    END asc,
    CASE
        WHEN @aggregation_name_desc:: bool THEN agg.aggregation_name
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @profile_user_asc:: bool THEN pcu.profile_user
    END asc,
    CASE
        WHEN @profile_user_desc:: bool THEN pcu.profile_user
    END desc,
    CASE
        WHEN @team_asc:: bool THEN pcu.team
    END asc,
    CASE
        WHEN @team_desc:: bool THEN pcu.team
    END desc,
    CASE
        WHEN @number_of_users_asc:: bool THEN pcu.number_of_users
    END asc,
    CASE
        WHEN @number_of_users_desc:: bool THEN pcu.number_of_users
    END desc,
    CASE
        WHEN @purchase_date_asc:: bool THEN pcu.purchase_date
    END asc,
    CASE
        WHEN @purchase_date_desc:: bool THEN pcu.purchase_date
    END desc;

-- name: ExportNominativeUsersProducts :many

SELECT
    count(*) OVER() AS totalRecords,
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    p.product_version,
    p.product_name
FROM nominative_user nU
    INNER JOIN products p on nu.swidtag = p.swidtag and nu.scope = ANY(@scope:: TEXT []) AND p.scope = ANY(@scope:: TEXT [])
where
    nu.swidtag != ''
    AND nu.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @lk_product_name:: bool THEN lower(p.product_name) LIKE '%' || lower(@product_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_name:: bool THEN lower(p.product_name) = lower(@product_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_product_version:: bool THEN lower(p.product_version) LIKE '%' || lower(@product_version:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_product_version:: bool THEN lower(p.product_version) = lower(@product_version)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_name:: bool THEN lower(nu.user_name) LIKE '%' || lower(@user_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_name:: bool THEN lower(nu.user_name) = lower(@user_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_first_name:: bool THEN lower(nu.first_name) LIKE '%' || lower(@first_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_first_name:: bool THEN lower(nu.first_name) = lower(@first_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_email:: bool THEN lower(nu.user_email) LIKE '%' || lower(@user_email:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_email:: bool THEN lower(nu.user_email) = lower(@user_email)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_profile:: bool THEN lower(nu.profile) LIKE '%' || lower(@profile:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_profile:: bool THEN lower(nu.profile) = lower(@profile)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_activation_date:: bool THEN date(nu.activation_date):: text = @activation_date:: text
            ELSE TRUE
        END
    )
GROUP BY
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    p.product_version,
    p.product_name
ORDER BY
    CASE
        WHEN @product_name_asc:: bool THEN p.product_name
    END asc,
    CASE
        WHEN @product_name_desc:: bool THEN p.product_name
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @user_name_asc:: bool THEN nu.user_name
    END asc,
    CASE
        WHEN @user_name_desc:: bool THEN nu.user_name
    END desc,
    CASE
        WHEN @first_name_asc:: bool THEN nu.first_name
    END asc,
    CASE
        WHEN @first_name_desc:: bool THEN nu.first_name
    END desc,
    CASE
        WHEN @user_email_asc:: bool THEN nu.user_email
    END asc,
    CASE
        WHEN @user_email_desc:: bool THEN nu.user_email
    END desc,
    CASE
        WHEN @profile_asc:: bool THEN nu.profile
    END asc,
    CASE
        WHEN @profile_desc:: bool THEN nu.profile
    END desc,
    CASE
        WHEN @activation_date_asc:: bool THEN nu.activation_date
    END asc,
    CASE
        WHEN @activation_date_desc:: bool THEN nu.activation_date
    END desc;

-- name: ExportNominativeUsersAggregation :many

SELECT
    count(*) OVER() AS totalRecords,
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    nu.aggregations_id,
    agg.aggregation_name
FROM nominative_user nU
    LEFT JOIN aggregations agg on nu.aggregations_id = agg.id
WHERE
    nu.aggregations_id != 0
    AND nu.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @lk_aggregation_name:: bool THEN lower(agg.aggregation_name) LIKE '%' || lower(@aggregation_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_aggregation_name:: bool THEN lower(agg.aggregation_name) = lower(@aggregation_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_name:: bool THEN lower(nu.user_name) LIKE '%' || lower(@user_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_name:: bool THEN lower(nu.user_name) = lower(@user_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_first_name:: bool THEN lower(nu.first_name) LIKE '%' || lower(@first_name:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_first_name:: bool THEN lower(nu.first_name) = lower(@first_name)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_user_email:: bool THEN lower(nu.user_email) LIKE '%' || lower(@user_email:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_user_email:: bool THEN lower(nu.user_email) = lower(@user_email)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @lk_profile:: bool THEN lower(nu.profile) LIKE '%' || lower(@profile:: TEXT) || '%'
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_profile:: bool THEN lower(nu.profile) = lower(@profile)
            ELSE TRUE
        END
    )
    AND (
        CASE
            WHEN @is_activation_date:: bool THEN date(nu.activation_date):: text = @activation_date:: text
            ELSE TRUE
        END
    )
GROUP BY
    nu.user_id,
    nu.swidtag,
    nu.activation_date,
    nu.user_email,
    nu.user_name,
    nu.first_name,
    nu.profile,
    nu.product_editor,
    nu.updated_at,
    nu.created_at,
    nu.created_by,
    nu.updated_by,
    nu.aggregations_id,
    agg.aggregation_name
ORDER BY
    CASE
        WHEN @aggregation_name_asc:: bool THEN agg.aggregation_name
    END asc,
    CASE
        WHEN @aggregation_name_desc:: bool THEN agg.aggregation_name
    END desc,
    CASE
        WHEN @user_name_asc:: bool THEN nu.user_name
    END asc,
    CASE
        WHEN @user_name_desc:: bool THEN nu.user_name
    END desc,
    CASE
        WHEN @first_name_asc:: bool THEN nu.first_name
    END asc,
    CASE
        WHEN @first_name_desc:: bool THEN nu.first_name
    END desc,
    CASE
        WHEN @user_email_asc:: bool THEN nu.user_email
    END asc,
    CASE
        WHEN @user_email_desc:: bool THEN nu.user_email
    END desc,
    CASE
        WHEN @profile_asc:: bool THEN nu.profile
    END asc,
    CASE
        WHEN @profile_desc:: bool THEN nu.profile
    END desc,
    CASE
        WHEN @activation_date_asc:: bool THEN nu.activation_date
    END asc,
    CASE
        WHEN @activation_date_desc:: bool THEN nu.activation_date
    END desc;



-- name: ListUnderusageByEditor :many
SELECT 
    count(*) OVER() AS totalRecords,
    (delta_number):: Numeric(15, 2) as delta,
    metrics,
    product_names,
    aggregation_name,
    scope  FROM
    overall_computed_licences
WHERE
    scope = ANY(@scope:: TEXT [])
    AND cost_optimization = FALSE
    AND metic_not_defined = FALSE
    AND (
        CASE
            WHEN @lk_editor:: bool THEN lower(editor) = lower(@editor)
            ELSE TRUE
        END 
    )
    AND (
        CASE
            WHEN @lk_product_names:: bool THEN lower(product_names) = lower(@product_names)
            ELSE TRUE
        END 
    )
GROUP BY
    overall_computed_licences.scope,
    overall_computed_licences.metrics,
    overall_computed_licences.delta_number,
    overall_computed_licences.product_names,
    overall_computed_licences.aggregation_name
HAVING (overall_computed_licences.delta_number) > 0
ORDER BY
    CASE
        WHEN @scope_asc:: bool THEN overall_computed_licences.scope
    END asc,
    CASE
        WHEN @scope_desc:: bool THEN overall_computed_licences.scope
    END desc,
    CASE
        WHEN @metrics_asc:: bool THEN overall_computed_licences.metrics
    END asc,
    CASE
        WHEN @metrics_desc:: bool THEN overall_computed_licences.metrics
    END desc,
    CASE
        WHEN @delta_number_asc:: bool THEN overall_computed_licences.delta_number
    END asc,
    CASE
        WHEN @delta_number_desc:: bool THEN overall_computed_licences.delta_number
    END desc;


-- name: InsertNominativeUserFileUploadDetails :exec

INSERT INTO
    nominative_user_file_uploaded_details (
        upload_id,
        scope,
        swidtag,
        aggregations_id,
        product_editor,
        uploaded_by,
        nominative_users_details,
        record_succeed,
        record_failed,
        file_name,
        sheet_name,
        file_status
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    );

-- name: ListNominativeUsersUploadedFiles :many

SELECT
   	count(*) OVER() AS totalRecords,
    fd.id,
    upload_id,
    fd.scope,
    fd.swidtag,
    aggregations_id,
    fd.product_editor,
    uploaded_by,
    uploaded_at,
    nominative_users_details,
    record_succeed,
    record_failed,
    file_name,
    sheet_name,
    file_status,
        CASE 
   	     WHEN aggregations_id IS NOT NULL THEN a.aggregation_name
        ELSE  p.product_name
  	    END AS pname,
        CASE 
   	    WHEN aggregations_id IS NOT NULL THEN 'aggregation'
        ELSE  'individual'
  	    END AS nametype,
    p.product_name,
    p.product_version,
    a.aggregation_name
FROM
    nominative_user_file_uploaded_details fd
    left join products p on fd.swidtag=p.swidtag and fd.scope=p.scope
    left join aggregations a on fd.aggregations_id=a.id and fd.scope=a.scope
where
    fd.scope = ANY(@scope:: TEXT [])
    AND (
        CASE
            WHEN @file_upload_id:: bool THEN fd.id = @id
            ELSE TRUE
        END
    )
    ORDER BY
    CASE
        WHEN @file_name_asc:: bool THEN file_name
    END asc,
    CASE
        WHEN @file_name_desc:: bool THEN file_name
    END desc,
    CASE
        WHEN @file_status_asc:: bool THEN file_status
    END asc,
    CASE
        WHEN @file_status_desc:: bool THEN file_status
    END desc,
    CASE
        WHEN @product_editor_asc:: bool THEN fd.product_editor
    END asc,
    CASE
        WHEN @product_editor_desc:: bool THEN fd.product_editor
    END desc,
    CASE
        WHEN @name_asc:: bool THEN  CASE 
                WHEN aggregations_id IS NOT NULL THEN a.aggregation_name
                ELSE  p.product_name
            END
    END asc,
    CASE
        WHEN @name_desc:: bool THEN  CASE 
                WHEN aggregations_id IS NOT NULL THEN a.aggregation_name
                ELSE  p.product_name
            END
    END desc,
    CASE
        WHEN @product_version_asc:: bool THEN p.product_version
    END asc,
    CASE
        WHEN @product_version_desc:: bool THEN p.product_version
    END desc,
    CASE
        WHEN @uploaded_by_asc:: bool THEN uploaded_by
    END asc,
    CASE
        WHEN @uploaded_by_desc:: bool THEN uploaded_by
    END desc,
    CASE
        WHEN @uploaded_on_asc:: bool THEN uploaded_at
    END asc,
    CASE
        WHEN @uploaded_on_desc:: bool THEN uploaded_at
    END desc,
    CASE
        WHEN @productType_asc:: bool THEN CASE 
   	    WHEN aggregations_id IS NOT NULL THEN 'aggregation'
        ELSE  'individual'
  	    END
    END asc,
    CASE
        WHEN @productType_desc:: bool THEN CASE 
   	    WHEN aggregations_id IS NOT NULL THEN 'aggregation'
        ELSE  'individual'
  	    END
    END desc
    LIMIT @page_size
    OFFSET @page_num;

-- name: GetEditorExpensesByScopeData :many

SELECT
    editor,
    coalesce(SUM(total_purchase_cost), 0.0):: FLOAT as total_purchase_cost,
    coalesce(
        SUM(total_maintenance_cost),
        0.0
    ):: FLOAT as total_maintenance_cost,
    coalesce(SUM(total_cost), 0.0):: FLOAT as total_cost
FROM (
        SELECT
            product_editor as editor,
            total_purchase_cost,
            total_maintenance_cost,
            total_cost
        FROM acqrights
        WHERE
            acqrights.scope = ANY(@scope:: TEXT [])
        GROUP BY
            editor,
            total_purchase_cost,
            total_maintenance_cost,
            total_cost
        UNION
        SELECT
            a.product_editor as editor,
            ar.total_purchase_cost,
            ar.total_maintenance_cost,
            ar.total_cost
        FROM aggregations as a
            INNER JOIN aggregated_rights as ar ON a.id = ar.aggregation_id
        WHERE
            a.scope = ANY(@scope:: TEXT [])
        GROUP BY
            a.product_editor,
            ar.total_purchase_cost,
            ar.total_maintenance_cost,
            ar.total_cost
    ) as editorExpenseByScopeData
GROUP BY editor;

-- name: GetConcurrentNominativeUsersBySwidTag :many

SELECT scope, swidtag
FROM product_concurrent_user
WHERE
    product_concurrent_user.swidtag = ANY(@swidtag:: TEXT [])
    AND product_concurrent_user.scope = ANY(@scope:: TEXT [])
UNION
SELECT scope, swidtag
FROM nominative_user
WHERE
    nominative_user.swidtag = ANY(@swidtag:: TEXT [])
    AND nominative_user.scope = ANY(@scope:: TEXT []);

-- name: DeleteProductsBySwidTagScope :exec

DELETE FROM products
WHERE
    products.scope = @scope
    AND products.swidtag = @swidtag
    AND product_type = 'SAAS';
    
-- name: GetEditor :many
SELECT name from editor_catalog;

-- name: GetProductByNameEditor :many
SELECT * from products 
    WHERE 
    product_name=ANY(@product_name:: TEXT [])
    AND product_editor=ANY(@product_editor:: TEXT []);
