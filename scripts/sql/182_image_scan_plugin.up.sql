INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Vulnerability Scanning','Scan a image','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/ic-plugin-vulnerability-scan.png','f', 'now()', 1, 'now()', 1);

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Vulnerability Scanning'),0,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script", "type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),
        '#!/bin/sh
        echo "IMAGE SCAN"
        curl -X POST $IMAGE_SCANNER_ENDPOINT/scanner/image -H "Content-Type: application/json" -d "{\"image\": \"$DEST\", \"imageDigest\": \"$DIGEST\", \"userId\":
 $TRIGGERED_BY, \"dockerRegistryId\": \"$DOCKER_REGISTRY_ID\" }" >/dev/null 2>&1
        if [ $? != 0 ]
        then
        echo -e "\033[1m======== Vulnerability Scanning request failed ========"
        exit 1
        fi',
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1);




INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Vulnerability Scanning'),'Step 1','Step 1 - Vulnerability Scanning','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'DEST','STRING','image dest',false,true,'INPUT','GLOBAL',1 ,'DEST','f','now()', 1, 'now()', 1),
    (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'DIGEST','STRING','Image Digest',false,true,'INPUT','GLOBAL',1 ,'DIGEST','f','now()', 1, 'now()', 1),
    (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'PIPELINE_ID','STRING','Pipeline id',false,true,'INPUT','GLOBAL',1 ,'PIPELINE_ID','f','now()', 1, 'now()', 1),
    (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'TRIGGERED_BY','STRING','triggered by user',false,true,'INPUT','GLOBAL',1 ,'TRIGGERED_BY','f','now()', 1, 'now()', 1),
    (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'DOCKER_REGISTRY_ID','STRING','docker registry id',false,true,'INPUT','GLOBAL',1 ,'DOCKER_REGISTRY_ID','f','now()', 1, 'now()', 1),
    (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'IMAGE_SCANNER_ENDPOINT','STRING','image scanner endpoint',false,true,'INPUT','GLOBAL',1 ,'IMAGE_SCANNER_ENDPOINT','f','now()', 1, 'now()', 1);

