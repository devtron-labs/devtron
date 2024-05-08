INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Codacy','Codacy is an automated code analysis/quality tool that helps developers ship better software, faster.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/codacy-plugin-icon.png',false,'now()',1,'now()',1);

INSERT INTO plugin_pipeline_script (id,script,type,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_pipeline_script'),E'if [[ ! -z "$CodacyApiToken" ]]
then
  CODACY_API_TOKEN=$CodacyApiToken
fi
data_raw="{\\\"branchName\\\":\\\"$Branch\\\",\\\"categories\\\":[\\\"Security\\\"],\\\"levels\\\":[\\\"Error\\\"]}"
raw_url="curl -X POST \\\"$CodacyEndpoint/api/v3/analysis/organizations/$GitProvider/$Organisation/repositories/$RepoName/issues/search\\\" -H \\\"Content-Type:application/json\\\" -H \\\"api-token:$CODACY_API_TOKEN\\\" --data-raw \'$data_raw\'"
result=`eval $raw_url`
echo $result
export NUMBER_OF_ISSUES=$(echo $result | jq -r ".data | length")
echo "***********number of issue***********"
echo "Number of issues are: $NUMBER_OF_ISSUES"
echo "***********number of issue***********"
if [ "$NUMBER_OF_ISSUES" -gt "0" ]
then
    echo "This code has critical Vulnerabilities . Visit https://app.codacy.com/gh/delhivery/$REPO/issues  for more Info"
else
    exit 0
fi','SHELL',false,'now()',1,'now()',1);

INSERT INTO plugin_step (id,plugin_id,name,description,index,step_type,script_id,ref_plugin_id,output_directory_path,dependent_on_step,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_step'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'Step 1','Step 1 for Codacy',1,'INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),null,null,null,false,'now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Codacy' and ps."index"=1 and ps.deleted=false),'CodacyEndpoint','STRING','Api Endpoint for Codacy','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Codacy' and ps."index"=1 and ps.deleted=false),'GitProvider','STRING','Git provider for the scan','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Codacy' and ps."index"=1 and ps.deleted=false),'CodacyApiToken','STRING','If provided, this token will be used. If not provided it will be picked from global secret(CODACY_API_TOKEN)','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Codacy' and ps."index"=1 and ps.deleted=false),'Organisation','STRING','Org for the Codacy','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Codacy' and ps."index"=1 and ps.deleted=false),'RepoName','STRING','Repo name','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Codacy' and ps."index"=1 and ps.deleted=false),'Branch','STRING','Branch name ','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Codacy' and ps."index"=1 and ps.deleted=false),'NUMBER_OF_ISSUES','STRING','Number of issue in code source','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


INSERT INTO plugin_tag (id,name,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_tag'),'Code Review',false,'now()',1,'now()',1);

INSERT INTO plugin_tag_relation (id,tag_id,plugin_id,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_tag_relation'),2,(SELECT id FROM plugin_metadata WHERE name='Codacy'),'now()',1,'now()',1),
(nextval('id_seq_plugin_tag_relation'),3,(SELECT id FROM plugin_metadata WHERE name='Codacy'),'now()',1,'now()',1),
(nextval('id_seq_plugin_tag_relation'),(SeLECT id FROM plugin_tag WHERE name='Code Review'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'now()',1,'now()',1);
