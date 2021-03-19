DELETE FROM "public"."chart_ref" WHERE ("location" = 'reference-chart_3-10-0' AND "version" = '3.10.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_3-9-0' AND "version" = '3.9.0';