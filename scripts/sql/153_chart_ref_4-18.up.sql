INSERT INTO "public"."chart_ref" ("location", "version","deployment_strategy_path", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
    ('reference-chart_4-18-0', '4.18.0','pipeline-values.yaml', 'f', 't', 'now()', 1, 'now()', 1);


INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by","default")
VALUES (1,(select id from chart_ref where version='4.18.0' and name is null), true, now(), 1, now(), 1,true),
(2,(select id from chart_ref where version='4.18.0' and name is null), true, now(), 1, now(), 1,false),
(3,(select id from chart_ref where version='4.18.0' and name is null), true, now(), 1, now(), 1,false),
(4,(select id from chart_ref where version='4.18.0' and name is null), true, now(), 1, now(), 1,false);
