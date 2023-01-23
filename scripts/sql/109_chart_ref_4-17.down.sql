DELETE FROM "public"."chart_ref" WHERE ("location" = 'reference-chart_4-17-0' AND "version" = '4.17.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-16-0' AND "version" = '4.16.0';