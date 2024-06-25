INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Go-migrate','This plugin is used to Cosign to sign docker images.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/migrate-logo.png',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='Go-migrate'), 0,'now()',1,'now()',1);

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

apk add git

if [ -z $REPO_URL  ]; then
  echo "running on current directory"
else
    cd .. 
    mkdir migration
    cd migration
    echo "git cloning"
    git clone $REPO_URL .
fi




if [ -z $BRANCH  ]; then
    echo "we are on main commit"
else
    echo "git checkout on"
    echo $BRANCH
    git checkout $BRANCH
fi

if [ -z $COMMIT_HASH  ]; then
    echo "we are on latest commit"
else
    echo "git checkout on"
    echo $COMMIT_HASH
    # Check out the specific commit hash
    git checkout $COMMIT_HASH
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
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database postgres://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database postgres://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
elif [ "$DB_TYPE" = "mongodb" ]; then
    echo "migration for mongodb"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database mongodb://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database mongodb://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
elif [ "$DB_TYPE" = "mongodb+srv" ]; then
    echo "migration for mongodb"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database mongodb+srv://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database mongodb+srv://$DB_CRED$DB_HOST:$DB_PORT/$DB_NAME?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
elif [ "$DB_TYPE" = "mysql" ]; then
    echo "migration for mysql"
    DB="tcp($DB_HOST:$DB_PORT)"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database mysql://$DB_CRED$DB/$DB_NAME?"$PARAM" up;
    else
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database mysql://$DB_CRED$DB_HOST:$DB?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
elif [ "$DB_TYPE" = "sqlserver" ]; then
    echo "migration for sqlserver"
    if [ $MIGRATE_TO_VERSION -eq "0" ]; then
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database sqlserver://$DB_CRED$DB_HOST:$DB_PORT?"$PARAM" up;
    else
        docker run -v $PWD:$PWD migrate/migrate   -path $PWD/$SCRIPT_LOCATION -database sqlserver://$DB_CRED$DB_HOST:$DB_PORT?"$PARAM" goto $MIGRATE_TO_VERSION;
    fi
else
    echo "no database matched"
fi
$POSTCOMMAND
if [ -z $REPO_URL  ]; then
  echo "current directory"
  pwd
else
    cd ..
    cd devtroncd
    pwd
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
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Go-migrate'),'Step 1','Step 1 - Go-migrate','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO plugin_step_variable (id,plugin_step_id,name,format, description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'DB_TYPE','STRING','database type (postgres,sqlserver,mysql etc)','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'DB_HOST','STRING','Database service url.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'DB_NAME','STRING','database name','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'DB_USER','STRING','database user name','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'DB_PASSWORD','STRING','database password','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'DB_PORT','STRING',' databse port','t','f',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'REPO_URL','STRING','git repo url (used in git clone ___)', 't','t',-1,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'BRANCH','STRING','branch name','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'COMMIT_HASH','STRING','commit hash','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'SCRIPT_LOCATION','STRING','sql files location in git repo','t','t',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'MIGRATE_TO_VERSION','STRING','migrate to which version of sql script (0 is for all files in directory)','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'PARAM','STRING','extra params that runs with db queries', 't','t',-1,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Go-migrate' and ps."index"=1 and ps.deleted=false),'POSTCOMMAND','STRING','post commands that runs at the end of script','t','t',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1);
