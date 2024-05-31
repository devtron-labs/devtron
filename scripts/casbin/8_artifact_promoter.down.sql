/*
 * Copyright (c) 2024. Devtron Inc.
 */

DELETE FROM "public"."casbin_rule" WHERE ("p_type" = 'p' AND "v0" = 'role:super-admin___' AND "v1" = 'artifact' AND "v2" = '*' AND "v3" = '*' AND "v4" = 'allow' AND "v5" = '');