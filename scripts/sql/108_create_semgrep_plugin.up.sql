INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Semgrep','Semgrep is a fast, open source, static analysis engine for finding bugs, detecting dependency vulnerabilities, and enforcing code standards.','SHARED','link_to_icon','f', 'now()', 1, 'now()', 1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), 2, 7,'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), 3, 7,'now()', 1, 'now()', 1);



INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    '#!/bin/sh
set -eo pipefail
PathToCodeDir=/devtroncd$CheckoutPath
chmod 741 $PathToCodeDir
apk add py3-pip
pip install pip==21.3.1
pip install semgrep
cp $PathToCodeDir $SemgrepProjectName -r
cd $SemgrepProjectName
export SEMGREP_APP_TOKEN=$SemgrepAppToken
semgrep ci
cd $PathToCodeDir
',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), 7,'Step 1','Step 1 - Dependency Track for Semgrep','1','INLINE','14','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), 7, 'SemgrepAppToken','STRING','App token for Semgrep account',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), 7, 'CheckoutPath','STRING','git repo checkout path',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), 7, 'SemgrepProjectName','STRING','Semgrep dashboard project name',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);


