INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Codacy','Enhance Your Workflow with Continuous Code Quality & Code Security','PRESET','https://avatars.githubusercontent.com/u/1834093?s=200&v=4',false,'now()',1,'now()',1);

INSERT INTO plugin_pipeline_script (id,script,type,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_pipeline_script'),E'if [[ "$PickTokenFromGlobalConfigs" == false && ! -z "$CODACY_API_TOKEN" ]]
then
  ApiToken=$CODACY_API_TOKEN
fi
data_raw="{\\\"branchName\\\":\\\"$Branch\\\",\\\"categories\\\":[\\\"Security\\\"],\\\"levels\\\":[\\\"Error\\\"]}"
raw_url="curl -X POST \\\"$CodacyEndpoint/api/v3/analysis/organizations/$GitProvider/$Organisation/repositories/$RepoName/issues/search\\\" -H \\\"Content-Type:application/json\\\" -H \\\"api-token:$ApiToken\\\" --data-raw \'$data_raw\'"
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
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'CodacyEndpoint','STRING','Api Endpoint for Codacy','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'GitProvider','STRING','Git provider for the scan','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'PickTokenFromGlobalConfigs','BOOL','Whether to use global token or not','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'CODACY_API_TOKEN','STRING','API Toekn for the Codacy','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'Organisation','STRING','Org for the Codacy','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'RepoName','STRING','Repo name','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'Branch','STRING','Branch name ','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'NUMBER_OF_ISSUES','STRING','Number of issue in code source','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


INSERT INTO plugin_tag (id,name,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_tag'),'Code Review',false,'now()',1,'now()',1);

INSERT INTO plugin_tag_relation (id,tag_id,plugin_id,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_tag_relation'),2,(SELECT id FROM plugin_metadata WHERE name='Codacy'),'now()',1,'now()',1),
(nextval('id_seq_plugin_tag_relation'),3,(SELECT id FROM plugin_metadata WHERE name='Codacy'),'now()',1,'now()',1),
(nextval('id_seq_plugin_tag_relation'),(SeLECT id FROM plugin_tag WHERE name='Code Review'),(SELECT id FROM plugin_metadata WHERE name='Codacy'),'now()',1,'now()',1);
