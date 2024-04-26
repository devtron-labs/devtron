CREATE SEQUENCE IF NOT EXISTS id_seq_watcher;
CREATE TABLE "public"."watcher" (
                                    "id" integer NOT NULL default nextval('id_seq_watcher'::regclass),
                                    "name" varchar(50) NOT NULL,
                                    "description" text ,
                                    "filter_expression" text NOT NULL,
                                    "gvks" text,
                                    "active" bool NOT NULL,
                                    "created_on"                timestamptz NOT NULL,
                                    "created_by"                int4        NOT NULL,
                                    "updated_on"                timestamptz,
                                    "updated_by"                int4,

                                    PRIMARY KEY ("id"),
                                        UNIQUE ("name")

);
CREATE UNIQUE INDEX "idx_unique_watcher_name"
    ON watcher(name)
    WHERE  watcher.active =true;

CREATE SEQUENCE IF NOT EXISTS id_seq_trigger;
CREATE TABLE "public"."trigger"(
                                   "id" integer NOT NULL default nextval('id_seq_trigger'::regclass),
                                   "type" varchar(50) , -- DEVTRON_JOB
                                   "watcher_id" integer ,
                                   "data" text,
                                   "active" bool NOT NULL,
                                   "created_on"                timestamptz NOT NULL,
                                   "created_by"                int4        NOT NULL,
                                   "updated_on"                timestamptz,
                                   "updated_by"                int4,

                                   CONSTRAINT trigger_watcher_id_fkey FOREIGN KEY ("watcher_id") REFERENCES "public"."watcher" ("id"),
                                   PRIMARY KEY ("id")

);

CREATE SEQUENCE IF NOT EXISTS id_seq_intercepted_events;
CREATE TABLE "public"."intercepted_event_execution"(
                                              "id" integer NOT NULL default nextval('id_seq_intercepted_events'::regclass),
                                              "cluster_id" int ,
                                              "namespace" character varying(250) NOT NULL,
                                              "message" text,
                                              "execution_message" text,
                                              "message_type" varchar(100),
                                              "event" text,
                                              "involved_object" text,
                                              "intercepted_at" timestamptz,
                                              "status" varchar(32) , -- failure,success,inprogress
                                              "trigger_id" integer,
                                              "trigger_execution_id" integer,
                                              "created_on"                timestamptz NOT NULL,

                                              CONSTRAINT intercepted_events_trigger_id_fkey FOREIGN KEY ("trigger_id") REFERENCES "public"."trigger" ("id"),
                                              CONSTRAINT intercepted_events_cluster_id_fkey FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster" ("id"),
                                              PRIMARY KEY ("id")
);

