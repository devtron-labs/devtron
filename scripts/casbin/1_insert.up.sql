/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

CREATE DATABASE casbin;

-- Table Definition
CREATE TABLE "public"."casbin_rule" (
    "p_type" varchar(100) NOT NULL DEFAULT ''::character varying,
    "v0" varchar(100) NOT NULL DEFAULT ''::character varying,
    "v1" varchar(100) NOT NULL DEFAULT ''::character varying,
    "v2" varchar(100) NOT NULL DEFAULT ''::character varying,
    "v3" varchar(100) NOT NULL DEFAULT ''::character varying,
    "v4" varchar(100) NOT NULL DEFAULT ''::character varying,
    "v5" varchar(100) NOT NULL DEFAULT ''::character varying
);


INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'docker', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'team', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'admin', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'migrate', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'environment', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'git', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'notification', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'user', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'cluster', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'applications', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'global-environment', '*', '*', 'allow', '');
INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('p', 'role:super-admin___', 'chart-group', '*', '*', 'allow', '');

INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES ('g', 'admin', 'role:super-admin___', '', '', '', '');

INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5") VALUES
('p', 'role:chart-group_admin', 'chart-group', '*', '*', 'allow', ''),
('g', 'admin', 'role:chart-group_admin', '', '', '', '');