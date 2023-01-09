DELETE FROM "default_auth_role"
    WHERE role_type in ('clusterAdmin','clusterEdit','clusterView');

ALTER TABLE "roles"
    DROP COLUMN "cluster",
    DROP COLUMN "namespace",
    DROP COLUMN "group",
    DROP COLUMN "kind",
    DROP COLUMN "resource";

DELETE FROM "default_auth_policy"
    WHERE role_type in ('clusterAdmin','clusterEdit','clusterView');