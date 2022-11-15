UPDATE chart_ref_metadata SET "chart_name" = replace("chart_name", 'CronJob & Job', 'Job & CronJob');
UPDATE chart_ref SET "name" = replace("name", 'CronJob & Job', 'Job & CronJob');