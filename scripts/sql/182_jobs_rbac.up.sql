INSERT INTO "public"."rbac_policy_data" ("entity", "access_type", "role", "policy_data", "created_on", "created_by",
                                         "updated_on", "updated_by", "is_preset_role", "deleted")
VALUES ('apps', 'jobs', 'manager', '{
  "type": {
    "value": "p",
    "indexKeyMap": {}
  },
  "sub": {
    "value": "apps/jobs:manager_%_%_%",
    "indexKeyMap": {
      "18": "Team",
      "20": "Env",
      "22": "App"
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
        "value": "appEnv",
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
       ('apps', 'jobs', 'admin', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "apps/jobs:admin_%_%_%",
           "indexKeyMap": {
             "16": "Team",
             "18": "Env",
             "20": "App"
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
               "value": "appEnv",
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
       ('apps', 'jobs', 'trigger', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "apps/jobs:trigger_%_%_%",
           "indexKeyMap": {
             "18": "Team",
             "20": "Env",
             "22": "App"
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
               "value": "appEnv",
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
               "value": "appEnv",
               "indexKeyMap": {}
             },
             "act": {
               "value": "trigger",
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
       ('apps', 'jobs', 'view', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "apps/jobs:view_%_%_%",
           "indexKeyMap": {
             "15": "Team",
             "17": "Env",
             "19": "App"
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
               "value": "appEnv",
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
       }', 'now()', '1', 'now()', '1', true, false);


INSERT INTO "public"."rbac_role_data" ("entity", "access_type", "role", "role_display_name", "role_description",
                                       "role_data", "created_on", "created_by", "updated_on", "updated_by",
                                       "is_preset_role", "deleted")
VALUES ('apps', 'jobs', 'manager', 'Manager',
        'Can view, run and edit jobs in selected scope. Can also manage user access within the scope.', '{
    "role": {
      "value": "apps/jobs:manager_%_%_%",
      "indexKeyMap": {
        "18": "Team",
        "20": "Env",
        "22": "App"
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
      "value": "jobs",
      "indexKeyMap": {}
    }
  }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'jobs', 'admin', 'Admin', 'Can view, run and edit jobs in selected scope', '{
         "role": {
           "value": "apps/jobs:admin_%_%_%",
           "indexKeyMap": {
             "16": "Team",
             "18": "Env",
             "20": "App"
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
           "value": "jobs",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'jobs', 'trigger', 'Run job', 'Can run jobs in selected scope build', '{
         "role": {
           "value": "apps/jobs:trigger_%_%_%",
           "indexKeyMap": {
             "18": "Team",
             "20": "Env",
             "22": "App"
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
           "value": "jobs",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('apps', 'jobs', 'view', 'View only', 'Can view selected jobs', '{
         "role": {
           "value": "apps/jobs:view_%_%_%",
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
           "value": "jobs",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false);


INSERT INTO "public"."rbac_policy_resource_detail" ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('appEnv', '{ "value": "appEnv", "indexKeyMap": {}}', ARRAY['get','update','create','delete','trigger'],'{"value": "%/%/%","indexKeyMap": {"0": "TeamObj","2": "EnvObj","4": "AppObj"}}', ARRAY['apps/jobs'],'f','now()', '1', 'now()', '1');

UPDATE "public"."rbac_policy_resource_detail" set
eligible_entity_access_types = ARRAY['apps/devtron-app','apps/jobs'] where resource='applications' OR resource ='user';

UPDATE "public"."rbac_policy_resource_detail" set
eligible_entity_access_types = ARRAY['apps/devtron-app','apps/helm-app','apps/jobs'] where resource='project' OR resource ='global-environment' OR resource='terminal';

UPDATE "public"."rbac_role_resource_detail" set
eligible_entity_access_types = ARRAY['apps/devtron-app','apps/helm-app','apps/jobs'] where resource='applications' OR resource ='project' OR resource ='environment';

