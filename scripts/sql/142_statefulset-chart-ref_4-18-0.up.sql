INSERT INTO "public"."chart_ref" ("location", "version","deployment_strategy_path", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by","name") VALUES
    ('statefulset-chart_5-0-0', '5.0.0','pipeline-values.yaml', 'f', 't', 'now()', 1, 'now()', 1,'StatefulSet');

INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id","chart_ref_id", "active","default","created_on", "created_by", "updated_on", "updated_by") VALUES 
((select id from global_strategy_metadata where name='ROLLINGUPDATE') ,(select id from chart_ref where location='statefulset-chart_5-0-0'), true,true,now(), 1, now(), 1),
((select id from global_strategy_metadata where name='ONDELETE') ,(select id from chart_ref where location='statefulset-chart_5-0-0'), true, false,now(), 1, now(), 1);