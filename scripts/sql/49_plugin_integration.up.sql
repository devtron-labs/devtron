CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_metadata;

-- Table Definition
CREATE TABLE "public"."plugin_metadata"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_metadata'::regclass),
    "name"                        text,
    "description"                 text,
    "type"                        varchar(255),  -- SHARED, PRESET etc
    "icon"                        text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_tag;

-- Table Definition
CREATE TABLE "public"."plugin_tag"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_tag'::regclass),
    "name"                        varchar(255),
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_tag_relation;

-- Table Definition
CREATE TABLE "public"."plugin_tag_relation"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_tag_relation'::regclass),
    "tag_id"                      integer,
    "plugin_id"                   integer,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "plugin_tag_relation_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "public"."plugin_tag" ("id"),
    CONSTRAINT "plugin_tag_relation_plugin_id_fkey" FOREIGN KEY ("plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_pipeline_script;

-- Table Definition
CREATE TABLE "public"."plugin_pipeline_script"
(
    "id"                           integer NOT NULL DEFAULT nextval('id_seq_plugin_pipeline_script'::regclass),
    "script"                       text,
    "type"                         varchar(255),   -- SHELL, DOCKERFILE, CONTAINER_IMAGE etc
    "store_script_at"              text,
    "dockerfile_exists"            bool,
    "mount_path"                   text,
    "mount_code_to_container"      bool,
    "mount_code_to_container_path" text,
    "mount_directory_from_host"    bool,
    "container_image_path"         text,
    "image_pull_secret_type"       varchar(255),   -- CONTAINER_REGISTRY or SECRET_PATH
    "image_pull_secret"            text,
    "deleted"                      bool,
    "created_on"                   timestamptz,
    "created_by"                   int4,
    "updated_on"                   timestamptz,
    "updated_by"                   int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_script_path_arg_port_mapping;

-- Table Definition
CREATE TABLE "public"."script_path_arg_port_mapping"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_script_path_arg_port_mapping'::regclass),
    "type_of_mapping"             varchar(255),      -- FILE_PATH, DOCKER_ARG, PORT
    "file_path_on_disk"           text,
    "file_path_on_container"      text,
    "command"                     text,
    "args"                        text[],
    "port_on_local"               integer,
    "port_on_container"           integer,
    "script_id"                   integer,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "script_path_arg_port_mapping_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step;

-- Table Definition
CREATE TABLE "public"."plugin_step"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_step'::regclass),
    "plugin_id"                   integer,        -- id of plugin - parent of this step
    "name"                        varchar(255),
    "description"                 text,
    "index"                       integer,
    "step_type"                   varchar(255),   -- INLINE or REF_PLUGIN
    "script_id"                   integer,
    "ref_plugin_id"               integer,        -- id of plugin used as reference
    "output_directory_path"       text[],
    "dependent_on_step"           text,           -- name of step this step is dependent on
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "plugin_step_plugin_id_fkey" FOREIGN KEY ("plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    CONSTRAINT "plugin_step_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    CONSTRAINT "plugin_step_ref_plugin_id_fkey" FOREIGN KEY ("ref_plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step_variable;

-- Table Definition
CREATE TABLE "public"."plugin_step_variable"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_plugin_step_variable'::regclass),
    "plugin_step_id"                integer,
    "name"                          varchar(255),
    "format"                        varchar(255),
    "description"                   text,
    "is_exposed"                    bool,
    "allow_empty_value"             bool,
    "default_value"                 varchar(255),
    "value"                         varchar(255),
    "variable_type"                 varchar(255),   -- INPUT or OUTPUT
    "value_type"                    varchar(255),   -- NEW, FROM_PREVIOUS_STEP or GLOBAL
    "previous_step_index"           integer,
    "variable_step_index"           integer,
    "variable_step_index_in_plugin" integer,        -- will contain step index of variable in case of ref plugin
    "reference_variable_name"       text,
    "deleted"                       bool,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "plugin_step_variable_plugin_step_id_fkey" FOREIGN KEY ("plugin_step_id") REFERENCES "public"."plugin_step" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step_condition;

-- Table Definition
CREATE TABLE "public"."plugin_step_condition"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_step_condition'::regclass),
    "plugin_step_id"              integer,
    "condition_variable_id"       integer,      -- id of variable on which condition is written
    "condition_type"              varchar(255), -- SKIP, TRIGGER, SUCCESS or FAILURE
    "conditional_operator"        varchar(255),
    "conditional_value"           varchar(255),
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "plugin_step_condition_plugin_step_id_fkey" FOREIGN KEY ("plugin_step_id") REFERENCES "public"."plugin_step" ("id"),
    CONSTRAINT "plugin_step_condition_condition_variable_id_fkey" FOREIGN KEY ("condition_variable_id") REFERENCES "public"."plugin_step_variable" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage;

-- Table Definition
CREATE TABLE "public"."pipeline_stage"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage'::regclass),
    "name"                        text,
    "description"                 text,
    "type"                        varchar(255),  -- PRE_CI, POST_CI, PRE_CD, POST_CD etc
    "deleted"                     bool,
    "ci_pipeline_id"              integer,
    "cd_pipeline_id"              integer,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_ci_pipeline_id_fkey" FOREIGN KEY ("ci_pipeline_id") REFERENCES "public"."ci_pipeline" ("id"),
    CONSTRAINT "pipeline_stage_cd_pipeline_id_fkey" FOREIGN KEY ("cd_pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage_step'::regclass),
    "pipeline_stage_id"           integer,
    "name"                        varchar(255),
    "description"                 text,
    "index"                       integer,
    "step_type"                   varchar(255),   -- INLINE or REF_PLUGIN
    "script_id"                   integer,
    "ref_plugin_id"               integer,        -- id of plugin used as reference
    "output_directory_path"       text[],
    "dependent_on_step"           text,           -- name of step this step is dependent on
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_step_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    CONSTRAINT "pipeline_stage_step_ref_plugin_id_fkey" FOREIGN KEY ("ref_plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step_variable;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step_variable"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage_step_variable'::regclass),
    "pipeline_stage_step_id"        integer,
    "name"                          varchar(255),
    "format"                        varchar(255),
    "description"                   text,
    "is_exposed"                    bool,
    "allow_empty_value"             bool,
    "default_value"                 varchar(255),
    "value"                         varchar(255),
    "variable_type"                 varchar(255),   -- INPUT or OUTPUT
    "index"                         integer,
    "value_type"                    varchar(255),   -- NEW, FROM_PREVIOUS_STEP or GLOBAL
    "previous_step_index"           integer,
    "variable_step_index_in_plugin" integer,
    "reference_variable_name"       text,
    "reference_variable_stage"      text,
    "deleted"                       bool,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "pipeline_stage_step_variable_pipeline_stage_step_id_fkey" FOREIGN KEY ("pipeline_stage_step_id") REFERENCES "public"."pipeline_stage_step" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step_condition;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step_condition"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage_step_condition'::regclass),
    "pipeline_stage_step_id"      integer,
    "condition_variable_id"       integer,      -- id of variable on which condition is written
    "condition_type"              varchar(255), -- SKIP, TRIGGER, SUCCESS or FAILURE
    "conditional_operator"        varchar(255),
    "conditional_value"           varchar(255),
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_step_condition_plugin_step_id_fkey" FOREIGN KEY ("pipeline_stage_step_id") REFERENCES "public"."pipeline_stage_step" ("id"),
    CONSTRAINT "pipeline_stage_step_condition_condition_variable_id_fkey" FOREIGN KEY ("condition_variable_id") REFERENCES "public"."pipeline_stage_step_variable" ("id"),
    PRIMARY KEY ("id")
);

----------- inserting values for PRESET plugins


INSERT INTO "public"."plugin_tag" ("id", "name", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'Load testing','f', 'now()', '1', 'now()', '1'),
('2', 'Code quality','f', 'now()', '1', 'now()', '1'),
('3', 'Security','f', 'now()', '1', 'now()', '1');

SELECT pg_catalog.setval('public.id_seq_plugin_tag', 3, true);

INSERT INTO "public"."plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'K6 Load testing','K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/k6-plugin-icon.png','f', 'now()', '1', 'now()', '1'),
('2', 'Sonarqube','Enhance Your Workflow with Continuous Code Quality & Code Security.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/sonarqube-plugin-icon.png','f', 'now()', '1', 'now()', '1');

SELECT pg_catalog.setval('public.id_seq_plugin_metadata', 2, true);

INSERT INTO "public"."plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', '1','1','now()', '1', 'now()', '1'),
('2', '2','2', 'now()', '1', 'now()', '1'),
('3', '3','2', 'now()', '1', 'now()', '1');

SELECT pg_catalog.setval('public.id_seq_plugin_tag_relation', 3, true);


INSERT INTO "public"."plugin_pipeline_script" ("id", "script", "type","deleted","created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'PathToScript=/devtroncd$RelativePathToScript

if [ $OutputType == "PROMETHEUS" ]
then
    wget https://go.dev/dl/go1.18.1.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.1.linux-amd64.tar.gz
    export GOPATH=/usr/local/go
    export GOCACHE=/usr/local/go/cache
    export PATH=$PATH:/usr/local/go/bin
    go install go.k6.io/xk6/cmd/xk6@latest
    xk6 build --with github.com/grafana/xk6-output-prometheus-remote
    K6_PROMETHEUS_USER=$PrometheusUsername \
    K6_PROMETHEUS_PASSWORD=$PrometheusApiKey \
    K6_PROMETHEUS_REMOTE_URL=$PrometheusRemoteWriteEndpoint \
    ./k6 run $PathToScript -o output-prometheus-remote
elif [ $OutputType == "LOG" ]
then
    docker pull grafana/k6
	docker run --rm -i grafana/k6 run - <$PathToScript
else
    echo "OUTPUT_TYPE: $OutputType not supported"
fi','SHELL','f','now()', '1', 'now()', '1'),
('2', 'PathToCodeDir=/devtroncd$CheckoutPath

cd $PathToCodeDir
echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
docker run \
    --rm \
    -e SONAR_HOST_URL=$SonarqubeEndpoint \
    -e SONAR_LOGIN=$SonarqubeApiKey \
    -v "/$PWD:/usr/src" \
    sonarsource/sonar-scanner-cli','SHELL','f', 'now()', '1', 'now()', '1');

SELECT pg_catalog.setval('public.id_seq_plugin_pipeline_script', 2, true);

INSERT INTO "public"."plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', '1','Step 1','Step 1 for K6 load testing','1','INLINE','1','f','now()', '1', 'now()', '1'),
('2', '2','Step 1','Step 1 for Sonarqube','1','INLINE','2','f','now()', '1', 'now()', '1');

SELECT pg_catalog.setval('public.id_seq_plugin_step', 2, true);


INSERT INTO "public"."plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', '1','RelativePathToScript','STRING','checkout path + script path along with script name','t','f','INPUT','NEW','/./script.js','1','f','now()', '1', 'now()', '1'),
('2', '1','PrometheusUsername','STRING','username of prometheus account','t','t','INPUT','NEW',null, '1' ,'f','now()', '1', 'now()', '1'),
('3', '1','PrometheusApiKey','STRING','api key of prometheus account','t','t','INPUT','NEW',null, '1','f','now()', '1', 'now()', '1'),
('4', '1','PrometheusRemoteWriteEndpoint','STRING','remote write endpoint of prometheus account','t','t','INPUT','NEW',null, '1','f','now()', '1', 'now()', '1'),
('5', '1','OutputType','STRING','output type - LOG or PROMETHEUS','t','f','INPUT','NEW','LOG', '1','f','now()', '1', 'now()', '1'),
('6', '2','SonarqubeProjectKey','STRING','project key of grafana sonarqube account','t','t','INPUT','NEW',null, '1', 'f','now()', '1', 'now()', '1'),
('7', '2','SonarqubeApiKey','STRING','api key of sonarqube account','t','t','INPUT','NEW',null, '1', 'f','now()', '1', 'now()', '1'),
('8', '2','SonarqubeEndpoint','STRING','api endpoint of sonarqube account','t','t','INPUT','NEW',null, '1','f','now()', '1', 'now()', '1'),
('9', '2','CheckoutPath','STRING','checkout path of git material','t','t','INPUT','NEW',null, '1','f','now()', '1', 'now()', '1');

SELECT pg_catalog.setval('public.id_seq_plugin_step_variable', 9, true);
