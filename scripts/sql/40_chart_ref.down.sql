DELETE FROM "public"."chart_ref" WHERE ("location" = 'cronjob-chart_1-3-0' AND "version" = '1.3.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'cronjob-chart_1-2-0' AND "version" = '1.2.0';

UPDATE chart_ref SET name = replace(name, 'CronJob & Job', 'Cron Job & Job');