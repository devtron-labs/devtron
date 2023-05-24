INSERT INTO "public"."chart_ref" ("name","location", "version", "deployment_strategy_path","is_default", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
     ('Deployment','deployment-chart_1-1-0', '1.1.0','pipeline-values.yaml','f', 't', 'now()', 1, 'now()', 1);

INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by")
VALUES (1,(select id from chart_ref where version='1.1.0' and name='Deployment'), true, now(), 1, now(), 1),
(4,(select id from chart_ref where version='1.1.0' and name='Deployment'), true, now(), 1, now(), 1);