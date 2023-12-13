INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'CloudanixImagescan v1.0.0','Cloudanix Image Scanner is a cutting-edge solution designed for efficient and thorough scanning of images in cloud environments.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/cloudanix-logo.png',false,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),E'
echo "Initiating the scanning process..."  
docker run $ScannerImage image $ImageToScan --format table --api-endpoint $EndPoint --authz-token $AuthToken --identifier $AccountId --no-progress $ExtraArgs > result.txt
ScanningResult=$?
if [ $ScanningResult -eq 0 ]; then
    echo "Scan completed successfully"
else
    echo "Scan failed"
fi
Vulnerability=$(sed -n "4p" result.txt)
LowVulnerabilities=$(echo "$Vulnerability" | grep -o "LOW: [0-9]*" | sed "s/LOW: //")
MediumVulnerabilities=$(echo "$Vulnerability" | grep -o "MEDIUM: [0-9]*" | sed "s/MEDIUM: //")
HighVulnerabilities=$(echo "$Vulnerability" | grep -o "HIGH: [0-9]*" | sed "s/HIGH: //")
CriticalVulnerabilities=$(echo "$Vulnerability" | grep -o "CRITICAL: [0-9]*" | sed "s/CRITICAL: //")
cat result.txt 
','SHELL','f','now()',1,'now()',1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='CloudanixImagescan v1.0.0'),'Step 1','Step 1 - CloudanixImagescan v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'ScannerImage','STRING','Image that will be used as the basis for the scanning process in the Cloudanix vulnerability scanner default-:cloudanix/image-scanner:v0.0.1.','t','t','cloudanix/image-scanner:v0.0.1',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'ImageToScan','STRING','The image that need to scan.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'EndPoint','STRING','Cloudanix api endpoint for scanning.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'AuthToken','STRING','Cloudanix Authorization token for authentication.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'AccountId','STRING','Account ID associated with Cloudanix','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'AdditionalParams','STRING','Provide additional parameter like e.g-: --help, --ignore-unfixed .','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'ScanningResult','STRING','Stores the exit status of the vulnerability scanning eg-: 0/1.','t','t',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'LowVulnerabilities','STRING','Stores the count of low  vulnerability.','t','t',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'MediumVulnerabilities','STRING','Stores the count of medium vulnerability.','t','t',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'HighVulnerabilities','STRING','Stores the count of high vulnerability.','t','t',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CloudanixImagescan v1.0.0' and ps."index"=1 and ps.deleted=false),'CriticalVulnerabilities','STRING','Stores the count of critical vulnerability.','t','t',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);





INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='CloudanixImagescan v1.0.0'), 0,'now()',1,'now()',1);