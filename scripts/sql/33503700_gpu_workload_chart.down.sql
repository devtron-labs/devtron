DELETE FROM global_strategy_metadata_chart_ref_mapping WHERE chart_ref_id=(select id from chart_ref where version='4.21.0' and name='GPU-Workload');

DELETE FROM "public"."chart_ref" WHERE ("location" = 'gpu-workload-4-21-0' AND "version" = '4.21.0');