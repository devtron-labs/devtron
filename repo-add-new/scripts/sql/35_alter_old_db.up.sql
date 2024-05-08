----Dropping tables which are not being used and are not deleted earlier by migration

DROP TABLE IF EXISTS casbin_role CASCADE;

DROP TABLE IF EXISTS casbin CASCADE;

DROP TABLE IF EXISTS external_apps CASCADE;

DROP TABLE IF EXISTS pipeline_config CASCADE;