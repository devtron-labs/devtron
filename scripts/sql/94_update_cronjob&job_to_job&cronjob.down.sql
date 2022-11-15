UPDATE chart_ref_metadata SET "chart_name" = replace("chart_name", 'Job & CronJob', 'CronJob & Job');
UPDATE chart_ref SET "name" = replace("name", 'Job & CronJob', 'CronJob & Job');