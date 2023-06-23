INSERT INTO "public"."chart_ref" ("location", "version","deployment_strategy_path", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by","name") VALUES
    ('statefulset-chart_1-0-0', '1.0.0','pipeline-values.yaml', 'f', 't', 'now()', 1, 'now()', 1,'StatefulSet');

INSERT INTO global_strategy_metadata("id","name","key","description","deleted","created_on","created_by","updated_on","updated_by") VALUES 
(nextval('id_seq_global_strategy_metadata'),'ROLLINGUPDATE','rollingUpdate','Rolling update strategy for statefulset.',false,now(),1,now(),1),
(nextval('id_seq_global_strategy_metadata'),'ONDELETE','onDelete','OnDelete strategy for statefulset.',false,now(),1,now(),1);


INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES 
((select id from global_strategy_metadata where name='ROLLINGUPDATE') ,(select id from chart_ref where location='statefulset-chart_1-0-0'), true, now(), 1, now(), 1),
((select id from global_strategy_metadata where name='ONDELETE') ,(select id from chart_ref where location='statefulset-chart_1-0-0'), true, now(), 1, now(), 1);



INSERT INTO chart_ref_metadata("chart_name","chart_description") VALUES 
('StatefulSet','StatefulSet  is a controller object that manages the deployment and scaling of stateful applications while providing guarantees around the order of deployment and uniqueness of names for each pod.');