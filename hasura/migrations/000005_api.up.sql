-- 創建或修改函數來生成視圖參數
CREATE OR REPLACE FUNCTION api.fn_gen_view_params(view_name TEXT, schema_name TEXT DEFAULT 'public')
    RETURNS TEXT
    LANGUAGE plpgsql
AS
$$
DECLARE
    view_params TEXT;
BEGIN
    SELECT json_agg(json_build_object('name', a.attname, 'type', format_type(a.atttypid, a.atttypmod)))::TEXT
    INTO view_params
    FROM pg_attribute a
             JOIN pg_class c ON a.attrelid = c.oid
             JOIN pg_namespace n ON c.relnamespace = n.oid
    WHERE c.relname = view_name
      AND n.nspname = schema_name
      AND a.attnum > 0 -- 過濾掉系統列
      AND NOT a.attisdropped; -- 排除已刪除的列

    RETURN LOWER(view_params);
END;
$$;

-- 創建 UUID To View 表
CREATE TABLE api.uuid_to_view
(
    uuid       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    view_name   TEXT        NOT NULL,
    view_schema TEXT        NOT NULL DEFAULT 'public',
    auth       VARCHAR(10) NOT NULL DEFAULT 'Authenticated',
    check_class BOOLEAN     NOT NULL DEFAULT FALSE,
    view_params TEXT        NOT NULL DEFAULT ''
);

-- 添加擴展屬性
COMMENT ON COLUMN api.uuid_to_view.auth IS 'Admin|Authenticated|Public';

-- 創建或修改觸發器
CREATE OR REPLACE FUNCTION api.tr_uuid2view()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS
$$
BEGIN
    UPDATE api.uuid_to_view
    SET view_params = api.fn_gen_view_params(NEW.viewname, NEW.viewschema)
    WHERE uuid = NEW.uuid;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION api.tr_uuid_to_view()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS
$$
BEGIN
    IF NEW.view_name <> OLD.view_name OR NEW.schema_name <> OLD.schema_name THEN
        UPDATE api.uuid_to_view
        SET view_params = api.fn_gen_view_params(NEW.view_name, NEW.schema_name)
        WHERE uuid = NEW.uuid;
    END IF;
    RETURN NEW;
END;
$$;

-- 創建 vd_TableColumns 視圖
CREATE OR replace view api.vd_table_columns AS
SELECT t.tablename                                                                    AS table_name,
       c.column_name                                                                  AS column_name,
       t.schemaname                                                                   AS table_schema,
       col_description(
               (t.schemaname || '.' || t.tablename)::regclass, ordinal_position::int) AS description
FROM pg_tables t
         JOIN information_schema.columns c ON t.tablename = c.table_name
    AND t.schemaname = c.table_schema;

-- 創建 vd_ViewColumns 視圖
CREATE OR replace view api.vd_view_columns AS
SELECT v.viewname                                                                    AS table_name,
       c.column_name                                                                 AS column_name,
       v.schemaname                                                                  AS table_schema,
       col_description(
               (v.schemaname || '.' || v.viewname)::regclass, ordinal_position::int) AS description
FROM pg_views v
         JOIN information_schema.columns c ON v.viewname = c.table_name
    AND v.schemaname = c.table_schema;

-- 插入資料到 UUID To View
INSERT INTO api.uuid_to_view (view_name, auth)
VALUES ('vd_table_columns', 'Admin'),
       ('vd_view_columns', 'Admin');
