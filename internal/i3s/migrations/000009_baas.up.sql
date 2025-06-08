INSERT INTO dbo.entities (chinese_name, english_name)
VALUES ('專案', 'Project');

CREATE TABLE dbo.projects
(
    id        VARCHAR(21) NOT NULL UNIQUE,
    reference VARCHAR(20) NOT NULL UNIQUE,
    CONSTRAINT pk_dbo_projects PRIMARY KEY (id),
    CONSTRAINT fk_dbo_projects_id FOREIGN KEY (id) REFERENCES dbo.objects
);

CREATE OR REPLACE VIEW dbo.vd_projects AS
(
SELECT o.id,
       o.chinese_name        name,
       o.chinese_description description,
       o.owner_id,
       o.entity_id,
       p.reference,
       o.created_at,
       o.updated_at
FROM dbo.objects o,
     dbo.projects p
WHERE o.id = p.id
    );
