CREATE TABLE auth.users
(
    id                   VARCHAR(21)  NOT NULL UNIQUE,
    username             VARCHAR(255) NULL UNIQUE,
    email                VARCHAR(255) NULL UNIQUE,
    encrypted_password   VARCHAR(255) NULL,
    confirmed_at         timestamptz  NULL,
    invited_at           timestamptz  NULL,
    confirmation_token   VARCHAR(255) NULL,
    confirmation_sent_at timestamptz  NULL,
    recovery_token       VARCHAR(255) NULL,
    recovery_sent_at     timestamptz  NULL,
    email_change_token   VARCHAR(255) NULL,
    email_change         VARCHAR(255) NULL,
    email_change_sent_at timestamptz  NULL,
    last_sign_in_at      timestamptz  NULL,
    raw_app_meta_data    jsonb        NULL,
    raw_user_meta_data   jsonb        NULL,
    CONSTRAINT pk_auth_user PRIMARY KEY (id)
);

-- CREATE INDEX users_instance_id_email_idx ON auth.user USING btree (instance_id, email);
-- CREATE INDEX users_instance_id_idx ON auth.user USING btree (instance_id);
COMMENT ON TABLE auth.users IS 'Auth: Stores user login data within a secure schema.';

CREATE TABLE auth.identities
(
    id              VARCHAR(21) NOT NULL DEFAULT nanoid() UNIQUE,
    provider_id     TEXT        NOT NULL,
    user_id         VARCHAR(21) NOT NULL,
    identity_data   jsonb       NOT NULL,
    provider        TEXT        NOT NULL,
    last_sign_in_at timestamptz,
    created_at      timestamptz,
    updated_at      timestamptz,
    email           TEXT GENERATED ALWAYS AS (LOWER((identity_data ->> 'email'::TEXT))) STORED,
    CONSTRAINT pk_auth_idp PRIMARY KEY (id),
    CONSTRAINT fk_auth_idp_user_id FOREIGN KEY (user_id) REFERENCES auth.users ON DELETE CASCADE,
    CONSTRAINT uq_auth_idp_id UNIQUE (provider_id, provider)
);
COMMENT ON TABLE auth.identities IS 'Auth: Stores identities associated to a user.';

CREATE TYPE auth.aal_level AS ENUM ('aal1', 'aal2', 'aal3');
COMMENT ON TYPE auth.aal_level IS 'Auth: The level of assurance for a user session.';

CREATE TABLE auth.sessions
(
    id           VARCHAR(21) NOT NULL DEFAULT nanoid() UNIQUE,
    user_id      VARCHAR(21) NOT NULL,
    created_at   timestamptz,
    updated_at   timestamptz,
    factor_id    uuid,
    aal          auth.aal_level,
    expires_at   timestamptz,
    refreshed_at TIMESTAMP,
    user_agent   TEXT,
    ip           inet,
    tag          TEXT,
    CONSTRAINT pk_auth_sessions PRIMARY KEY (id),
    CONSTRAINT fk_auth_sessions_user_id FOREIGN KEY (user_id) REFERENCES auth.users ON DELETE CASCADE
);
COMMENT
    ON TABLE auth.sessions IS 'Auth: Stores session data associated to a user.';
COMMENT
    ON COLUMN auth.sessions.expires_at IS 'Auth: Expires at is a nullable column that contains a timestamp after which the session should be regarded as expired.';

CREATE INDEX session_not_after_idx
    ON auth.sessions (expires_at DESC);

CREATE INDEX session_user_id_idx
    ON auth.sessions (user_id);

CREATE INDEX user_id_created_at_idx
    ON auth.sessions (user_id, created_at);


CREATE TABLE auth.audit_log_entries
(
--     instance_id uuid,
    id         VARCHAR(21) NOT NULL DEFAULT nanoid() UNIQUE,
    payload    json,
    created_at TIMESTAMP WITH TIME ZONE,
    ip_address VARCHAR(64)          DEFAULT ''::CHARACTER VARYING NOT NULL,
    CONSTRAINT pk_auth_audit_log_entries PRIMARY KEY (id)
);

COMMENT
    ON TABLE auth.audit_log_entries IS 'Auth: Audit trail for user actions.';

CREATE TABLE auth.roles
(
    id          VARCHAR(21)  NOT NULL DEFAULT nanoid() UNIQUE,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  timestamptz           DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamptz           DEFAULT CURRENT_TIMESTAMP,
    deleted_at  timestamptz,
    is_enabled  BOOLEAN               DEFAULT TRUE,
    CONSTRAINT pk_auth_role PRIMARY KEY (id),
    CONSTRAINT uq_auth_role_name UNIQUE (name)
);

CREATE TABLE auth.user_roles
(
    user_id    VARCHAR(21) NOT NULL,
    role_id    VARCHAR(21) NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_auth_user_role PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_auth_user_role_user_id FOREIGN KEY (user_id) REFERENCES auth.users ON DELETE CASCADE,
    CONSTRAINT fk_auth_user_role_role_id FOREIGN KEY (role_id) REFERENCES auth.roles ON DELETE CASCADE
);

-- Auth tuples
INSERT INTO auth.roles (name, description)
VALUES ('admin', 'Admin role'),
       ('user', 'User role'),
       ('guest', 'Guest role');

CREATE OR REPLACE FUNCTION auth.fn_is_role_active_by_name(TEXT)
    RETURNS BOOLEAN AS
$$
BEGIN
    RETURN EXISTS(SELECT 1
                  FROM auth.roles r
                  WHERE r.name = $1
                    AND r.deleted_at IS NULL
                    AND r.is_enabled);
END;
$$
    LANGUAGE plpgsql
    STABLE;

CREATE OR REPLACE FUNCTION auth.fn_is_role_active_by_id(TEXT)
    RETURNS BOOLEAN AS
$$
BEGIN
    RETURN EXISTS(SELECT 1
                  FROM auth.roles r
                  WHERE r.id = $1
                    AND r.deleted_at IS NULL
                    AND r.is_enabled);
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION auth.fn_user_has_role(
    user_id TEXT,
    role_name TEXT
) RETURNS BOOLEAN AS
$$
DECLARE
    result    BOOLEAN;
    v_role_id TEXT;
BEGIN
    IF user_id IS NULL THEN
        RAISE EXCEPTION 'user_id cannot be null'
            USING ERRCODE = '22000';
    END IF;

    SELECT id
    FROM auth.roles
    WHERE name = role_name
    INTO v_role_id;
    IF NOT found THEN
        RAISE EXCEPTION 'Role % not found', role_name
            USING ERRCODE = '22000';
    END IF;

    SELECT EXISTS(SELECT 1
                  FROM auth.user_roles ur
                  WHERE ur.user_id = fn_user_has_role.user_id
                    AND ur.role_id = v_role_id)
    INTO result;

    RETURN result;
END;
$$ LANGUAGE plpgsql STABLE;
