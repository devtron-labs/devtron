/*
 * Copyright (c) 2024. Devtron Inc.
 */

INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES
                                                                                      ('p','role:super-admin___','release','*','*','allow',''),
                                                                                      ('p','role:super-admin___','release-requirement','*','*','allow',''),
                                                                                      ('p','role:super-admin___','release-track-requirement','*','*','allow',''),
                                                                                      ('p','role:super-admin___','release-track','*','*','allow','');