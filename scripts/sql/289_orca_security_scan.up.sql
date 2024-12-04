INSERT INTO "plugin_parent_metadata" ("id", "name","identifier", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_parent_metadata'), 'ORCA SECURITY SCAN','security-image-scan','Orca Cloud Security plugin for scanning container images and detecting vulnerabilities','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png','f', 'now()', 1, 'now()', 1);


UPDATE plugin_metadata SET is_latest = false WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'ORCA SECURITY SCAN' and is_latest= true);


INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'ORCA SECURITY SCAN','Plugin for scanning images and detecting vulnerabilities using Orca Cloud Security tool','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='security-image-scan'),'1.0.0', false, true);


INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='ORCA SECURITY SCAN' and deleted= false),3,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/bash
    
    echo "Installing Orca CLI..."
    curl -sfL "https://raw.githubusercontent.com/orcasecurity/orca-cli/55c2742bb04c5a11f25e061a8e715cf6bc281a67/install.sh" | bash

    if ! command -v orca-cli &> /dev/null; then
      echo "Error: Orca CLI installation failed"
      exit 1
    fi
   
   if [ -z "${ScanTimeout}" ]; then
      ScanTimeout="15m"
      echo "No timeout specified. Using default: 15 minutes."
   fi

    echo "Starting Orca image scan for ${ImageName}..."  
    orca-cli -p "${ProjectKey}" image scan "${ImageName}" --api-token "${OrcaSecurityApiToken}" --timeout "${ScanTimeout}" 
    echo "Orca image scan completed successfully"
    ',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'),(SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='ORCA SECURITY SCAN' and deleted= false),'Step 1','Step 1 - Triggering ORCA SECURITY SCAN','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by","default_value")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='ORCA SECURITY SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'ImageName','STRING','Provide the image that needs to be scanned',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1, null),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='ORCA SECURITY SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'ProjectKey','STRING','Name of the Project Key present in Orca Security platform',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1,null),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='ORCA SECURITY SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'OrcaSecurityApiToken','STRING','Orca API Token used for Authentication',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1, null),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='ORCA SECURITY SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'ScanTimeout','STRING','Maximum time allowed for the scan (e.g., ''10m'', ''1h'', default: ''15m'' )',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1, null);
