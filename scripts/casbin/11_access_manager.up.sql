BEGIN;
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES
    ('p','role:all-access-manager___','user/*/*','*','*/*/*/*/*','allow','');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES
    ('p','role:super-admin___','user/*/*','*','*/*/*/*/*','allow','');
COMMIT;