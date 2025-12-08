-- revision: 1_initial_db_schema_up.sql
-- description: Initial database schema creation
-- created: 2025-10-25 12:00:00



-- DROP TYPE public.approvalstatus;

CREATE TYPE public.approvalstatus AS ENUM (
	'APPROVAL_PENDING',
	'APPROVED',
	'REJECTED');

-- DROP TYPE public.messagestatusenum;

CREATE TYPE public.messagestatusenum AS ENUM (
	'QUEUED',
	'STARTED',
	'FAILED',
	'COMPLETED',
	'IN_TRANSIT');

-- DROP TYPE public.recipienttypeenum;

CREATE TYPE public.recipienttypeenum AS ENUM (
	'USER',
	'BOT',
	'SYSTEM');

-- DROP TYPE public.recommendationfeedpriority;

CREATE TYPE public.recommendationfeedpriority AS ENUM (
	'HIGH',
	'MEDIUM',
	'LOW');

-- DROP TYPE public.recommendationfeedstatus;

CREATE TYPE public.recommendationfeedstatus AS ENUM (
	'PENDING_DISCOVERY',
	'RECOMMENDATION_DISCOVERED',
	'ACTION_IN_PROGRESS',
	'ACTION_REQUIRED',
	'READY_FOR_TRIGGER',
	'COMPLETED',
	'REJECTED',
	'ERRORED',
	'REVERTED');

-- DROP TYPE public.runbookstatus;

CREATE TYPE public.runbookstatus AS ENUM (
	'ACTIVE',
	'DELETED',
	'DRAFT');

-- DROP TYPE public.sendertypeenum;

CREATE TYPE public.sendertypeenum AS ENUM (
	'USER',
	'BOT',
	'SYSTEM');

-- DROP TYPE public.threadstatusenum;

CREATE TYPE public.threadstatusenum AS ENUM (
	'ACTIVE',
	'ARCHIVED',
	'DELETED');

-- DROP TYPE public.userrecommendationfeedinteractionstatus;

CREATE TYPE public.userrecommendationfeedinteractionstatus AS ENUM (
	'READ',
	'UNREAD');

-- DROP TYPE public.windowtype;

CREATE TYPE public.windowtype AS ENUM (
	'RELATIVE',
	'ABSOLUTE');

-- DROP TYPE public.windowunit;

CREATE TYPE public.windowunit AS ENUM (
	'HOURS',
	'MINUTES',
	'SECONDS',
	'DAYS',
	'WEEKS',
	'MONTHS',
	'YEARS',
	'EPOCH_TIME');


-- public.chat_threads definition

-- Drop table

-- DROP TABLE public.chat_threads;

CREATE TABLE public.chat_threads (
                                     id varchar(32) NOT NULL,
                                     thread_name varchar(255) NOT NULL,
                                     description text NULL,
                                     created_by varchar(255) NULL,
                                     status public.threadstatusenum NOT NULL,
                                     created_at timestamp NOT NULL,
                                     CONSTRAINT chat_threads_pkey PRIMARY KEY (id)
);

-- public.master_audit_logs definition

-- Drop table

-- DROP TABLE public.master_audit_logs;

CREATE TABLE public.master_audit_logs (
                                          id varchar(32) NOT NULL,
                                          "action" varchar(100) NOT NULL,
                                          action_type varchar(255) NOT NULL,
                                          user_id varchar(255) NOT NULL,
                                          resource_id varchar(255) NOT NULL,
                                          resource_type varchar(100) NOT NULL,
                                          resource_name varchar(255) NULL,
                                          correlation_id varchar(255) NULL,
                                          "module" varchar(100) NOT NULL,
                                          subject_attributes jsonb NULL,
                                          resource_attributes jsonb NULL,
                                          payload jsonb NOT NULL,
                                          created_at timestamp NOT NULL,
                                          created_by varchar(255) NOT NULL,
                                          CONSTRAINT master_audit_logs_pkey PRIMARY KEY (id)
);


-- public.recommendation_feeds definition

-- Drop table

-- DROP TABLE public.recommendation_feeds;

CREATE TABLE public.recommendation_feeds (
                                             id varchar(32) NOT NULL,
                                             processed_by varchar(255) NULL,
                                             title varchar(255) NOT NULL,
                                             description text NULL,
                                             tags json NULL,
                                             resource_attributes json NULL,
                                             priority public.recommendationfeedpriority NOT NULL,
                                             status public.recommendationfeedstatus NOT NULL,
                                             cluster_name varchar(255) NULL,
                                             created_at timestamp NOT NULL,
                                             updated_at timestamp NULL,
                                             created_by varchar(255) NOT NULL,
                                             updated_by varchar(255) NOT NULL,
                                             CONSTRAINT recommendation_feeds_pkey PRIMARY KEY (id)
);


-- public.runbook definition

-- Drop table

-- DROP TABLE public.runbook;

CREATE TABLE public.runbook (
                                runbook_id varchar(32) NOT NULL,
                                status public.runbookstatus NOT NULL,
                                "name" varchar(255) NOT NULL,
                                description text NULL,
                                "content" json NULL,
                                meta_data json NULL,
                                tags json NULL,
                                created_at timestamp NOT NULL,
                                updated_at timestamp NOT NULL,
                                created_by varchar(255) NOT NULL,
                                updated_by varchar(255) NOT NULL,
                                deleted_at timestamp NULL,
                                deleted_by varchar(255) NULL,
                                last_triggered_at timestamp NULL,
                                active bool NULL,
                                CONSTRAINT runbook_pkey PRIMARY KEY (runbook_id)
);


-- public.wrk_eng_activity_trail definition

-- Drop table

-- DROP TABLE public.wrk_eng_activity_trail;

CREATE TABLE public.wrk_eng_activity_trail (
                                               id varchar(32) NOT NULL,
                                               feed_id varchar(60) NULL,
                                               created_at timestamp NOT NULL,
                                               activity text NOT NULL,
                                               created_by varchar(255) NOT NULL,
                                               correlation_id varchar(255) NULL,
                                               "data" json NULL,
                                               CONSTRAINT wrk_eng_activity_trail_pkey PRIMARY KEY (id)
);


-- public.wrk_eng_resource_discovery definition

-- Drop table

-- DROP TABLE public.wrk_eng_resource_discovery;

CREATE TABLE public.wrk_eng_resource_discovery (
                                                   res_id varchar(32) NOT NULL,
                                                   cluster_id int4 NOT NULL,
                                                   "group" varchar(255) NULL,
                                                   "version" varchar(255) NULL,
                                                   kind varchar(255) NULL,
                                                   "name" varchar(255) NULL,
                                                   "namespace" varchar(255) NULL,
                                                   labels jsonb NULL,
                                                   created_at timestamp NOT NULL,
                                                   CONSTRAINT uq_wrk_eng_resource_discovery_composite UNIQUE (cluster_id, version, kind, namespace, name, "group"),
                                                   CONSTRAINT wrk_eng_resource_discovery_pkey PRIMARY KEY (res_id)
);


-- public.wrk_eng_resource_recommendation_bucket definition

-- Drop table

-- DROP TABLE public.wrk_eng_resource_recommendation_bucket;

CREATE TABLE public.wrk_eng_resource_recommendation_bucket (
                                                               id varchar(32) NOT NULL,
                                                               feed_id varchar(60) NULL,
                                                               res_id varchar(32) NOT NULL,
                                                               thread_id varchar(32) NOT NULL,
                                                               container_name varchar(255) NOT NULL,
                                                               cpu_request json NULL,
                                                               cpu_limit json NULL,
                                                               memory_request json NULL,
                                                               memory_limit json NULL,
                                                               created_at timestamp NOT NULL,
                                                               updated_at timestamp NOT NULL,
                                                               status varchar(48) NULL,
                                                               status_updated_at timestamp NOT NULL,
                                                               correlation_id varchar(255) NULL,
                                                               meta_data json NULL,
                                                               history json NULL,
                                                               CONSTRAINT uq_wrk_eng_resource_recommendation_bucket_composite UNIQUE (res_id, container_name),
                                                               CONSTRAINT wrk_eng_resource_recommendation_bucket_pkey PRIMARY KEY (id)
);


-- public.wrk_eng_runbook_transition_log definition

-- Drop table

-- DROP TABLE public.wrk_eng_runbook_transition_log;

CREATE TABLE public.wrk_eng_runbook_transition_log (
                                                       id varchar(64) NOT NULL,
                                                       runbook_id varchar(64) NOT NULL,
                                                       step_order int4 NOT NULL,
                                                       feed_id varchar(64) NULL,
                                                       thread_id varchar(64) NOT NULL,
                                                       "action" varchar(255) NOT NULL,
                                                       action_type varchar(64) NOT NULL,
                                                       action_name varchar(255) NULL,
                                                       change_type varchar(32) NULL,
                                                       status varchar(32) NULL,
                                                       current_snapshot json NULL,
                                                       proposed_snapshot json NULL,
                                                       previous_snapshot json NULL,
                                                       new_snapshot json NULL,
                                                       applied_params json NULL,
                                                       master_data json NULL,
                                                       created_at timestamp NOT NULL,
                                                       created_by varchar(255) NOT NULL,
                                                       updated_at timestamp NULL,
                                                       updated_by varchar(255) NULL,
                                                       meta_data json NULL,
                                                       transition_log json NULL,
                                                       status_message varchar(1024) NULL,
                                                       CONSTRAINT wrk_eng_runbook_transition_log_pkey PRIMARY KEY (id)
);


-- public.wrk_eng_thought_process definition

-- Drop table

-- DROP TABLE public.wrk_eng_thought_process;

CREATE TABLE public.wrk_eng_thought_process (
                                                id varchar(32) NOT NULL,
                                                thread_id varchar(32) NOT NULL,
                                                feed_id varchar(60) NULL,
                                                "content" text NOT NULL,
                                                created_at timestamp NOT NULL,
                                                created_by varchar(255) NOT NULL,
                                                correlation_id varchar(255) NULL,
                                                CONSTRAINT wrk_eng_thought_process_pkey PRIMARY KEY (id)
);


/*
 *
 *  * Copyright (c) 2020-2024. Devtron Inc.
 *
 */


-- public.chat_messages definition

-- Drop table

-- DROP TABLE public.chat_messages;

CREATE TABLE public.chat_messages (
                                      id varchar(32) NOT NULL,
                                      thread_id varchar(32) NOT NULL,
                                      sender_id varchar(255) NOT NULL,
                                      recipient_id varchar(255) NOT NULL,
                                      sender_type public.sendertypeenum NOT NULL,
                                      recipient_type public.recipienttypeenum NOT NULL,
                                      sequence_id int4 NOT NULL,
                                      metadata jsonb NULL,
                                      status public.messagestatusenum NOT NULL,
                                      generated_at timestamp NOT NULL,
                                      created_at timestamp NOT NULL,
                                      CONSTRAINT chat_messages_pkey PRIMARY KEY (id),
                                      CONSTRAINT uq_thread_sequence UNIQUE (thread_id, sequence_id),
                                      CONSTRAINT chat_messages_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.chat_threads(id)
);


-- public.runbook_approval_status definition

-- Drop table

-- DROP TABLE public.runbook_approval_status;

CREATE TABLE public.runbook_approval_status (
                                                approval_id varchar(32) NOT NULL,
                                                runbook_id varchar(32) NOT NULL,
                                                status public.approvalstatus NOT NULL,
                                                cluster_id int4 NOT NULL,
                                                status_effective_at timestamp NOT NULL,
                                                status_effective_until timestamp NULL,
                                                window_value int8 NULL,
                                                window_unit public.windowunit NULL,
                                                window_type public.windowtype NULL,
                                                meta_data json NULL,
                                                created_at timestamp NOT NULL,
                                                created_by varchar(255) NOT NULL,
                                                updated_at timestamp NOT NULL,
                                                updated_by varchar(255) NOT NULL,
                                                CONSTRAINT runbook_approval_status_pkey PRIMARY KEY (approval_id),
                                                CONSTRAINT runbook_approval_status_runbook_id_fkey FOREIGN KEY (runbook_id) REFERENCES public.runbook(runbook_id) ON DELETE CASCADE
);


-- public.user_recommendation_feed_interactions definition

-- Drop table

-- DROP TABLE public.user_recommendation_feed_interactions;

CREATE TABLE public.user_recommendation_feed_interactions (
                                                              id varchar(32) NOT NULL,
                                                              user_id varchar(255) NOT NULL,
                                                              recommendation_feed_id varchar(32) NOT NULL,
                                                              status public.userrecommendationfeedinteractionstatus NOT NULL,
                                                              read_at timestamp NULL,
                                                              created_at timestamp NOT NULL,
                                                              updated_at timestamp NULL,
                                                              created_by varchar(255) NOT NULL,
                                                              updated_by varchar(255) NULL,
                                                              CONSTRAINT user_recommendation_feed_interactions_pkey PRIMARY KEY (id),
                                                              CONSTRAINT user_recommendation_feed_interactio_recommendation_feed_id_fkey FOREIGN KEY (recommendation_feed_id) REFERENCES public.recommendation_feeds(id),
                                                              CONSTRAINT uq_user_recommendation_feed_interactions_feed_id_user_id UNIQUE (recommendation_feed_id, user_id)
);


-- public.chat_message_parts definition

-- Drop table

-- DROP TABLE public.chat_message_parts;

CREATE TABLE public.chat_message_parts (
                                           message_id varchar(32) NOT NULL,
                                           id int4 NOT NULL,
                                           part_type varchar(50) NOT NULL,
                                           "content" json NOT NULL,
                                           created_at timestamp NOT NULL,
                                           updated_at timestamp NOT NULL,
                                           CONSTRAINT chat_message_parts_pk PRIMARY KEY (message_id, id),
                                           CONSTRAINT chat_message_parts_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.chat_messages(id)
);


-- public.chat_graph_data definition

-- Drop table

-- DROP TABLE public.chat_graph_data;

CREATE TABLE public.chat_graph_data (
                                        id varchar(32) NOT NULL,
                                        message_id varchar(32) NOT NULL,
                                        graph_type varchar(50) NOT NULL,
                                        title varchar(255) NOT NULL,
                                        description text NULL,
                                        "data" jsonb NOT NULL,
                                        part_type varchar(50) NOT NULL,
                                        part_id int4 NOT NULL,
                                        created_at timestamp NOT NULL,
                                        CONSTRAINT chat_graph_data_pkey PRIMARY KEY (id),
                                        CONSTRAINT chat_graph_data_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.chat_messages(id),
                                        CONSTRAINT fk_graph_part FOREIGN KEY (message_id,part_id) REFERENCES public.chat_message_parts(message_id,id)
);