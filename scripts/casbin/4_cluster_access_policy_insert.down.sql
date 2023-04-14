DELETE FROM "public"."casbin_rule"
       WHERE  ("p_type" = 'p' AND "v0" = 'role:super-admin___' AND "v1" = '*/*' AND "v2" = '*' AND "v3" = '*/*/*' AND "v4" = 'allow' AND "v5" = '');

DELETE FROM "public"."casbin_rule"
WHERE  ("p_type" = 'p' AND "v0" = 'role:super-admin___' AND "v1" = '*/*/user' AND "v2" = '*' AND "v3" = '*/*/*' AND "v4" = 'allow' AND "v5" = '');