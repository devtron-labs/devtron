UPDATE chart_ref_metadata SET "chart_name" = replace("chart_name", 'Job & CronJob', 'CronJob & Job');
UPDATE chart_ref SET "name" = 'CronJob & Job' WHERE "name" = 'Job & CronJob' and "user_uploaded" = false;
