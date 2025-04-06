CREATE OR REPLACE FUNCTION api.fn_insert_class(
    hasura_session json,
    parent_class_id VARCHAR(21),
    entity_id VARCHAR(21),
    chinese_name TEXT,
    chinese_description TEXT,
    english_name TEXT,
    english_description TEXT,
    owner_id VARCHAR(21) DEFAULT NULL
) RETURNS SETOF dbo.classes AS
$$
$$
    LANGUAGE plpgsql;
