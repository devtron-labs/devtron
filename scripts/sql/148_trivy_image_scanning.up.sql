INSERT INTO public.scan_tool_metadata(name, version, server_base_url, result_descriptor_template, scan_target, active, deleted, created_on, created_by, updated_on, updated_by,tool_metadata) VALUES ('CLAIR','V2',null,null,'IMAGE',false,false,now()::timestamp,'1',now()::timestamp,'1',null),('CLAIR','V4',null,null,'IMAGE',false,false,now()::timestamp,'1',now()::timestamp,'1',null),('TRIVY','V1',null,'[{{$size1:= len .Results}}{{range $i1, $v1 := .Results}}{{ if  $v1.Vulnerabilities}}{{$size2:= len $v1.Vulnerabilities}}{{range $i2, $v2 := $v1.Vulnerabilities}}{{if and (eq $i1 (add $size1 -1)) (eq $i2 (add $size2 -1)) }}
{
"package": "{{$v2.PkgName}}",
"packageVersion": "{{$v2.InstalledVersion}}",
"fixedInVersion": "{{$v2.FixedVersion}}",
"severity": "{{$v2.Severity}}",
"name": "{{$v2.VulnerabilityID}}"
}{{else}}{
"package": "{{$v2.PkgName}}",
"packageVersion": "{{$v2.InstalledVersion}}",
"fixedInVersion": "{{$v2.FixedVersion}}",
"severity": "{{$v2.Severity}}",
"name": "{{$v2.VulnerabilityID}}"
},{{end}}{{end}}{{end}}{{end}}]' ,'IMAGE',false,false,now()::timestamp,'1',now()::timestamp,'1','{"timeout":"10m"}');

INSERT INTO public.scan_tool_step(scan_tool_id, index, step_execution_type, step_execution_sync, retry_count, execute_step_on_fail, execute_step_on_pass, render_input_data_from_step, http_input_payload, http_method_type, http_req_headers, http_query_params, cli_command, cli_output_type, deleted, created_on, created_by, updated_on, updated_by) VALUES (3,1,'CLI',true,1,-1,-1,-1,null,null,null,null,'trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}} --username {{.USERNAME}} --password {{.PASSWORD}}', 'STATIC',false,now()::timestamp,'1',now()::timestamp,'1'),(3,2,'CLI',true,1,-1,-1 ,-1,null,null,null,null,'(export AWS_ACCESS_KEY_ID={{.AWS_ACCESS_KEY_ID}} AWS_SECRET_ACCESS_KEY={{.AWS_SECRET_ACCESS_KEY}} AWS_DEFAULT_REGION={{.AWS_DEFAULT_REGION}}; trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}})','STATIC', false,now()::timestamp,'1',now()::timestamp,'1'), (3,3,'CLI',true,1,-1,-1,-1,null,null,null,null,'GOOGLE_APPLICATION_CREDENTIALS="{{.FILE_PATH}}/credentials.json" trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}}', 'STATIC',false,now()::timestamp,'1',now()::timestamp,'1'),(3,4,'CLI',true,1,-1,3,-1,null,null,null,null,'echo {{.PASSWORD}} > {{.FILE_PATH}}/credentials.json','STATIC',false,now()::timestamp,'1',now()::timestamp,'1');
INSERT INTO public.registry_index_mapping(scan_tool_id, registry_type, starting_index) VALUES (3,'ecr',2), (3,'other',1),(3,'artifact-registry',4),( 3,'docker-hub',1), (3,'gcr',4), (3,'acr',1),(3,'quay',1);

ALTER TABLE public.module
    ADD "module_type" varchar(30),
    ADD "enabled" bool;

UPDATE public.module SET module_type = 'security',enabled=true where name='security.clair';
UPDATE public.module SET enabled=true where status='installed';
UPDATE public.module SET enabled=false where status!='installed';
UPDATE scan_tool_metadata SET active='true' where id in (SELECT stmd.id FROM module m, scan_tool_metadata stmd where m.name ='security.clair' and m.enabled =true and stmd.name ='CLAIR' and stmd.version='V4');