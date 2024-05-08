DELETE FROM "public"."chart_ref" WHERE ("location" = 'reference-chart_4-10-0' AND "version" = '4.10.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_3-12-0' AND "version" = '3.12.0';
