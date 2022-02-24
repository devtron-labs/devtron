DELETE FROM "public"."chart_ref" WHERE ("location" = 'cronjob-chart_1-3-0' AND "version" = '1.3.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-11-0' AND "version" = '4.11.0';
