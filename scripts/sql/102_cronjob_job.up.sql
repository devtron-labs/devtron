UPDATE chart_ref SET is_default=false;
INSERT INTO "public"."chart_ref" ("location", "version", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "name") VALUES
('cronjob-chart_1-4-0', '1.4.0', 'f', 't', 'now()', 1, 'now()', 1, 'Job & CronJob');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-16-0' AND "version" = '4.16.0';
