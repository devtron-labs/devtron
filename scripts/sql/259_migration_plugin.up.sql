INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'GoLang-migrate','This plugin provides a seamless way to manage and execute database migrations, ensuring your database schema evolves safely and predictably with your application code. ','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/plugin-golang-migrate.png',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='GoLang-migrate'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/sh 
set -e
set -o pipefail
if [ -z $DB_HOST  ]; then
    echo "Please enter DB_HOST"
    exit 1
fi
if [ -z $DB_NAME  ]; then
    echo "Please enter DB_NAME"
    exit 1
fi

# Verify that you are on the correct branch and commit
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
CURRENT_COMMIT=$(git rev-parse HEAD)
if [ -z $SCRIPT_LOCATION  ]; then
    echo "we are running migration on current location"
    pwd
    SCRIPT_LOCATION="."
fi
echo "Current branch: $CURRENT_BRANCH"
echo "Current commit: $CURRENT_COMMIT"
echo "Migrating to version (0 for latest)"
echo $MIGRATE_TO_VERSION;

DB_CRED=""
if [ -n "$DB_USER"  ]; then
    DB_CRED="$DB_USER:$DB_PASSWORD@"
fi

if [ "$DB_TYPE" = "postgres" ]; then
    echo "migration for postgres"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database postgres://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database postgres://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
    docker run -v "$PWD:$PWD" "$MIGRATE_IMAGE" -path "$PWD/$SCRIPT_LOCATION" -database "postgres://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?$PARAM" version > migration-golang-current-version.txt 2>&1
elif [ "$DB_TYPE" = "mongodb" ]; then
    echo "migration for mongodb"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mongodb://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mongodb://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
    docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mongodb://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" version > migration-golang-current-version.txt 2>&1
elif [ "$DB_TYPE" = "mongodb+srv" ]; then
    echo "migration for mongodb"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mongodb+srv://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mongodb+srv://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
    docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mongodb+srv://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" version > migration-golang-current-version.txt 2>&1
elif [ "$DB_TYPE" = "mysql" ]; then
    echo "migration for mysql"
    DB="tcp($DB_HOST:$DB_PORT)"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mysql://$DB_CRED$DB/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mysql://$DB_CRED$DB/$DB_NAME?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
    docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database mysql://$DB_CRED$DB/$DB_NAME?"$PARAM" version > migration-golang-current-version.txt 2>&1
elif [ "$DB_TYPE" = "sqlserver" ]; then
    echo "migration for sqlserver"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database sqlserver://$DB_CRED$DB_HOST:$DB_PORT?"$PARAM" up;
    else
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database sqlserver://$DB_CRED$DB_HOST:$DB_PORT?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
    docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database sqlserver://$DB_CRED$DB_HOST:$DB_PORT?"$PARAM" version > migration-golang-current-version.txt 2>&1
else
    echo "no database matched"
fi
$POST_COMMAND
export POST_MIGRATION_VERION=$(cat migration-golang-current-version.txt)
if [ -z $POST_MIGRATION_VERION  ]; then
    POST_MIGRATION_VERION="0"
fi
echo "migration completed"$$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);






INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='GoLang-migrate'),'Step 1','Step 1 - GoLang-migrate','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO plugin_step_variable (id,plugin_step_id,name,format, description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'DB_TYPE','STRING','Currently this plugin support postgres,mongodb,mongodb+srv,mysql,sqlserver.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'DB_HOST','STRING','The hostname endpoint or IP address of the database server.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'DB_PORT','STRING','The port number on which the database server is listening.','t','f',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'DB_NAME','STRING','The name of the specific database instance you want to connect to.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'DB_USER','STRING','The username required to authenticate to the database.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'DB_PASSWORD','STRING','The password required to authenticate to the database.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'SCRIPT_LOCATION','STRING','sql files location in git repo','t','t',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'MIGRATE_IMAGE','STRING','Docker image of golang-migrate default:migrate/migrate','t','t','migrate/migrate',null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'MIGRATE_TO_VERSION','STRING','migrate to which version of sql script (default: 0 is for all files in directory)','t','f','0',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'PARAM','STRING','Additional connection parameters (optional), typically specified as key-value pairs. example: `sslmode=disable`', 't','t',-1,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'POST_COMMAND','STRING','post commands that runs at the end of script','t','t',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GoLang-migrate' and ps."index"=1 and ps.deleted=false),'POST_MIGRATION_VERION','STRING','migration version after running the SQL files', 't','t',-1,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
