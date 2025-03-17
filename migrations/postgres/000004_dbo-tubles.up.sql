DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1
                      FROM dbo.entities
                      WHERE chinese_name = '權限角色') THEN
            PERFORM dbo.fn_insert_entity('權限角色', 'Role', TRUE);
        END IF;
    END
$$;

DO
$$
    BEGIN
        BEGIN
            PERFORM dbo.fn_insert_entity('資源', 'Resource', TRUE);
        EXCEPTION
            WHEN OTHERS THEN
                RAISE NOTICE '忽略插入資源的錯誤: %', SQLERRM;
        END;
        INSERT INTO dbo.classes(entity_id, chinese_name, chinese_description, english_name, english_description,
                                id_path,
                                name_path, hierarchy_level)
        SELECT 1,
               '根目錄',
               '初始的根目錄，之後的所有Class都會掛在根目錄底下',
               'Root',
               'Initial root, all subsequent Classes will mount under the root',
               '1',
               '根目錄',
               1
        WHERE NOT EXISTS(SELECT 1
                         FROM dbo.classes
                         WHERE entity_id = 1
                           AND chinese_name = '根目錄');
    END
$$;