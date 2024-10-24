UPDATE plugin_metadata SET is_latest = false WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'Devtron CI Trigger v1.0.0' and is_latest= true);

INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'Devtron CI Trigger','Triggers the CI pipeline of Devtron Application','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='devtron-ci-trigger-v1-0-0'),'1.1.0', false, true);

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='1.1.0' and name='Devtron CI Trigger' and deleted= false),0,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/sh
    docker run -e DevtronApiToken=$DevtronApiToken -e DevtronEndpoint=$DevtronEndpoint -e DevtronApp=$DevtronApp -e CiPipeline=$CiPipeline -e DevtronEnv=$DevtronEnv -e GitCommitHash=$GitCommitHash -e Timeout=$Timeout -e IgnoreCache=$IgnoreCache --name devtron-ci-trigger quay.io/devtron/devtron-utils:ci-trigger-plugin-v1.1.0
    exit_code=$?
    if [ $ExitOnFail == true ];then
        if [ $exit_code == 2 ];then
            echo "The triggered build has been failed terminating the current process."
            exit $exit_code
        fi
    fi
    if [ $exit_code -ne 0 ] && [ $exit_code -ne 2 ] ; then
          echo "The Docker container exited with code $exit_code. Terminating current process."
          exit $exit_code  
    fi','SHELL','f','now()',1,'now()',1);


INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Devtron CI Trigger'),'Step 1','Runnig the plugin','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format, description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'DevtronApiToken','STRING','Enter Devtron API Token with required permissions.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'DevtronEndpoint','STRING','Enter the URL of Devtron Dashboard for.eg (https://devtron.example.com).','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'DevtronApp','STRING','Enter the name or ID of the Application whose build is to be triggered.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'DevtronEnv','STRING','Enter the name or ID of the Environment to which the CI is attached. Required if CiPipeline is not given.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'CiPipeline','STRING','Enter the name or ID of the CI pipeline to be triggered. Required if DevtronEnv is not given.','t','t',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'GitCommitHash','STRING','Enter the commit hash from which the build is to be triggered. If not given then will pick the latest.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'Timeout','NUMBER','Enter the maximum time to wait for the build status.', 't','t',-1,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger' and ps."index"=1 and ps.deleted=false),'IgnoreCache','STRING','Set true if you want to ignore cache for the build.', 't','t','false',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
