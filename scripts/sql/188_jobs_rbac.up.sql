INSERT INTO "public"."rbac_policy_data" ("entity", "access_type", "role", "policy_data", "created_on", "created_by",
                                         "updated_on", "updated_by", "is_preset_role", "deleted")
VALUES ('jobs', '', 'admin', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "jobs:admin_%_%_%_%",
           "indexKeyMap": {
             "11": "Team",
             "13": "Env",
             "15": "App",
             "17": "Workflow"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "jobs",
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
               "value": "jobEnv",
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
           },
           {
             "res": {
               "value": "workflow",
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
                 "2": "AppObj",
                 "4": "WorkflowObj"
               }
             }
            }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('jobs', '', 'trigger', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "jobs:trigger_%_%_%_%",
           "indexKeyMap": {
             "13": "Team",
             "15": "Env",
             "17": "App",
             "19": "Workflow"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "jobs",
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
               "value": "jobs",
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
               "value": "jobEnv",
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
               "value": "jobEnv",
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
           },
           {
             "res": {
               "value": "workflow",
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
                 "2": "AppObj",
                 "4": "WorkflowObj"
               }
             }
           },
           {
             "res": {
               "value": "workflow",
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
                 "2": "AppObj",
                 "4": "WorkflowObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false),
       ('jobs', '', 'view', '{
         "type": {
           "value": "p",
           "indexKeyMap": {}
         },
         "sub": {
           "value": "jobs:view_%_%_%_%",
           "indexKeyMap": {
             "10": "Team",
             "12": "Env",
             "14": "App",
             "16": "Workflow"
           }
         },
         "resActObjSet": [
           {
             "res": {
               "value": "jobs",
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
               "value": "jobEnv",
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
           },
           {
             "res": {
               "value": "workflow",
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
                 "2": "AppObj",
                 "4": "WorkflowObj"
               }
             }
           }
         ]
       }', 'now()', '1', 'now()', '1', true, false);









INSERT INTO "public"."rbac_role_data" ("entity", "access_type", "role", "role_display_name", "role_description",
                                       "role_data", "created_on", "created_by", "updated_on", "updated_by",
                                       "is_preset_role", "deleted")
VALUES ('jobs', '', 'admin', 'Admin', 'Can view, run and edit jobs in selected scope', '{
         "role": {
           "value": "jobs:admin_%_%_%_%",
           "indexKeyMap": {
             "11": "Team",
             "13": "Env",
             "15": "App",
             "17": "Workflow"
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
         "workflow": {
           "value": "%",
           "indexKeyMap": {
             "0": "Workflow"
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
           "value": "",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('jobs', '', 'trigger', 'Run job', 'Can run jobs in selected scope build', '{
         "role": {
           "value": "jobs:trigger_%_%_%_%",
           "indexKeyMap": {
             "13": "Team",
             "15": "Env",
             "17": "App",
             "19": "Workflow"
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
         "workflow": {
           "value": "%",
           "indexKeyMap": {
             "0": "Workflow"
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
           "value": "",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false),
       ('jobs', '', 'view', 'View only', 'Can view selected jobs', '{
         "role": {
           "value": "jobs:view_%_%_%_%",
           "indexKeyMap": {
             "10": "Team",
             "12": "Env",
             "14": "App",
             "16": "Workflow"
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
         "workflow": {
           "value": "%",
           "indexKeyMap": {
             "0": "Workflow"
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
           "value": "",
           "indexKeyMap": {}
         }
       }', 'now()', '1', 'now()', '1', true, false);





INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('jobEnv', '{ "value": "jobEnv", "indexKeyMap": {}}', ARRAY['get','update','create','delete','trigger'],'{"value": "%/%/%","indexKeyMap": {"0": "TeamObj","2": "EnvObj","4": "AppObj"}}', ARRAY['jobs'],'f','now()', '1', 'now()', '1');

INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('workflow', '{ "value": "workflow", "indexKeyMap": {}}', ARRAY['get','update','create','delete','trigger'],'{"value": "%/%/%","indexKeyMap": {"0": "TeamObj","2": "AppObj","4": "WorkflowObj"}}', ARRAY['jobs'],'f','now()', '1', 'now()', '1');




UPDATE rbac_policy_resource_detail set eligible_entity_access_types = ARRAY['apps/devtron-app','apps/helm-app','jobs'] where resource='project' OR resource ='global-environment' OR resource='terminal';

UPDATE rbac_role_resource_detail set eligible_entity_access_types = ARRAY['apps/devtron-app','apps/helm-app','jobs'] where resource ='project' OR resource ='environment';

INSERT INTO rbac_role_resource_detail ("resource", "role_resource_key", "role_resource_update_key",
                                       "eligible_entity_access_types", "deleted", "created_on", "created_by",
                                       "updated_on", "updated_by")
VALUES ('workflow', 'Workflow', 'Workflow', ARRAY ['jobs'], false, now(), 1, now(), 1);

Alter table roles add column IF NOT EXISTS workflow text;


