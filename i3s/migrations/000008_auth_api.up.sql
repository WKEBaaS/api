CREATE OR REPLACE VIEW api.vd_user AS
SELECT u.*, o.chinese_name, o.chinese_description, o.created_at, o.updated_at, o.deleted_at
FROM auth.users u
         JOIN dbo.objects o ON u.id = o.id;

CREATE OR REPLACE FUNCTION api.create_user(
    chinese_name VARCHAR(255),
    email VARCHAR(255),
    chinese_description TEXT DEFAULT NULL,
    encrypted_password TEXT DEFAULT NULL
)
    RETURNS SETOF api.vd_user AS
$$
DECLARE
    v_user_id VARCHAR(21);
BEGIN
    INSERT INTO dbo.objects(chinese_name, chinese_description)
    VALUES (create_user.chinese_name, create_user.chinese_description)
    RETURNING id INTO v_user_id;

    INSERT INTO auth.users(id, email, encrypted_password)
    VALUES (v_user_id, create_user.email, create_user.encrypted_password);

    RETURN QUERY SELECT * FROM api.vd_user WHERE id = v_user_id;
END;
$$
    LANGUAGE plpgsql
    VOLATILE
