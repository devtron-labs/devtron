/*
 *
 *  * Copyright (c) 2020-2024. Devtron Inc.
 *
 */

-- revision: 1_initial_db_schema_up.sql
-- description: Initial database schema creation
-- created: 2025-10-25 12:00:00

DROP TABLE public.checkpoint_blobs CASCADE;

DROP TABLE public.checkpoint_migrations CASCADE;

DROP TABLE public.checkpoint_writes CASCADE;

DROP TABLE public.checkpoints CASCADE;



DROP TABLE public.chat_graph_data CASCADE;

DROP TABLE public.chat_message_parts CASCADE;

DROP TABLE public.chat_messages CASCADE;

DROP TABLE public.chat_threads CASCADE;

DROP TABLE public.master_audit_logs CASCADE;

DROP TABLE public.recommendation_feeds CASCADE;

DROP TABLE public.runbook CASCADE;

DROP TABLE public.runbook_approval_status CASCADE;

DROP TABLE public.user_recommendation_feed_interactions CASCADE;

DROP TABLE public.wrk_eng_activity_trail CASCADE;

DROP TABLE public.wrk_eng_resource_discovery CASCADE;

DROP TABLE public.wrk_eng_resource_recommendation_bucket CASCADE;

DROP TABLE public.wrk_eng_runbook_transition_log CASCADE;

DROP TABLE public.wrk_eng_thought_process CASCADE;



DROP TYPE public.approvalstatus;

DROP TYPE public.messagestatusenum;

DROP TYPE public.recipienttypeenum;

DROP TYPE public.recommendationfeedpriority;

DROP TYPE public.recommendationfeedstatus;

DROP TYPE public.runbookstatus;

DROP TYPE public.sendertypeenum;

DROP TYPE public.threadstatusenum;

DROP TYPE public.userrecommendationfeedinteractionstatus;

DROP TYPE public.windowtype;

DROP TYPE public.windowunit;

