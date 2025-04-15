DO
$$
    DECLARE
        web_resource_id         VARCHAR(21);
        user_entity_id          VARCHAR(21);
        root_class_id           VARCHAR(21);
        home_class_id           VARCHAR(21);
        default_home_permission SMALLINT := (SELECT BIT_OR(bit)
                                             FROM dbo.permission_enum
                                             WHERE id IN ('read-class', 'read-object'));
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM dbo.entities) THEN
            SELECT dbo.fn_insert_entity('WEB資源', 'WEB Resource', TRUE, 1) INTO web_resource_id;
            SELECT dbo.fn_insert_entity('使用者', 'User', TRUE, 2) INTO user_entity_id;
            PERFORM dbo.fn_insert_entity('檔案', 'File', TRUE, 3);
            PERFORM dbo.fn_insert_entity('公告', 'Announcement', TRUE, 4);
        END IF;

        IF NOT EXISTS(SELECT 1 FROM dbo.classes) THEN
            SELECT id
            FROM dbo.fn_insert_class(
                    parent_class_id := NULL,
                    entity_id := web_resource_id,
                    chinese_name := '/',
                    chinese_description := NULL,
                    english_name := 'Root',
                    english_description := NULL,
                    owner_id := NULL
                 )
            INTO root_class_id;

            PERFORM dbo.fn_insert_class(
                    parent_class_id := root_class_id,
                    entity_id := user_entity_id,
                    chinese_name := '使用者',
                    chinese_description := NULL,
                    english_name := 'User',
                    english_description := NULL,
                    owner_id := NULL
                    );

            SELECT id
            FROM dbo.fn_insert_class(
                    parent_class_id := root_class_id,
                    entity_id := web_resource_id,
                    chinese_name := '首頁',
                    chinese_description := NULL,
                    english_name := 'Home',
                    english_description := NULL,
                    owner_id := NULL
                 )
            INTO home_class_id;


            -- Insert default home permissions for user and guest roles
            INSERT INTO dbo.permissions(class_id, role_type, role_id, permission_bits)
            SELECT home_class_id, FALSE, roles.id, default_home_permission
            FROM auth.roles
            WHERE roles.name IN ('user', 'guest');
        END IF;
    END
$$;
