UPDATE chart_ref SET is_default=false;
INSERT INTO "public"."chart_ref" ("location", "version", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "name","deployment_strategy_path") VALUES
('cronjob-chart_1-5-0', '1.5.0', 'f', 't', 'now()', 1, 'now()', 1, 'Job & CronJob','pipeline-values.yaml');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-17-0' AND "version" = '4.17.0';

INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id","chart_ref_id", "active","default","created_on", "created_by", "updated_on", "updated_by") VALUES 
((select id from global_strategy_metadata where name='ROLLING') ,(select id from chart_ref where location='cronjob-chart_1-5-0'), true,true,now(), 1, now(), 1);
