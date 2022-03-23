UPDATE chart_ref SET is_default=false;
INSERT INTO "public"."chart_ref" ("location", "version", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "name") VALUES
('cronjob-chart_1-3-0', '1.3.0', 'f', 't', 'now()', 1, 'now()', 1, 'Cron Job & Job');

UPDATE chart_ref SET name = replace(name, 'Cron Job & Job', 'CronJob & Job');