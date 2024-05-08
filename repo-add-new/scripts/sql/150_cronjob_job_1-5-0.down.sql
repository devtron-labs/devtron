DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='1.5.0' and name ='Job & CronJob');

DELETE FROM "public"."chart_ref" WHERE ("location" = 'cronjob-chart_1-5-0' AND "version" = '1.5.0');
