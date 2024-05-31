/*
 * Copyright (c) 2024. Devtron Inc.
 */

INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Mail Master v1.0.0','The Plugin is designed for sending bulk emails directly through your preferred SMTP server.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/MailMaster.png',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='Mail Master v1.0.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/sh 
set -eo pipefail 

#!/bin/bash

HoldTime=${BatchDelayTime:-1}
BatchSize=${BatchSize:-10}
SmtpPort=${SmtpPort:-587}
EMAIL_FILE_EXT="${EmailContentFile##*.}"
DIR=$(pwd)

CONFIG_FILE_ARGS=""
if [ -n "$RecipientConfigFile" ]; then
    CONFIG_FILE_ARGS="-e RecipientConfigFile=/app/config.json -v $DIR/$RecipientConfigFile:/app/config.json"
fi

docker run \
  -e SmtpServer="$SmtpServer" \
  -e SmtpPort="$SmtpPort" \
  -e SmtpUsername="$SmtpUsername" \
  -e SmtpPassword="$SmtpPassword" \
  -e SenderEmail="$SenderEmail" \
  -e Subject="$EmailSubject" \
  -e EmailContentFile="/app/email_content.$EMAIL_FILE_EXT" \
  -e SenderName="$SenderName" \
  -e RecipientsGroupName="$RecipientsGroupName" \
  -e RecipientsSubGroupName="$RecipientsSubGroupName" \
  -e Recipients="$Recipients" \
  -e BatchSize="$BatchSize" \
  -e HoldTime="$BatchDelayTime" \
  -v "$DIR/$EmailContentFile:/app/email_content.$EMAIL_FILE_EXT" \
    $CONFIG_FILE_ARGS \
  irawal007/mailmaster:v1.0
  
  $$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Mail Master v1.0.0'),'Step 1','Step 1 - Mail Master v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'SmtpServer','STRING','The Hostname of SMTP Server','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'SmtpUsername','STRING','The Username for the SMTP connection.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'SmtpPassword','STRING','The Password for the SMTP connection.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'SenderEmail','STRING','Sender Email Address.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'EmailContentFile','STRING','Enter the path to file whose contents is to be send in Email.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'SmtpPort','NUMBER','Port Number to use for SMTP connection.Defaults to 587 if not set','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'RecipientConfigFile','STRING','Enter the path to config.json file which contains the list of Recipients','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'EmailSubject','STRING','Subject of the email. Can either be specified here or in the first line of EmailContentFile after "Subejct:". ','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'BatchSize','NUMBER','Number of Emails to be sent per Batch. Defaults to 10 if not set','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'BatchDelayTime','NUMBER','Time to wait (in seconds) before scheduling the next batch','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'SenderName','STRING','Name of Email Sender','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'RecipientsGroupName','STRING','The Group id for selecting recipients in the config file.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'RecipientsSubGroupName','STRING','The SubGroup id for selecting recipients in the JSON configuration file.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Mail Master v1.0.0' and ps."index"=1 and ps.deleted=false),'Recipients','STRING','The emails of recipients separated by "," if user has not provided config file.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
