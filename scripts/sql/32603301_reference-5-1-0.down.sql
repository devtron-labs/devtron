DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='5.1.0' and name is null);

DELETE FROM "public"."chart_ref" WHERE ("location" = 'reference-chart_5-1-0' AND "version" = '5.1.0');