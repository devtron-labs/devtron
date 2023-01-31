UPDATE chart_ref SET is_default=false;
INSERT INTO "public"."chart_ref" ("location", "version","deployment_strategy_path", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
    ('reference-chart_4-17-0', '4.17.0','pipeline-values.yaml', 't', 't', 'now()', 1, 'now()', 1);


INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by")
VALUES (1,(select id from chart_ref where version='4.17.0' and name is null), true, now(), 1, now(), 1),
(2,(select id from chart_ref where version='4.17.0' and name is null), true, now(), 1, now(), 1),
(3,(select id from chart_ref where version='4.17.0' and name is null), true, now(), 1, now(), 1),
(4,(select id from chart_ref where version='4.17.0' and name is null), true, now(), 1, now(), 1);