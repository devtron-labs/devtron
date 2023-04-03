UPDATE chart_ref SET is_default=false;
INSERT INTO "public"."chart_ref" ("location", "version","deployment_strategy_path", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
    ('statefulset-chart_1-0-0', '1.0.0','pipeline-values.yaml', 't', 't', 'now()', 1, 'now()', 1);

INSERT INTO global_strategy_metadata("id","name","description","deleted","created_on","created_by","updated_on","updated_by")
VALUES (nextval("id_seq_global_strategy_metadata"),"ONDELETE","OnDelete strategy for statefulset.",false,now(),1,now(),1);


INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by")
VALUES (1,(select id from chart_ref where location='statefulset-chart_1-0-0' and name is null), true, now(), 1, now(), 1),
VALUES ((select id from global_strategy_metadata where name="ONDELETE") ,(select id from chart_ref where location='statefulset-chart_1-0-0' and name is null), true, now(), 1, now(), 1);

