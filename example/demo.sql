DO
$$
    DECLARE
        q_entity_id      VARCHAR(21);
        root_class_id    VARCHAR(21);
        user_id          VARCHAR(21);
        problem_class_id VARCHAR(21);
    BEGIN
        SELECT dbo.fn_insert_entity('題目', 'Question', TRUE) INTO q_entity_id;
        SELECT id FROM postgres.dbo.classes WHERE name_path = '/' INTO root_class_id;

        INSERT INTO auth.users(aud, email) VALUES ('admin', 'wke@wke.csie.ncnu.edu.tw') RETURNING id INTO user_id;


        SELECT dbo.fn_insert_class(root_class_id, q_entity_id, '題目', '題目描述', user_id)
        INTO problem_class_id;

        INSERT INTO postgres.dbo.objects(entity_id, chinese_name, chinese_description, english_name,
                                         english_description, owner_id)
        VALUES (q_entity_id, '題目1', '題目描述1', 'Question1', 'Description1', user_id),
               (q_entity_id, '題目2', '題目描述2', 'Question2', 'Description2', user_id),
               (q_entity_id, '題目3', '題目描述3', 'Question3', 'Description3', user_id);

        INSERT INTO dbo.co (cid, oid) SELECT problem_class_id, id FROM postgres.dbo.objects;
    END
$$;