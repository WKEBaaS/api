CREATE TABLE auth.users
(
--     instance_id          uuid         NULL,
    id                   uuid         NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    aud                  VARCHAR(255) NULL,
    "role"               VARCHAR(255) NULL,
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
    created_at           timestamptz           DEFAULT CURRENT_TIMESTAMP,
    updated_at           timestamptz           DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_auth_user PRIMARY KEY (id)
);

-- CREATE INDEX users_instance_id_email_idx ON auth.user USING btree (instance_id, email);
-- CREATE INDEX users_instance_id_idx ON auth.user USING btree (instance_id);
COMMENT ON TABLE auth.users IS 'Auth: Stores user login data within a secure schema.';

CREATE TABLE auth.identities
(
    id              uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    provider_id     TEXT                            NOT NULL,
    user_id         uuid                            NOT NULL
        REFERENCES auth.users
            ON DELETE CASCADE,
    identity_data   jsonb                           NOT NULL,
    provider        TEXT                            NOT NULL,
    last_sign_in_at timestamptz,
    created_at      timestamptz,
    updated_at      timestamptz,
    email           TEXT GENERATED ALWAYS AS (LOWER((identity_data ->> 'email'::TEXT))) STORED,
    CONSTRAINT uq_identities_provider_id_provider
        UNIQUE (provider_id, provider)
);
COMMENT ON TABLE auth.identities IS 'Auth: Stores identities associated to a user.';

CREATE TYPE auth.aal_level AS ENUM ('aal1', 'aal2', 'aal3');
COMMENT ON TYPE auth.aal_level IS 'Auth: The level of assurance for a user session.';

CREATE TABLE auth.sessions
(
    id           uuid NOT NULL PRIMARY KEY,
    user_id      uuid NOT NULL
        REFERENCES auth.users
            ON DELETE CASCADE,
    created_at   timestamptz,
    updated_at   timestamptz,
    factor_id    uuid,
    aal          auth.aal_level,
    expires_at   timestamptz,
    refreshed_at TIMESTAMP,
    user_agent   TEXT,
    ip           inet,
    tag          TEXT
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
    id         uuid                                      NOT NULL,
    payload    json,
    created_at TIMESTAMP WITH TIME ZONE,
    ip_address VARCHAR(64) DEFAULT ''::CHARACTER VARYING NOT NULL,
    CONSTRAINT pk_auth_audit_log_entries PRIMARY KEY (id)
);

COMMENT
    ON TABLE auth.audit_log_entries IS 'Auth: Audit trail for user actions.';

CREATE TABLE auth.roles
(
    id          uuid         NOT NULL DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  timestamptz           DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamptz           DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_auth_role PRIMARY KEY (id),
    CONSTRAINT uq_auth_role_name UNIQUE (name)
);

CREATE TABLE auth.user_roles
(
    user_id    uuid NOT NULL
        REFERENCES auth.users
            ON DELETE CASCADE,
    role_id    uuid NOT NULL
        REFERENCES auth.roles
            ON DELETE CASCADE,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_auth_user_role PRIMARY KEY (user_id, role_id)
);

-- Gets the User ID from the request cookie
CREATE
    OR REPLACE FUNCTION auth.uid() RETURNS uuid AS
$$
SELECT NULLIF(CURRENT_SETTING('request.jwt.claims.sub', TRUE), '') ::uuid;
$$
    LANGUAGE sql STABLE;

-- Gets the User ID from the request cookie
CREATE
    OR REPLACE FUNCTION auth.role() RETURNS TEXT AS
$$
SELECT NULLIF(CURRENT_SETTING('request.jwt.claims.role', TRUE), '') ::TEXT;
$$
    LANGUAGE sql STABLE;
COMMENT
    ON FUNCTION auth.role() IS 'Auth: Returns the role of the current user.';

-- Gets the User email
CREATE
    OR REPLACE FUNCTION auth.email() RETURNS TEXT AS
$$
SELECT NULLIF(CURRENT_SETTING('request.jwt.claims.email', TRUE), '') ::TEXT;
$$
    LANGUAGE sql STABLE;
COMMENT
    ON FUNCTION auth.email() IS 'Auth: Returns the email of the current user.';
