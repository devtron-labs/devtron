-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_external_links_monitoring_tools;

-- Table Definition
CREATE TABLE "public"."external_links_monitoring_tools" (
                                                            "id" int4 NOT NULL DEFAULT nextval('id_seq_external_links_monitoring_tools'::regclass),
                                                            "name" varchar(255),
                                                            "icon" varchar(255),
                                                            "active" bool,
                                                            "created_on" timestamptz,
                                                            "created_by" int4,
                                                            "updated_on" timestamptz,
                                                            "updated_by" int4,
                                                            PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_external_links;

-- Table Definition
CREATE TABLE "public"."external_links" (
                                           "id" int4 NOT NULL DEFAULT nextval('id_seq_external_links'::regclass),
                                           "external_links_monitoring_tool_id" int4,
                                           "name" varchar(255),
                                           "url" varchar(255),
                                           "active" bool,
                                           "created_on" timestamptz,
                                           "created_by" int4,
                                           "updated_on" timestamptz,
                                           "updated_by" int4,
                                           PRIMARY KEY ("id")
);


-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_external_links_clusters;

-- Table Definition
CREATE TABLE "public"."external_links_clusters" (
                                                    "id" int4 NOT NULL DEFAULT nextval('id_seq_external_links_clusters'::regclass),
                                                    "external_links_id" int4,
                                                    "cluster_id" int4,
                                                    "active" bool,
                                                    "created_on" timestamptz,
                                                    "created_by" int4,
                                                    "updated_on" timestamptz,
                                                    "updated_by" int4,
                                                    PRIMARY KEY ("id")
);
ALTER TABLE "public"."external_links" ADD FOREIGN KEY ("external_links_monitoring_tool_id") REFERENCES "public"."external_links_monitoring_tools"("id");
ALTER TABLE "public"."external_links_clusters" ADD FOREIGN KEY ("external_links_id") REFERENCES "public"."external_links"("id");
ALTER TABLE "public"."external_links_clusters" ADD FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster"("id");