-- Dropping table user_deployment_request
DROP TABLE IF EXISTS "public"."user_deployment_request";

-- Dropping the sequence
DROP SEQUENCE IF EXISTS public.id_seq_user_deployment_request_sequence;

-- Delete priority deployment condition from attributes table
DELETE FROM "public"."attributes" WHERE key = 'priorityDeploymentCondition';