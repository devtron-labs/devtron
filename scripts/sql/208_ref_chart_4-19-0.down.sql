DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='4.19.0' and name='Deployment');

DELETE FROM "public"."chart_ref" WHERE ("location" = 'deployment-chart_4-19-0' AND "version" = '4.19.0');

DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='4.19.0' and name is null);

DELETE FROM "public"."chart_ref" WHERE ("location" = 'reference-chart_4-19-0' AND "version" = '4.19.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-18-0' AND "version" = '4.18.0';

DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='4.19.0' and location='statefulset-chart_4-19-0' and name is null);

DELETE FROM "public"."chart_ref" WHERE ("location" = 'statefulset-chart_4-19-0' AND "version" = '4.19.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'statefulset-chart_4-18-0' AND "version" = '4.18.0';