CREATE TABLE IF NOT EXISTS api.has_role_result
(
    has_role BOOLEAN
);

CREATE OR REPLACE FUNCTION api.has_role(
    hasura_session json,
    role TEXT
) RETURNS SETOF api.has_role_result AS
$$
DECLARE
    v_user_id VARCHAR(21) := hasura_session ->> 'x-hasura-user-id';
BEGIN
    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'x-hasura-user-id is null' USING ERRCODE = '22000';
    END IF;

    RETURN QUERY SELECT TRUE
                 FROM auth.user_roles ur
                          JOIN auth.roles r ON r.id = ur.role_id
                 WHERE ur.user_id = v_user_id
                   AND r.name = role
                 LIMIT 1;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION api.has_role IS 'Returns true if the user has role. Otherwise, returns null.';

CREATE TABLE IF NOT EXISTS api.check_class_permission_result
(
    result BOOLEAN
);

CREATE OR REPLACE FUNCTION api.check_class_permission(
    hasura_session json,
    class_id VARCHAR(21),
    permission TEXT
) RETURNS SETOF api.check_class_permission_result AS
$$
DECLARE
    v_user_id        VARCHAR(21) := hasura_session ->> 'x-hasura-user-id';
    v_permission_bit SMALLINT;
BEGIN
    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'x-hasura-user-id is null' USING ERRCODE = '22000';
    END IF;

    IF hasura_session ->> 'x-hasura-role' = 'admin' THEN
        RETURN QUERY SELECT TRUE;
        RETURN;
    END IF;

    -- Check if the user is an admin
    RETURN QUERY SELECT TRUE FROM api.has_role(hasura_session, 'admin');
    IF found THEN
        RETURN;
    END IF;

    SELECT bit
    INTO v_permission_bit
    FROM dbo.permission_enum
    WHERE id = permission;

    IF NOT found THEN
        RAISE EXCEPTION 'Permission % not found', permission
            USING ERRCODE = '22000';
    END IF;

    RETURN QUERY SELECT TRUE
                 FROM dbo.permissions p
                 WHERE p.class_id = check_class_permission.class_id
                   AND p.role_type = TRUE
                   AND p.role_id = v_user_id
                   AND (p.permission_bits & v_permission_bit) > 0
                 LIMIT 1;
    IF found THEN
        RETURN;
    END IF;

    -- Ref: https://www.postgresql.org/docs/current/plpgsql-control-structures.html
    -- Check user permission first, if not found, check group permission
    RETURN QUERY SELECT TRUE
                 FROM dbo.permissions p
                          JOIN auth.user_roles ur ON ur.user_id = v_user_id
                          JOIN auth.roles r ON r.id = ur.role_id
                 WHERE p.class_id = check_class_permission.class_id
                   AND p.role_type = FALSE
                   AND p.role_id = r.id
                   AND r.is_enabled
                   AND (p.permission_bits & v_permission_bit) > 0
                 LIMIT 1;

    RETURN;
END;
$$
    LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION api.check_class_permission IS 'Returns true if the user has permission. Otherwise, returns null.';

CREATE TABLE IF NOT EXISTS api.class_result
(
    LIKE dbo.classes
);

CREATE OR REPLACE FUNCTION api.insert_class(
    hasura_session json,
    parent_class_id VARCHAR(21),
    chinese_name VARCHAR(255),
    entity_id VARCHAR(21) DEFAULT NULL,
    chinese_description TEXT DEFAULT NULL,
    english_name VARCHAR(255) DEFAULT NULL,
    english_description TEXT DEFAULT NULL,
    owner_id VARCHAR(21) DEFAULT NULL
) RETURNS SETOF api.class_result AS
$$
DECLARE
    has_permission BOOLEAN;
BEGIN
    SELECT result INTO has_permission FROM api.check_class_permission(hasura_session, parent_class_id, 'insert');
    IF NOT found THEN
        RAISE EXCEPTION 'No % Permission', 'insert'
            USING ERRCODE = '22000';
    END IF;

    RETURN QUERY SELECT *
                 FROM dbo.fn_insert_class(
                         insert_class.parent_class_id,
                         insert_class.entity_id,
                         insert_class.chinese_name,
                         insert_class.chinese_description,
                         insert_class.english_name,
                         insert_class.english_description,
                         insert_class.owner_id
                      );
END;
$$
    LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE FUNCTION api.delete_class(
    hasura_session json,
    class_id VARCHAR(21)
) RETURNS SETOF api.class_result AS
$$
DECLARE
    has_permission BOOLEAN;
BEGIN
    SELECT result INTO has_permission FROM api.check_class_permission(hasura_session, class_id, 'delete');
    IF NOT found THEN
        RAISE EXCEPTION 'No % Permission', 'delete'
            USING ERRCODE = '22000';
    END IF;

    RETURN QUERY SELECT * FROM dbo.fn_delete_class(class_id);
END;
$$ LANGUAGE plpgsql VOLATILE;