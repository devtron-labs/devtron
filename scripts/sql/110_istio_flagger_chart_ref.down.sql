DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='1.1.0' and name='Deployment');

DELETE FROM "public"."chart_ref" WHERE ("location" = 'deployment-chart_1-1-0' AND "version" = '1.1.0');