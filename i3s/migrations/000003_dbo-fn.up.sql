CREATE OR REPLACE FUNCTION dbo.fn_get_name_path(_parent_class_id INT, _chinese_name VARCHAR(255))
    RETURNS VARCHAR(900)
AS
$$
DECLARE
    hierarchy_level INT;
    name_path       VARCHAR(900);
BEGIN
    -- Check if _chinese_name is NULL
    IF _chinese_name IS NULL THEN
        -- Return the name_path of the parent class directly
        SELECT c.name_path
        INTO name_path
        FROM dbo.classes c
        WHERE c.cid = _parent_class_id;
        RETURN name_path;
    END IF;
    -- Retrieve the hierarchy level of the parent class
    SELECT c.hierarchy_level
    INTO hierarchy_level
    FROM dbo.classes c
    WHERE c.cid = _parent_class_id;
    -- Determine name_path based on the hierarchy_level
    IF hierarchy_level = 0 THEN
        -- Root level uses the chinese_name directly
        name_path := _chinese_name;
    ELSE
        -- Concatenate the parent's name_path and current node's chinese_name
        SELECT c.name_path || '/' || _chinese_name
        INTO name_path
        FROM dbo.classes c
        WHERE c.cid = _parent_class_id;
    END IF;
    -- Return the computed name_path
    RETURN name_path;
END;
$$
    LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbo.fn_get_id_path(_parent_class_id INT, _class_id INT)
    RETURNS VARCHAR(900)
AS
$$
DECLARE
    hierarchy_level INT;
    id_path         VARCHAR(900);
BEGIN
    -- Check if _class_id is NULL
    IF _class_id IS NULL THEN
        -- Return the id_path of the parent class directly
        SELECT c.id_path
        INTO id_path
        FROM dbo.classes c
        WHERE c.cid = _parent_class_id;
        RETURN id_path;
    END IF;
    -- Retrieve the hierarchy level of the parent class
    SELECT c.hierarchy_level
    INTO hierarchy_level
    FROM dbo.classes c
    WHERE c.cid = _parent_class_id;
    -- Check if the hierarchy_level is 0
    IF hierarchy_level = 0 THEN
        -- If root level, the id_path is just the current class_id
        id_path := CAST(_class_id AS VARCHAR);
    ELSE
        -- Concatenate the parent's id_path and the current class_id
        SELECT COALESCE(c.id_path, '') || '/' || CAST(_class_id AS VARCHAR)
        INTO id_path
        FROM dbo.classes c
        WHERE c.cid = _parent_class_id;
    END IF;
    -- Return the constructed id_path
    RETURN id_path;
END;
$$
    LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbo.fn_insert_class(_parent_class_id INT, _entity_id INT, _chinese_name VARCHAR(256),
                                               _chinese_description VARCHAR(4000), _owner_id UUID)
    RETURNS INT
AS
$$
DECLARE
    new_class_id          INT;
    parent_class_id_local INT;
    parent_name_path      VARCHAR(450);
    new_name_path         VARCHAR(450);
    new_level             SMALLINT;
    new_id_path           VARCHAR(900);
    error_message         TEXT;
    error_context         TEXT;
BEGIN
    -- Initialize the output value
    new_class_id := NULL;
    -- Fetch parent class details
    SELECT cid,
           name_path
    INTO parent_class_id_local,
        parent_name_path
    FROM dbo.classes
    WHERE cid = _parent_class_id;
    IF parent_class_id_local IS NOT NULL THEN
        -- Generate new name path
        new_name_path := dbo.fn_get_name_path(_parent_class_id, _chinese_name);
        -- Check if the class already exists
        IF (SELECT COUNT(*)
            FROM dbo.classes
            WHERE name_path = new_name_path) = 0 THEN
            -- Insert into dbo.class and get the new class_id
            INSERT INTO dbo.classes(entity_id,
                                  chinese_name,
                                  chinese_description,
                                  owner_uid)
            VALUES (_entity_id,
                    _chinese_name,
                    _chinese_description,
                    _owner_id)
            RETURNING
                cid INTO new_class_id;
            -- Insert into inheritance or relevant hierarchy table
            INSERT INTO dbo.inheritances(pcid, ccid)
            VALUES (parent_class_id_local,
                    new_class_id);
            -- Calculate new level and update the dbo.class record
            new_level := (SELECT hierarchy_level + 1
                          FROM dbo.classes
                          WHERE cid = _parent_class_id);
            new_id_path := dbo.fn_get_id_path(_parent_class_id, new_class_id);
            INSERT INTO dbo.permissions(class_id,
                                       role_type,
                                       role_id,
                                       permission_bits)
            SELECT new_class_id,
                   role_type,
                   role_id,
                   permission_bits
            FROM dbo.permissions
            WHERE class_id = _parent_class_id;
            UPDATE
                dbo.classes
            SET hierarchy_level = new_level,
                name_path       = new_name_path,
                id_path         = new_id_path
            WHERE cid = new_class_id;
        ELSE
            -- Raise an error if the class already exists
            RAISE EXCEPTION 'Error: Class already exists, cannot insert.';
        END IF;
    ELSE
        -- Raise an error if no parent class ID is provided
        RAISE EXCEPTION 'Error: Parent class ID not found.';
    END IF;
    -- Return the newly created class ID
    RETURN new_class_id;
EXCEPTION
    WHEN OTHERS THEN
        -- Capture error details and raise the error with additional context
        GET STACKED DIAGNOSTICS error_message = MESSAGE_TEXT,
            error_context = PG_EXCEPTION_CONTEXT;
        RAISE EXCEPTION 'Error code: %, Error message: %, Context: %', SQLSTATE, error_message, error_context;
END;

$$
    LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbo.fn_insert_entity(
    IN p_chinese_name VARCHAR(50) DEFAULT NULL,
    IN p_english_name VARCHAR(50) DEFAULT NULL,
    IN p_is_relational BOOLEAN DEFAULT NULL
) RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
    -- 檢查參數是否為空
    IF (p_chinese_name IS NULL OR p_english_name IS NULL OR p_is_relational IS NULL) THEN
        RAISE EXCEPTION 'Error: 所有參數 (chinese_name, english_name, is_relational) 都必須提供';
    END IF;
    -- 檢查 chinese_name 或 english_name 是否已存在
    IF EXISTS(SELECT 1
              FROM dbo.entities
              WHERE dbo.entities.chinese_name = p_chinese_name
                 OR dbo.entities.english_name = p_english_name) THEN
        RAISE EXCEPTION 'Error: chinese_name 或 english_name 已經存在，無法建立 entity';
    END IF;
    -- 插入數據到 entities 表中
    INSERT INTO dbo.entities(chinese_name,
                             english_name,
                             is_relational)
    VALUES (p_chinese_name,
            p_english_name,
            p_is_relational);
END;
$$;

CREATE OR REPLACE FUNCTION dbo.fn_delete_class(
    IN p_class_id INT
) RETURNS VOID
    LANGUAGE plpgsql
AS
$$
DECLARE
    error_message TEXT;
    error_context TEXT;
BEGIN
    -- Check if the class exists
    IF (SELECT COUNT(*)
        FROM dbo.classes
        WHERE cid = p_class_id) = 1 THEN
        -- Delete related records from class_object_relations and inheritances
        DELETE
        FROM dbo.co
        WHERE cid = p_class_id;
        DELETE
        FROM dbo.inheritances
        WHERE ccid = p_class_id
           OR pcid = p_class_id;
        DELETE
        FROM dbo.classes
        WHERE cid = p_class_id;
    ELSE
        -- Raise an error if the class does not exist
        RAISE EXCEPTION 'Error: This class does not exist';
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        -- Capture error details and raise the error with additional context
        GET STACKED DIAGNOSTICS error_message = MESSAGE_TEXT,
            error_context = PG_EXCEPTION_CONTEXT;
        RAISE EXCEPTION 'Error code: %, Error message: %, Context: %', SQLSTATE, error_message, error_context;
END;
$$;
