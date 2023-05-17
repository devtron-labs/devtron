DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='4.18.0' and name is null);

DELETE FROM "public"."chart_ref" WHERE ("location" = 'statefulset-chart_4-18-0' AND "version" = '4.18.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-17-0' AND "version" = '4.17.0';