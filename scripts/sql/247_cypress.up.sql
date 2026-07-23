INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_metadata'),'Cypress v1.0.0' , 'The plugin enables users to test their application.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/cypress-logo-plugin.jpeg',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='Cypress v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash

# Check if the directory already exists
if [ ! -d "testing" ]; then
    # Clone the repository
    git clone $RepositoryUrl
    # Check if git clone was successful
    if [ $? -eq 0 ]; then
        echo "Repository cloned successfully."
    else
        echo "Failed to clone repository. Exiting..."
        exit 1
    fi
else
    echo "Directory \'testing\' already exists, skipping cloning."
fi


REPO_NAME=$(echo $RepositoryUrl | awk -F\'/\' \'{print $NF}\' | awk -F\'.\' \'{print $1}\')

cd $REPO_NAME

echo \'ARG CypressImageName=""

FROM $CypressImageName

ENTRYPOINT ["sh"]\' > Dockerfile

docker build --build-arg CypressImageName=$CypressImageName -t cypress:v1.0 . 
if [ $? -eq 0 ]; then
    echo "Docker build image successfully."
else
    echo "Failed to build docker image. Exiting..."
    exit 1
fi

echo "Now running the test-case project."
docker run -v $PWD:/app/ cypress:v1.0 -c "cd /app && cypress run --browser $BrowserName --spec $TestCasePath"
',

    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);


INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") 
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Cypress v1.0.0'),'Step 1','Cypress v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cypress v1.0.0' and ps."index"=1 and ps.deleted=false),'RepositoryUrl','STRING','Enter repository url of the testing project.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cypress v1.0.0' and ps."index"=1 and ps.deleted=false),'CypressImageName','STRING','Enter cypress image name where you want to test your application. default: cypress/included:12.12.0','t','t','cypress/included:12.12.0',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cypress v1.0.0' and ps."index"=1 and ps.deleted=false),'BrowserName','STRING','Enter browser name default: chrome','t','t','chrome',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cypress v1.0.0' and ps."index"=1 and ps.deleted=false),'TestCasePath','STRING','Enter the testcase path in the repository default: ./cypress/e2e/*.cy.js','t','t','./cypress/e2e/*.cy.js',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);



