CREATE OR REPLACE FUNCTION dbo.fn_gen_name_path(
    parent_class_id VARCHAR(21),
    chinese_name VARCHAR(255)
)
    RETURNS TEXT
AS
$$
DECLARE
    result TEXT;
BEGIN
    IF parent_class_id IS NULL THEN
        RETURN chinese_name;
    END IF;

    SELECT CASE
               WHEN c.name_path = '/' THEN '/' || fn_gen_name_path.chinese_name
               ELSE c.name_path || '/' || fn_gen_name_path.chinese_name
               END
    INTO result
    FROM dbo.classes c
    WHERE c.id = parent_class_id;

    RETURN result;
END;
$$
    LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbo.fn_gen_id_path(_parent_class_id VARCHAR(21), _class_id VARCHAR(21) DEFAULT NULL)
    RETURNS TEXT
AS
$$
DECLARE
    result TEXT;
BEGIN
    IF _parent_class_id IS NULL THEN
        RETURN _class_id;
    END IF;

    SELECT c.id_path || '/' || fn_gen_id_path._class_id
    INTO result
    FROM dbo.classes c
    WHERE c.id = _parent_class_id;

    RETURN result;
END;
$$
    LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbo.fn_insert_class(
    parent_class_id VARCHAR(21),
    entity_id VARCHAR(21),
    chinese_name VARCHAR(256),
    chinese_description VARCHAR(4000),
    english_name VARCHAR(256),
    english_description VARCHAR(4000),
    owner_id VARCHAR(21) DEFAULT NULL)
    RETURNS SETOF dbo.classes
AS
$$
DECLARE
    new_class_id        VARCHAR(21) := nanoid();
    new_name_path       TEXT        := dbo.fn_gen_name_path(parent_class_id, fn_insert_class.chinese_name);
    new_id_path         TEXT        := dbo.fn_gen_id_path(parent_class_id, new_class_id);
    new_hierarchy_level INT;
BEGIN
    -- 檢查NamePath是否重複
    IF EXISTS(SELECT 1
              FROM dbo.classes
              WHERE name_path = new_name_path) THEN
        RAISE EXCEPTION 'Error: name_path 已經存在，無法建立 class。NamePath: %', new_name_path USING ERRCODE = '22000';
    END IF;

    IF parent_class_id IS NULL THEN
        new_hierarchy_level := 0;
    ELSE
        SELECT hierarchy_level + 1
        INTO new_hierarchy_level
        FROM dbo.classes
        WHERE id = parent_class_id;

        IF NOT found THEN
            RAISE EXCEPTION 'Error: parent_class_id % 不存在，無法建立 class。', parent_class_id USING ERRCODE = '22000';
        END IF;
    END IF;

    RETURN QUERY INSERT INTO dbo.classes (id, entity_id, chinese_name, chinese_description, english_name,
                                          english_description,
                                          owner_id,
                                          id_path,
                                          name_path,
                                          hierarchy_level)
        VALUES (new_class_id,
                fn_insert_class.entity_id,
                fn_insert_class.chinese_name,
                fn_insert_class.chinese_description,
                fn_insert_class.english_name,
                fn_insert_class.english_description,
                fn_insert_class.owner_id,
                new_id_path,
                new_name_path,
                new_hierarchy_level)
        RETURNING *;

    -- Inherit permissions from parent class
    INSERT INTO dbo.permissions
    SELECT new_class_id, role_type, role_id, permission_bits
    FROM dbo.permissions
    WHERE class_id = parent_class_id;

    RETURN;
END;
$$
    LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbo.fn_insert_entity(p_chinese_name CHARACTER VARYING DEFAULT NULL::CHARACTER VARYING,
                                                p_english_name CHARACTER VARYING DEFAULT NULL::CHARACTER VARYING,
                                                p_is_relational BOOLEAN DEFAULT NULL::BOOLEAN,
                                                p_custom_rank INTEGER DEFAULT NULL::INTEGER
) RETURNS CHARACTER VARYING(21)
    LANGUAGE plpgsql
AS
$$
DECLARE
    entity_id VARCHAR(21);
BEGIN
    -- 檢查參數是否為空
    IF (p_chinese_name IS NULL OR p_english_name IS NULL OR p_is_relational IS NULL) THEN
        RAISE EXCEPTION 'Error: 所有參數 (chinese_name, english_name, is_relational) 都必須提供' USING ERRCODE = '22000';
    END IF;

    -- 檢查 chinese_name 或 english_name 是否已存在
    IF EXISTS(SELECT 1
              FROM dbo.entities
              WHERE dbo.entities.chinese_name = p_chinese_name
                 OR dbo.entities.english_name = p_english_name) THEN
        RAISE EXCEPTION 'Error: chinese_name 或 english_name 已經存在，無法建立 entity' USING ERRCODE = '22000';
    END IF;

    -- 插入數據到 entities 表中並返回新插入的 ID
    -- 如果有提供 p_custom_rank，則使用它來覆蓋系統默認值
    IF p_custom_rank IS NOT NULL THEN
        INSERT INTO dbo.entities(chinese_name,
                                 english_name,
                                 is_relational,
                                 rank)
            OVERRIDING SYSTEM VALUE
        VALUES (p_chinese_name,
                p_english_name,
                p_is_relational,
                p_custom_rank)
        RETURNING id INTO entity_id;
    ELSE
        INSERT INTO dbo.entities(chinese_name,
                                 english_name,
                                 is_relational)
        VALUES (p_chinese_name,
                p_english_name,
                p_is_relational)
        RETURNING id INTO entity_id;
    END IF;

    -- 返回新插入的 entity_id
    RETURN entity_id;
END;
$$;

CREATE OR REPLACE FUNCTION dbo.fn_delete_class(
    class_id VARCHAR(21)
) RETURNS SETOF dbo.classes
    LANGUAGE plpgsql
AS
$$
DECLARE
BEGIN
    RETURN QUERY DELETE FROM dbo.classes WHERE id = class_id RETURNING *;
    IF NOT found THEN
        RAISE EXCEPTION 'Error: class_id % 不存在，無法刪除class。', id USING ERRCODE = '22000';
    END IF;
END;
$$;
