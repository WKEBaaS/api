DO
$$
    DECLARE
        q_entity_id      VARCHAR(21);
        root_class_id    VARCHAR(21);
        normal_user_id   VARCHAR(21);
        admin_user_id    VARCHAR(21);
        problem_class_id VARCHAR(21);
    BEGIN
        SELECT dbo.fn_insert_entity('題目', 'Question', TRUE) INTO q_entity_id;
        SELECT id FROM postgres.dbo.classes WHERE name_path = '/' INTO root_class_id;

        INSERT INTO dbo.objects(chinese_name) VALUES ('超級使用者') RETURNING id INTO admin_user_id;
        INSERT INTO auth.users(id, role, email) VALUES (admin_user_id, 'admin', 'wke_admin@wke.csie.ncnu.edu.tw');

        INSERT INTO dbo.objects(chinese_name) VALUES ('一般使用者') RETURNING id INTO normal_user_id;
        INSERT INTO auth.users(id, role, email) VALUES (normal_user_id, 'user', 'wke_normal@wke.csie.ncnu.edu.tw');

        SELECT id
        FROM dbo.fn_insert_class(
                parent_class_id := root_class_id,
                entity_id := q_entity_id,
                chinese_name := '題目',
                chinese_description := '題目描述',
                english_name := 'Question',
                english_description := 'Question description',
                owner_id := NULL
             )
        INTO problem_class_id;

        INSERT INTO postgres.dbo.objects(entity_id, chinese_name, chinese_description, english_name,
                                         english_description, owner_id)
        VALUES (q_entity_id, '題目1', '題目描述1', 'Question1', 'Description1', normal_user_id),
               (q_entity_id, '題目2', '題目描述2', 'Question2', 'Description2', normal_user_id),
               (q_entity_id, '題目3', '題目描述3', 'Question3', 'Description3', normal_user_id);

        INSERT INTO dbo.co (cid, oid) SELECT problem_class_id, id FROM postgres.dbo.objects;

        INSERT INTO dbo.permissions (class_id, role_type, role_id, permission_bits)
        VALUES (problem_class_id, TRUE, normal_user_id, 1 | 2 | 4);

        INSERT INTO auth.user_roles(user_id, role_id)
        VALUES (admin_user_id, (SELECT id FROM auth.roles WHERE name = 'admin'));
    END
$$;