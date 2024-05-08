DELETE FROM "public"."chart_ref" WHERE ("location" = 'reference-chart_4-15-0' AND "version" = '4.15.0');

UPDATE "public"."chart_ref" SET "is_default" = 't' WHERE "location" = 'reference-chart_4-14-0' AND "version" = '4.14.0';