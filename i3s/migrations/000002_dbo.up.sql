CREATE TABLE IF NOT EXISTS dbo.objects
(
    oid                 UUID        DEFAULT uuid_generate_v4() NOT NULL,
    entity_id           SMALLINT,
    chinese_name        VARCHAR(512),
    chinese_description VARCHAR(4000),
    english_name        VARCHAR(512),
    english_description VARCHAR(4000),
    created_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP  NOT NULL,
    updated_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP  NOT NULL,
    deleted_at          TIMESTAMPTZ,
    owner_id            UUID,
    click_count         INT         DEFAULT 0                  NOT NULL,
    outlink_count       INT,
    inlink_count        INT,
    is_hidden           BOOLEAN     DEFAULT FALSE              NOT NULL,
    CONSTRAINT pk_objects PRIMARY KEY (oid),
    CONSTRAINT fk_objects_owner_uid FOREIGN KEY (owner_id) REFERENCES auth.users
);

CREATE INDEX IF NOT EXISTS idx_object_chinese_name ON dbo.objects (chinese_name);

CREATE TABLE IF NOT EXISTS dbo.object_relations
(
    first_oid   UUID                                  NOT NULL,
    second_oid  UUID                                  NOT NULL,
    rank        INT,
    description VARCHAR(1000),
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT pk_object_relation PRIMARY KEY (first_oid, second_oid),
    CONSTRAINT fk_object_relation_first_oid FOREIGN KEY (first_oid) REFERENCES dbo.objects (oid),
    CONSTRAINT fk_object_relation_second_oid FOREIGN KEY (second_oid) REFERENCES dbo.objects (oid)
);

CREATE TABLE IF NOT EXISTS dbo.entities
(
    eid           INT GENERATED ALWAYS AS IDENTITY,
    chinese_name  VARCHAR(50),
    english_name  VARCHAR(50),
    is_relational BOOLEAN DEFAULT FALSE NOT NULL,
    is_hideable   BOOLEAN DEFAULT FALSE NOT NULL,
    is_deletable  BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT pk_entity PRIMARY KEY (eid)
);

CREATE TABLE IF NOT EXISTS dbo.classes
(
    cid                 INT GENERATED ALWAYS AS IDENTITY,
    entity_id           INT,
    chinese_name        VARCHAR(256),
    chinese_description VARCHAR(4000),
    english_name        VARCHAR(256),
    english_description VARCHAR(4000),
    id_path             VARCHAR(1000) CONSTRAINT uq_class_id_path UNIQUE,
    name_path           VARCHAR(1000) CONSTRAINT uq_class_name_path UNIQUE,
    created_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at          TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMPTZ,
    object_count        INT         DEFAULT 0                 NOT NULL,
    class_rank          SMALLINT    DEFAULT 0                 NOT NULL,
    object_rank         SMALLINT    DEFAULT 0                 NOT NULL,
    hierarchy_level     SMALLINT,
    click_count         INT         DEFAULT 0                 NOT NULL,
    keywords            TEXT[]      DEFAULT '{}',
    owner_id            UUID        DEFAULT NULL,
    is_hidden           BOOLEAN     DEFAULT FALSE             NOT NULL,
    is_child            BOOLEAN     DEFAULT FALSE             NOT NULL,
    CONSTRAINT pk_class PRIMARY KEY (cid),
    CONSTRAINT fk_class_entities FOREIGN KEY (entity_id) REFERENCES dbo.entities (eid),
    CONSTRAINT fk_class_owner_id FOREIGN KEY (owner_id) REFERENCES auth.users
);

CREATE TABLE IF NOT EXISTS dbo.co
(
    cid              INT                                   NOT NULL,
    oid              UUID                                  NOT NULL,
    rank             INT,
    membership_grade INT,
    description      VARCHAR(1000),
    created_at       TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT pk_co PRIMARY KEY (cid, oid),
    CONSTRAINT fk_co_cid FOREIGN KEY (cid) REFERENCES dbo.classes (cid),
    CONSTRAINT fk_co_oid FOREIGN KEY (oid) REFERENCES dbo.objects (oid)
);

CREATE TABLE IF NOT EXISTS dbo.inheritances
(
    pcid             INT NOT NULL,
    ccid             INT NOT NULL,
    rank             SMALLINT,
    membership_grade INT,
    CONSTRAINT pk_inheritance PRIMARY KEY (pcid, ccid),
    CONSTRAINT fk_inheritance_pcid FOREIGN KEY (pcid) REFERENCES dbo.classes (cid),
    CONSTRAINT fk_inheritance_ccid FOREIGN KEY (ccid) REFERENCES dbo.classes (cid)
);

CREATE TABLE IF NOT EXISTS dbo.permissions
(
    class_id        INT                NOT NULL,
    role_type       BOOLEAN            NOT NULL,
    role_id         UUID               NOT NULL,
    permission_bits SMALLINT DEFAULT 1 NOT NULL,
    CONSTRAINT uq_permissions UNIQUE (class_id, role_type, role_id)
);

COMMENT ON COLUMN dbo.permissions.role_type IS '0表示是群組，1表示是使用者';
COMMENT ON COLUMN dbo.permissions.role_id IS '由RoleType決定值為Auth.Groups.GID或Auth.User.UID';
