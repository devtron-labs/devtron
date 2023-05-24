-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_external_link;

-- Table Definition
CREATE TABLE "public"."external_link" (
                                          "id" int4 NOT NULL DEFAULT nextval('id_seq_external_link'::regclass),
                                          "external_link_monitoring_tool_id" int4 NOT NULL,
                                          "name" varchar(255) NOT NULL,
                                          "url" varchar(255),
                                          "active" bool NOT NULL,
                                          "created_on" timestamptz,
                                          "created_by" int4,
                                          "updated_on" timestamptz,
                                          "updated_by" int4,
                                          PRIMARY KEY ("id")
);

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_external_link_cluster_mapping;

-- Table Definition
CREATE TABLE "public"."external_link_cluster_mapping" (
                                                          "id" int4 NOT NULL DEFAULT nextval('id_seq_external_link_cluster_mapping'::regclass),
                                                          "external_link_id" int4 NOT NULL,
                                                          "cluster_id" int4 NOT NULL,
                                                          "active" bool NOT NULL,
                                                          "created_on" timestamptz,
                                                          "created_by" int4,
                                                          "updated_on" timestamptz,
                                                          "updated_by" int4,
                                                          PRIMARY KEY ("id")
);

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_external_link_monitoring_tool;

-- Table Definition
CREATE TABLE "public"."external_link_monitoring_tool" (
                                                          "id" int4 NOT NULL DEFAULT nextval('id_seq_external_link_monitoring_tool'::regclass),
                                                          "name" varchar(255) NOT NULL,
                                                          "icon" varchar(255),
                                                          "active" bool NOT NULL,
                                                          "created_on" timestamptz,
                                                          "created_by" int4,
                                                          "updated_on" timestamptz,
                                                          "updated_by" int4,
                                                          PRIMARY KEY ("id")
);

ALTER TABLE "public"."external_link" ADD FOREIGN KEY ("external_link_monitoring_tool_id") REFERENCES "public"."external_link_monitoring_tool"("id");
ALTER TABLE "public"."external_link_cluster_mapping" ADD FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster"("id");
ALTER TABLE "public"."external_link_cluster_mapping" ADD FOREIGN KEY ("external_link_id") REFERENCES "public"."external_link"("id");


INSERT INTO "public"."external_link_monitoring_tool" ("name", "icon", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
('Grafana', '', 't', 'now()', 1, 'now()', 1),
('Kibana', '', 't', 'now()', 1, 'now()', 1),
('Newrelic', '', 't', 'now()', 1, 'now()', 1),
('Coralogix', '', 't', 'now()', 1, 'now()', 1),
('Datadog', '', 't', 'now()', 1, 'now()', 1),
('Loki', '', 't', 'now()', 1, 'now()', 1),
('Cloudwatch', '', 't', 'now()', 1, 'now()', 1),
('Other', '', 't', 'now()', 1, 'now()', 1);