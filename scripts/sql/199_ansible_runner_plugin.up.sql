INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Ansible Runner-v1.0.0','Ansible Runner allows running the Ansible Playbooks using the ansible-runner tool.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/ansible-runner.jpg',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='Ansible Runner-v1.0.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/bash
set -e

pip install ansible
pip install ansible-runner

# Set the project directory
project_dir=$ProjectDirectory
runner_dir=$RunnerDirectory
args=$RunnerArguments
cd "/devtroncd/${runner_dir}/${project_dir}"
pwd
# Install Ansible requirements
if [ -f requirements.txt ]; then
pip3 install --user -r requirements.txt
fi

if [ -f requirements.yml ]; then
ansible-galaxy role install -vv -r requirements.yml
ansible-galaxy collection install -vv -r requirements.yml
fi
bash -c "ansible-runner run $args $ProjectDirectory"$$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Ansible Runner-v1.0.0'),'Step 1','Step 1 - Ansible Runner-v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Ansible Runner-v1.0.0' and ps."index"=1 and ps.deleted=false),'ProjectDirectory','STRING','Provide the project directory','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Ansible Runner-v1.0.0' and ps."index"=1 and ps.deleted=false),'RunnerDirectory','STRING','Provide the Runner Diectory','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Ansible Runner-v1.0.0' and ps."index"=1 and ps.deleted=false),'RunnerArguments','STRING','Provide the Runner Arguments','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
