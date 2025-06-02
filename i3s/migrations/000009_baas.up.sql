-- CREATE TABLE IF NOT EXISTS dbo.objects
-- (
--     id                  VARCHAR(21) NOT NULL DEFAULT nanoid() UNIQUE,
--     entity_id           VARCHAR(21),
--     chinese_name        VARCHAR(512),
--     chinese_description VARCHAR(4000),
--     english_name        VARCHAR(512),
--     english_description VARCHAR(4000),
--     created_at          timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at          timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     deleted_at          timestamptz,
--     owner_id            VARCHAR(21),
--     click_count         INT                  DEFAULT 0 NOT NULL,
--     outlink_count       INT,
--     inlink_count        INT,
--     is_hidden           BOOLEAN              DEFAULT FALSE NOT NULL,
--     CONSTRAINT pk_dbo_objects PRIMARY KEY (id),
--     CONSTRAINT fk_dbo_objects_owner_id FOREIGN KEY (owner_id) REFERENCES auth.users
-- );
CREATE TABLE dbo.projects
(
    id                VARCHAR(21)  NOT NULL UNIQUE,
    reference         VARCHAR(20)  NOT NULL UNIQUE,
    CONSTRAINT pk_dbo_projects PRIMARY KEY (id),
    CONSTRAINT fk_dbo_projects_id FOREIGN KEY (id) REFERENCES dbo.objects
)