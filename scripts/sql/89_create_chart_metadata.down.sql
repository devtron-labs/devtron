DROP TABLE "public"."chart_ref_metadata" CASCADE;

----Rows Deletion---
DELETE FROM "public"."chart_ref_metadata" WHERE ("chart_name"= 'Rollout Deployment')
DELETE FROM "public"."chart_ref_metadata" WHERE ("chart_name"= 'CronJob & Job')
DELETE FROM "public"."chart_ref_metadata" WHERE ("chart_name"= 'Knative')
DELETE FROM "public"."chart_ref_metadata" WHERE ("chart_name"= 'Deployment')
