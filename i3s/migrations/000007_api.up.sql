CREATE TABLE IF NOT EXISTS api.class_permission_enum
(
    id  TEXT,
    bit SMALLINT NOT NULL,
    CONSTRAINT pk_class_permission_enum PRIMARY KEY (id),
    CONSTRAINT uq_class_permission_enum UNIQUE (bit)
);

INSERT INTO api.class_permission_enum(id, bit)
VALUES ('read-class', 1),
       ('read-object', 2),
       ('insert', 4),
       ('delete', 8),
       ('update', 16),
       ('modify', 32),
       ('subscribe', 64);

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
    -- Check if the user is an admin
    RETURN QUERY SELECT TRUE FROM api.has_role(hasura_session, 'admin');
    IF found THEN
        RETURN;
    END IF;

    SELECT bit
    INTO v_permission_bit
    FROM api.class_permission_enum
    WHERE id = permission;

    IF NOT found THEN
        RAISE EXCEPTION 'Permission % not found', permission
            USING ERRCODE = 'P0002';
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
                          JOIN auth.user_groups ug ON ug.user_id = v_user_id
                          JOIN auth.groups g ON g.id = ug.group_id
                 WHERE p.class_id = check_class_permission.class_id
                   AND p.role_type = FALSE
                   AND p.role_id = g.id
                   AND g.is_enabled
                   AND (p.permission_bits & v_permission_bit) > 0
                 LIMIT 1;

    RETURN;
END;
$$
    LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION api.check_class_permission IS 'Returns true if the user has permission. Otherwise, returns null.';

CREATE OR REPLACE FUNCTION api.insert_class(
    hasura_session json,
    parent_class_id VARCHAR(21),
    entity_id VARCHAR(21),
    chinese_name TEXT,
    chinese_description TEXT,
    english_name TEXT,
    english_description TEXT,
    owner_id VARCHAR(21) DEFAULT NULL
) RETURNS SETOF dbo.classes AS
$$
DECLARE
    has_permission BOOLEAN;
BEGIN
    SELECT result INTO has_permission FROM api.check_class_permission(hasura_session, parent_class_id, 'insert');
    IF NOT found THEN
        RAISE EXCEPTION 'No % Permission', 'insert'
            USING ERRCODE = 'P0002';
    END IF;

END;
$$
    LANGUAGE plpgsql;