UPDATE rbac_role_data
SET deleted= true
where entity = 'cluster'
  and role = 'admin';


-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_rbac_policy_resource_detail;

-- Table Definition
CREATE TABLE IF NOT EXISTS rbac_policy_resource_detail
(
    "id"                           int4         NOT NULL DEFAULT nextval('id_seq_rbac_policy_resource_detail'::regclass),
    "resource"                     varchar(250) NOT NULL,
    "policy_resource_value"        jsonb,
    "allowed_actions"              varchar(100)[],
    "resource_object"              jsonb,
    "eligible_entity_access_types" varchar(250)[],
    "deleted"                      bool         NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE (resource)
);

INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")

VALUES ('applications', '{
  "value": "applications",
  "indexKeyMap": {}
}', ARRAY ['get','update','create','delete','trigger'],
        '{
          "value": "%/%",
          "indexKeyMap": {
            "0": "TeamObj",
            "2": "AppObj"
          }
        }', ARRAY ['apps/devtron-app'], false, now(), 1, now(), 1),

       ('project', '{
         "value": "team",
         "indexKeyMap": {}
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%",
          "indexKeyMap": {
            "0": "TeamObj"
          }
        }', ARRAY ['apps/devtron-app','apps/helm-app'], false, now(), 1, now(), 1),

       ('environment', '{
         "value": "environment",
         "indexKeyMap": {}
       }', ARRAY ['get','update','create','delete','trigger'],
        '{
          "value": "%/%",
          "indexKeyMap": {
            "0": "EnvObj",
            "2": "AppObj"
          }
        }', ARRAY ['apps/devtron-app'], false, now(), 1, now(), 1),

       ('user', '{
         "value": "user",
         "indexKeyMap": {}
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%",
          "indexKeyMap": {
            "0": "TeamObj"
          }
        }', ARRAY ['apps/devtron-app'], false, now(), 1, now(), 1),

       ('notification', '{
         "value": "notification",
         "indexKeyMap": {}
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%",
          "indexKeyMap": {
            "0": "TeamObj"
          }
        }', ARRAY ['apps/devtron-app'], false, now(), 1, now(), 1),

       ('global-environment', '{
         "value": "global-environment",
         "indexKeyMap": {}
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%",
          "indexKeyMap": {
            "0": "EnvObj"
          }
        }', ARRAY ['apps/devtron-app','apps/helm-app'], false, now(), 1, now(), 1),

       ('terminal', '{
         "value": "terminal",
         "indexKeyMap": {}
       }', ARRAY ['exec'],
        '{
          "value": "%/%/%",
          "indexKeyMap": {
            "0": "TeamObj",
            "2": "EnvObj",
            "4": "AppObj"
          }
        }', ARRAY ['apps/devtron-app','apps/helm-app'], false, now(), 1, now(), 1),

       ('chart-group', '{
         "value": "%",
         "indexKeyMap": {
           "0": "Entity"
         }
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%",
          "indexKeyMap": {
            "0": "EntityName"
          }
        }', ARRAY ['chart-group'], true, now(), 1, now(), 1),

       ('helm-app', '{
         "value": "helm-app",
         "indexKeyMap": {}
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%/%/%",
          "indexKeyMap": {
            "0": "TeamObj",
            "2": "EnvObj",
            "4": "AppObj"
          }
        }', ARRAY ['apps/helm-app'], false, now(), 1, now(), 1),

       ('cluster/namespace', '{
         "value": "%/%",
         "indexKeyMap": {
           "0": "ClusterObj",
           "2": "NamespaceObj"
         }
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%/%/%",
          "indexKeyMap": {
            "0": "GroupObj",
            "2": "KindObj",
            "4": "ResourceObj"
          }
        }', ARRAY ['cluster'], false, now(), 1, now(), 1),

       ('cluster/namespace/user', '{
         "value": "%/%/user",
         "indexKeyMap": {
           "0": "ClusterObj",
           "2": "NamespaceObj"
         }
       }', ARRAY ['get','update','create','delete'],
        '{
          "value": "%/%/%",
          "indexKeyMap": {
            "0": "GroupObj",
            "2": "KindObj",
            "4": "ResourceObj"
          }
        }', ARRAY ['cluster'], true, now(), 1, now(), 1),

       ('ci-pipeline/source-value', '{
         "value": "ci-pipeline/source-value",
         "indexKeyMap": {}
       }', ARRAY ['get','update'],
        '{
          "value": "%/%/%",
          "indexKeyMap": {
            "0": "TeamObj",
            "2": "EnvObj",
            "4": "AppObj"
          }
        }', ARRAY ['apps/devtron-app'], false, now(), 1, now(), 1);


-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_rbac_role_resource_detail;

-- Table Definition
CREATE TABLE IF NOT EXISTS rbac_role_resource_detail
(
    "id"                           int4         NOT NULL DEFAULT nextval('id_seq_rbac_role_resource_detail'::regclass),
    "resource"                     varchar(250) NOT NULL,
    "role_resource_key"            varchar(100),
    "role_resource_update_key"     varchar(100),
    "eligible_entity_access_types" varchar(250)[],
    "deleted"                      bool         NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE (resource)
);

INSERT INTO rbac_role_resource_detail ("resource", "role_resource_key", "role_resource_update_key",
                                       "eligible_entity_access_types", "deleted", "created_on", "created_by",
                                       "updated_on", "updated_by")

VALUES ('applications', 'EntityName', 'App', ARRAY ['apps/devtron-app','apps/helm-app'], false, now(), 1, now(), 1),
       ('project', 'Team', 'Team', ARRAY ['apps/devtron-app','apps/helm-app'], false, now(), 1, now(), 1),
       ('environment', 'Environment', 'Env', ARRAY ['apps/devtron-app','apps/helm-app'], false, now(), 1, now(), 1),
       ('cluster', 'Cluster', 'Cluster', ARRAY ['cluster'], false, now(), 1, now(), 1),
       ('namespace', 'Namespace', 'Namespace', ARRAY ['cluster'], false, now(), 1, now(), 1),
       ('group', 'Group', 'Group', ARRAY ['cluster'], false, now(), 1, now(), 1),
       ('kind', 'Kind', 'Kind', ARRAY ['cluster'], false, now(), 1, now(), 1),
       ('resource', 'Resource', 'Resource', ARRAY ['cluster'], false, now(), 1, now(), 1),
       ('chart-group-name', 'EntityName', 'EntityName', ARRAY ['chart-group'], true, now(), 1, now(), 1);
