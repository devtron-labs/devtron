INSERT INTO "public"."chart_ref" ("name","location", "version", "deployment_strategy_path","is_default", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
     ('GPU-Workload','gpu-workload-4-21-0', '4.21.0','pipeline-values.yaml','f', 't', 'now()', 1, 'now()', 1);

INSERT INTO global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by","default")
VALUES (1,(select id from chart_ref where version='4.21.0' and name='GPU-Workload'), true, now(), 1, now(), 1,true),
(4,(select id from chart_ref where version='4.21.0' and name='GPU-Workload'), true, now(), 1, now(), 1,false);

INSERT INTO chart_ref_metadata("chart_name","chart_description") VALUES 
('GPU-Workload','GPU Workload Charts enable the deployment of GPU workloads on Kubernetes Clusters.');