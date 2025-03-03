BEGIN;

INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5")
SELECT 'p', 'role:all-access-manager___', 'user/*/*', '*', '*/*/*/*/*', 'allow', ''
    WHERE NOT EXISTS (
    SELECT 1 FROM "public"."casbin_rule"
    WHERE p_type = 'p'
      AND v0 = 'role:all-access-manager___'
      AND v1 = 'user/*/*'
      AND v2 = '*'
      AND v3 = '*/*/*/*/*'
      AND v4 = 'allow'
      AND v5 = ''
);

INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5")
SELECT 'p', 'role:super-admin___', 'user/*/*', '*', '*/*/*/*/*', 'allow', ''
    WHERE NOT EXISTS (
    SELECT 1 FROM "public"."casbin_rule"
    WHERE p_type = 'p'
      AND v0 = 'role:super-admin___'
      AND v1 = 'user/*/*'
      AND v2 = '*'
      AND v3 = '*/*/*/*/*'
      AND v4 = 'allow'
      AND v5 = ''
);

COMMIT;