UPDATE chart_ref SET is_default=false;
INSERT INTO "public"."chart_ref" ("location", "version", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
    ('reference-chart_4-16-0', '4.16.0', 't', 't', 'now()', 1, 'now()', 1);
