DO
$$
    DECLARE
        role_entity_id VARCHAR(21);
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM dbo.entities) THEN
            SELECT dbo.fn_insert_entity('權限角色', 'Role', TRUE, 1) INTO role_entity_id;
            PERFORM dbo.fn_insert_entity('資源', 'Resource', TRUE, 2);
            INSERT INTO dbo.classes(entity_id, chinese_name, chinese_description, english_name, english_description,
                                    id_path,
                                    name_path, hierarchy_level)
            SELECT role_entity_id,
                   '/',
                   '初始的根目錄，之後的所有Class都會掛在根目錄底下',
                   'Root',
                   'Initial root, all subsequent Classes will mount under the root',
                   '/' || role_entity_id,
                   '/',
                   0
            WHERE NOT EXISTS(SELECT 1
                             FROM dbo.classes
                             WHERE entity_id = role_entity_id
                               AND chinese_name = '/');
        END IF;
    END
$$;
