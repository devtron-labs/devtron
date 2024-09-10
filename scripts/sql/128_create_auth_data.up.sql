-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_rbac_policy_data;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."rbac_policy_data"
(
    "id"             int          NOT NULL DEFAULT nextval('id_seq_rbac_policy_data'::regclass),
    "entity"         varchar(250) NOT NULL,
    "access_type"    varchar(250) NOT NULL,
    "role"           varchar(250) NOT NULL,
    "policy_data"    jsonb        NOT NULL,
    "created_on"     timestamptz,
    "created_by"     integer,
    "updated_on"     timestamptz,
    "updated_by"     integer,
    "is_preset_role" boolean      NOT NULL DEFAULT FALSE,
    "deleted"        boolean      NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("entity", "access_type", "role")
);


INSERT INTO "public"."rbac_policy_data" ("entity", "access_type", "role", "policy_data", "created_on", "created_by",
                                         "updated_on", "updated_by", "is_preset_role", "deleted")
VALUES ('apps', 'devtron-app', 'manager', '{
  "type": {
    "value": "p",
    "indexKeyMap": {}
  },
  "sub": {
    "value": "role:manager_%_%_%",
    "indexKeyMap": {
      "13": "Team",
      "15": "Env",
      "17": "App"
    }
  },
  "resActObjSet": [
    {
      "res": {
        "value": "applications",
        "indexKeyMap": {}
      },
      "act": {
        "value": "*",
        "indexKeyMap": {}
      },
      "obj": {
        "value": "%/%",
        "indexKeyMap": {
          "0": "TeamObj",
          "2": "AppObj"
        }
      }
    },
    {
      "res": {
        "value": "environment",
        "indexKeyMap": {}
      },
      "act": {
        "value": "*",
        "indexKeyMap": {}
      },
      "obj": {
        "value": "%/%",
        "indexKeyMap": {
          "0": "EnvObj",
          "2": "AppObj"
        }
      }
    },
    {
      "res": {
        "value": "team",
        "indexKeyMap": {}
      },
      "act": {
        "value": "*",
        "indexKeyMap": {}
      },
      "obj": {
        "value": "%",
        "indexKeyMap": {
          "0": "TeamObj"
        }
      }
    },
    {
      "res": {
        "value": "user",
        "indexKeyMap": {}
      },
      "act": {
        "value": "*",
        "indexKeyMap": {}
      },
      "obj": {
        "value": "%",
        "indexKeyMap": {
          "0": "TeamObj"
        }
      }
    },
    {
      "res": {
        "value": "notification",
        "indexKeyMap": {}
      },
      "act": {
        "value": "*",
        "indexKeyMap": {}
      },
      "obj": {
        "value": "%",
        "indexKeyMap": {
          "0": "TeamObj"
        }
      }
    },
    {
      "res": {
        "value": "global-environment",
        "indexKeyMap": {}
      },
      "act": {
        "value": "*",
        "indexKeyMap": {}
      },
      "obj": {
        "value": "%",
        "indexKeyMap": {
          "0": "EnvObj"
        }
      }
    }
  ]
}', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'devtron-app', 'admin', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:admin_%_%_%",
           "indexKeyMap": {
             "11": "Team",
             "13": "Env",
             "15": "App"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "applications",
               "indexKeyMap": {}
             },
             "act": {
               "value": "*",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "*",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "EnvObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "team",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "TeamObj"
               }
             }
           },
           {
             "res": {
               "value": "global-environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EnvObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'devtron-app', 'trigger', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:trigger_%_%_%",
           "indexKeyMap": {
             "13": "Team",
             "15": "Env",
             "17": "App"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "applications",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "applications",
               "indexKeyMap": {}
             },
             "act": {
               "value": "trigger",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "EnvObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "trigger",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "EnvObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "team",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "TeamObj"
               }
             }
           },
           {
             "res": {
               "value": "global-environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EnvObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'devtron-app', 'view', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:view_%_%_%",
           "indexKeyMap": {
             "10": "Team",
             "12": "Env",
             "14": "App"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "applications",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "EnvObj",
                 "2": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "team",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "TeamObj"
               }
             }
           },
           {
             "res": {
               "value": "global-environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EnvObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('chart-group', '', 'admin', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:%_admin",
           "indexKeyMap": {
             "5": "Entity"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "%",
               "indexKeyMap": {
                 "0": "Entity"
               }
             },
             "act": {
               "value": "*",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "*",
               "indexKeyMap": {}
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('chart-group', '', 'view', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:%_view",
           "indexKeyMap": {
             "5": "Entity"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "%",
               "indexKeyMap": {
                 "0": "Entity"
               }
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "*",
               "indexKeyMap": {}
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('chart-group', '', 'update', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:%_%_specific",
           "indexKeyMap": {
             "5": "Entity",
             "7": "EntityName"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "%",
               "indexKeyMap": {
                 "0": "Entity"
               }
             },
             "act": {
               "value": "update",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EntityName"
               }
             }
           },
           {
             "res": {
               "value": "%",
               "indexKeyMap": {
                 "0": "Entity"
               }
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EntityName"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'helm-app', 'admin', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "helm-app:admin_%_%_%",
           "indexKeyMap": {
             "15": "Team",
             "17": "Env",
             "19": "App"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "helm-app",
               "indexKeyMap": {}
             },
             "act": {
               "value": "*",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "EnvObj",
                 "4": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "team",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "TeamObj"
               }
             }
           },
           {
             "res": {
               "value": "global-environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EnvObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'helm-app', 'edit', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "helm-app:edit_%_%_%",
           "indexKeyMap": {
             "14": "Team",
             "16": "Env",
             "18": "App"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "helm-app",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "EnvObj",
                 "4": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "helm-app",
               "indexKeyMap": {}
             },
             "act": {
               "value": "update",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "EnvObj",
                 "4": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "global-environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EnvObj"
               }
             }
           },
           {
             "res": {
               "value": "team",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "TeamObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'helm-app', 'view', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "helm-app:view_%_%_%",
           "indexKeyMap": {
             "14": "Team",
             "16": "Env",
             "18": "App"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "helm-app",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "TeamObj",
                 "2": "EnvObj",
                 "4": "AppObj"
               }
             }
           },
           {
             "res": {
               "value": "global-environment",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "EnvObj"
               }
             }
           },
           {
             "res": {
               "value": "team",
               "indexKeyMap": {}
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%",
               "indexKeyMap": {
                 "0": "TeamObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('cluster', '', 'admin', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:clusterAdmin_%_%_%_%_%",
           "indexKeyMap": {
             "18": "Cluster",
             "20": "Namespace",
             "22": "Group",
             "24": "Kind",
             "26": "Resource"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "ClusterObj",
                 "2": "NamespaceObj"
               }
             },
             "act": {
               "value": "*",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "GroupObj",
                 "2": "KindObj",
                 "4": "ResourceObj"
               }
             }
           },
           {
             "res": {
               "value": "%/%/user",
               "indexKeyMap": {
                 "0": "ClusterObj",
                 "2": "NamespaceObj"
               }
             },
             "act": {
               "value": "*",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "GroupObj",
                 "2": "KindObj",
                 "4": "ResourceObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('cluster', '', 'edit', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:clusterEdit_%_%_%_%_%",
           "indexKeyMap": {
             "17": "Cluster",
             "19": "Namespace",
             "21": "Group",
             "23": "Kind",
             "25": "Resource"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "ClusterObj",
                 "2": "NamespaceObj"
               }
             },
             "act": {
               "value": "*",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "GroupObj",
                 "2": "KindObj",
                 "4": "ResourceObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('cluster', '', 'view', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "role:clusterView_%_%_%_%_%",
           "indexKeyMap": {
             "17": "Cluster",
             "19": "Namespace",
             "21": "Group",
             "23": "Kind",
             "25": "Resource"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "%/%",
               "indexKeyMap": {
                 "0": "ClusterObj",
                 "2": "NamespaceObj"
               }
             },
             "act": {
               "value": "get",
               "indexKeyMap": {}
             },
             "obj": {
               "value": "%/%/%",
               "indexKeyMap": {
                 "0": "GroupObj",
                 "2": "KindObj",
                 "4": "ResourceObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false);


-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_rbac_role_data;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."rbac_role_data"
(
    "id"                int          NOT NULL DEFAULT nextval('id_seq_rbac_role_data'::regclass),
    "entity"            varchar(250) NOT NULL,
    "access_type"       varchar(250) NOT NULL,
    "role"              varchar(250) NOT NULL,
    "role_display_name" varchar(250) NOT NULL,
    "role_data"         jsonb        NOT NULL,
    "role_description"  text,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    "is_preset_role"    boolean      NOT NULL DEFAULT FALSE,
    "deleted"           boolean      NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("entity", "access_type", "role")
);


INSERT INTO "public"."rbac_role_data" ("entity", "access_type", "role", "role_display_name", "role_description",
                                       "role_data", "created_on", "created_by", "updated_on", "updated_by",
                                       "is_preset_role", "deleted")
VALUES ('apps', 'devtron-app', 'manager', 'Manager',
        'Can view, trigger and edit selected applications. Can also manage user access.', '{
    "role": {
      "value": "role:manager_%_%_%",
      "indexKeyMap": {
        "13": "Team",
        "15": "Env",
        "17": "App"
      }
    },
    "team": {
      "value": "%",
      "indexKeyMap": {
        "0": "Team"
      }
    },
    "entityName": {
      "value": "%",
      "indexKeyMap": {
        "0": "App"
      }
    },
    "environment": {
      "value": "%",
      "indexKeyMap": {
        "0": "Env"
      }
    },
    "action": {
      "value": "manager",
      "indexKeyMap": {}
    },
    "entity": {
      "value": "%",
      "indexKeyMap": {
        "0": "Entity"
      }
    },
    "accessType": {
      "value": "devtron-app",
      "indexKeyMap": {}
    }
  }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'devtron-app', 'admin', 'Admin', 'Can view, trigger and edit selected applications', '{
         "role": {
           "value": "role:admin_%_%_%",
           "indexKeyMap": {
             "11": "Team",
             "13": "Env",
             "15": "App"
           }
         },
         "team": {
           "value": "%",
           "indexKeyMap": {
             "0": "Team"
           }
         },
         "entityName": {
           "value": "%",
           "indexKeyMap": {
             "0": "App"
           }
         },
         "environment": {
           "value": "%",
           "indexKeyMap": {
             "0": "Env"
           }
         },
         "action": {
           "value": "admin",
           "indexKeyMap": {}
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "accessType": {
           "value": "devtron-app",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'devtron-app', 'trigger', 'Build and deploy', 'Can build and deploy apps on selected environments', '{
         "role": {
           "value": "role:trigger_%_%_%",
           "indexKeyMap": {
             "13": "Team",
             "15": "Env",
             "17": "App"
           }
         },
         "team": {
           "value": "%",
           "indexKeyMap": {
             "0": "Team"
           }
         },
         "entityName": {
           "value": "%",
           "indexKeyMap": {
             "0": "App"
           }
         },
         "environment": {
           "value": "%",
           "indexKeyMap": {
             "0": "Env"
           }
         },
         "action": {
           "value": "trigger",
           "indexKeyMap": {}
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "accessType": {
           "value": "devtron-app",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'devtron-app', 'view', 'View only', 'Can view selected applications', '{
         "role": {
           "value": "role:view_%_%_%",
           "indexKeyMap": {
             "10": "Team",
             "12": "Env",
             "14": "App"
           }
         },
         "team": {
           "value": "%",
           "indexKeyMap": {
             "0": "Team"
           }
         },
         "entityName": {
           "value": "%",
           "indexKeyMap": {
             "0": "App"
           }
         },
         "environment": {
           "value": "%",
           "indexKeyMap": {
             "0": "Env"
           }
         },
         "action": {
           "value": "view",
           "indexKeyMap": {}
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "accessType": {
           "value": "devtron-app",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('chart-group', '', 'admin', 'Admin', '', '{
         "role": {
           "value": "role:%_admin",
           "indexKeyMap": {
             "5": "Entity"
           }
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "action": {
           "value": "admin",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('chart-group', '', 'view', 'View', '', '{
         "role": {
           "value": "role:%_view",
           "indexKeyMap": {
             "5": "Entity"
           }
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "action": {
           "value": "view",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('chart-group', '', 'update', 'Update', '', '{
         "role": {
           "value": "role:%_%_specific",
           "indexKeyMap": {
             "5": "Entity",
             "7": "EntityName"
           }
         },
         "entityName": {
           "value": "%",
           "indexKeyMap": {
             "0": "EntityName"
           }
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "action": {
           "value": "update",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'helm-app', 'admin', 'Admin', 'Complete access on selected applications', '{
         "role": {
           "value": "helm-app:admin_%_%_%",
           "indexKeyMap": {
             "15": "Team",
             "17": "Env",
             "19": "App"
           }
         },
         "team": {
           "value": "%",
           "indexKeyMap": {
             "0": "Team"
           }
         },
         "entityName": {
           "value": "%",
           "indexKeyMap": {
             "0": "App"
           }
         },
         "environment": {
           "value": "%",
           "indexKeyMap": {
             "0": "Env"
           }
         },
         "action": {
           "value": "admin",
           "indexKeyMap": {}
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "accessType": {
           "value": "helm-app",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'helm-app', 'edit', 'View & Edit', 'Can also edit resource manifests of selected application(s)', '{
         "role": {
           "value": "helm-app:edit_%_%_%",
           "indexKeyMap": {
             "14": "Team",
             "16": "Env",
             "18": "App"
           }
         },
         "team": {
           "value": "%",
           "indexKeyMap": {
             "0": "Team"
           }
         },
         "entityName": {
           "value": "%",
           "indexKeyMap": {
             "0": "App"
           }
         },
         "environment": {
           "value": "%",
           "indexKeyMap": {
             "0": "Env"
           }
         },
         "action": {
           "value": "edit",
           "indexKeyMap": {}
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "accessType": {
           "value": "helm-app",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'helm-app', 'view', 'View only',
        'Can view selected application(s) and resource manifests of selected application(s)', '{
         "role": {
           "value": "helm-app:view_%_%_%",
           "indexKeyMap": {
             "14": "Team",
             "16": "Env",
             "18": "App"
           }
         },
         "team": {
           "value": "%",
           "indexKeyMap": {
             "0": "Team"
           }
         },
         "entityName": {
           "value": "%",
           "indexKeyMap": {
             "0": "App"
           }
         },
         "environment": {
           "value": "%",
           "indexKeyMap": {
             "0": "Env"
           }
         },
         "action": {
           "value": "view",
           "indexKeyMap": {}
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "accessType": {
           "value": "helm-app",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('cluster', '', 'admin', 'Manager',
        'Create, view, edit & delete allowed K8s resources. Can also manage user access.', '{
         "role": {
           "value": "role:clusterAdmin_%_%_%_%_%",
           "indexKeyMap": {
             "18": "Cluster",
             "20": "Namespace",
             "22": "Group",
             "24": "Kind",
             "26": "Resource"
           }
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "cluster": {
           "value": "%",
           "indexKeyMap": {
             "0": "Cluster"
           }
         },
         "namespace": {
           "value": "%",
           "indexKeyMap": {
             "0": "Namespace"
           }
         },
         "group": {
           "value": "%",
           "indexKeyMap": {
             "0": "Group"
           }
         },
         "kind": {
           "value": "%",
           "indexKeyMap": {
             "0": "Kind"
           }
         },
         "resource": {
           "value": "%",
           "indexKeyMap": {
             "0": "Resource"
           }
         },
         "action": {
           "value": "admin",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('cluster', '', 'edit', 'Admin', 'Create, view, edit & delete allowed K8s resources.', '{
         "role": {
           "value": "role:clusterEdit_%_%_%_%_%",
           "indexKeyMap": {
             "17": "Cluster",
             "19": "Namespace",
             "21": "Group",
             "23": "Kind",
             "25": "Resource"
           }
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "cluster": {
           "value": "%",
           "indexKeyMap": {
             "0": "Cluster"
           }
         },
         "namespace": {
           "value": "%",
           "indexKeyMap": {
             "0": "Namespace"
           }
         },
         "group": {
           "value": "%",
           "indexKeyMap": {
             "0": "Group"
           }
         },
         "kind": {
           "value": "%",
           "indexKeyMap": {
             "0": "Kind"
           }
         },
         "resource": {
           "value": "%",
           "indexKeyMap": {
             "0": "Resource"
           }
         },
         "action": {
           "value": "edit",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('cluster', '', 'view', 'View', 'View allowed K8s resources.', '{
         "role": {
           "value": "role:clusterView_%_%_%_%_%",
           "indexKeyMap": {
             "17": "Cluster",
             "19": "Namespace",
             "21": "Group",
             "23": "Kind",
             "25": "Resource"
           }
         },
         "entity": {
           "value": "%",
           "indexKeyMap": {
             "0": "Entity"
           }
         },
         "cluster": {
           "value": "%",
           "indexKeyMap": {
             "0": "Cluster"
           }
         },
         "namespace": {
           "value": "%",
           "indexKeyMap": {
             "0": "Namespace"
           }
         },
         "group": {
           "value": "%",
           "indexKeyMap": {
             "0": "Group"
           }
         },
         "kind": {
           "value": "%",
           "indexKeyMap": {
             "0": "Kind"
           }
         },
         "resource": {
           "value": "%",
           "indexKeyMap": {
             "0": "Resource"
           }
         },
         "action": {
           "value": "view",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false);