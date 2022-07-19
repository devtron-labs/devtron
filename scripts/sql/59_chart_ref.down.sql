DELETE FROM "public"."chart_ref" WHERE ("location" = 'reference-chart_4-13-0' AND "version" = '4.13.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-12-0' AND "version" = '4.12.0';
