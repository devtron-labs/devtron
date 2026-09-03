INSERT INTO "public"."rbac_policy_data"
("entity","access_type","role","policy_data",
 "created_on","created_by","updated_on","updated_by","is_preset_role","deleted")
VALUES
    ('apps','argo-app','view','{
  "type": { "value": "p", "indexKeyMap": {} },
  "sub":  { "value": "argo-app:view_%_%",
            "indexKeyMap": { "14": "Env", "16": "App" } },
  "resActObjSet": [
    { "res": { "value": "argo-app", "indexKeyMap": {} },
      "act": { "value": "get", "indexKeyMap": {} },
      "obj": { "value": "%/%",
               "indexKeyMap": { "0": "EnvObj", "2": "AppObj" } } },
    { "res": { "value": "global-environment", "indexKeyMap": {} },
      "act": { "value": "get", "indexKeyMap": {} },
      "obj": { "value": "%", "indexKeyMap": { "0": "EnvObj" } } }
  ]
}','now()','1','now()','1',true,false),

    ('apps','argo-app','admin','{
  "type": { "value": "p", "indexKeyMap": {} },
  "sub":  { "value": "argo-app:admin_%_%",
            "indexKeyMap": { "15": "Env", "17": "App" } },
  "resActObjSet": [
    { "res": { "value": "argo-app", "indexKeyMap": {} },
      "act": { "value": "*", "indexKeyMap": {} },
      "obj": { "value": "%/%",
               "indexKeyMap": { "0": "EnvObj", "2": "AppObj" } } },
    { "res": { "value": "global-environment", "indexKeyMap": {} },
      "act": { "value": "get", "indexKeyMap": {} },
      "obj": { "value": "%", "indexKeyMap": { "0": "EnvObj" } } }
  ]
}','now()','1','now()','1',true,false),

    ('apps','flux-app','view','{
  "type": { "value": "p", "indexKeyMap": {} },
  "sub":  { "value": "flux-app:view_%_%",
            "indexKeyMap": { "14": "Env", "16": "App" } },
  "resActObjSet": [
    { "res": { "value": "flux-app", "indexKeyMap": {} },
      "act": { "value": "get", "indexKeyMap": {} },
      "obj": { "value": "%/%",
               "indexKeyMap": { "0": "EnvObj", "2": "AppObj" } } },
    { "res": { "value": "global-environment", "indexKeyMap": {} },
      "act": { "value": "get", "indexKeyMap": {} },
      "obj": { "value": "%", "indexKeyMap": { "0": "EnvObj" } } }
  ]
}','now()','1','now()','1',true,false),

    ('apps','flux-app','admin','{
  "type": { "value": "p", "indexKeyMap": {} },
  "sub":  { "value": "flux-app:admin_%_%",
            "indexKeyMap": { "15": "Env", "17": "App" } },
  "resActObjSet": [
    { "res": { "value": "flux-app", "indexKeyMap": {} },
      "act": { "value": "*", "indexKeyMap": {} },
      "obj": { "value": "%/%",
               "indexKeyMap": { "0": "EnvObj", "2": "AppObj" } } },
    { "res": { "value": "global-environment", "indexKeyMap": {} },
      "act": { "value": "get", "indexKeyMap": {} },
      "obj": { "value": "%", "indexKeyMap": { "0": "EnvObj" } } }
  ]
}','now()','1','now()','1',true,false);

INSERT INTO "public"."rbac_role_data"
("entity","access_type","role","role_display_name","role_description","role_data",
 "created_on","created_by","updated_on","updated_by","is_preset_role","deleted")
VALUES
    ('apps','argo-app','view','View only',
     'Can view selected Argo CD application(s) and resource manifests of selected application(s)','{
  "role":       { "value": "argo-app:view_%_%",
                  "indexKeyMap": { "14": "Env", "16": "App" } },
  "entityName": { "value": "%", "indexKeyMap": { "0": "App" } },
  "environment":{ "value": "%", "indexKeyMap": { "0": "Env" } },
  "action":     { "value": "view", "indexKeyMap": {} },
  "entity":     { "value": "%", "indexKeyMap": { "0": "Entity" } },
  "accessType": { "value": "argo-app", "indexKeyMap": {} }
}','now()','1','now()','1',true,false),

    ('apps','argo-app','admin','Admin',
     'Complete access on selected Argo CD application(s)','{
  "role":       { "value": "argo-app:admin_%_%",
                  "indexKeyMap": { "15": "Env", "17": "App" } },
  "entityName": { "value": "%", "indexKeyMap": { "0": "App" } },
  "environment":{ "value": "%", "indexKeyMap": { "0": "Env" } },
  "action":     { "value": "admin", "indexKeyMap": {} },
  "entity":     { "value": "%", "indexKeyMap": { "0": "Entity" } },
  "accessType": { "value": "argo-app", "indexKeyMap": {} }
}','now()','1','now()','1',true,false),

    ('apps','flux-app','view','View only',
     'Can view selected Flux CD application(s) and resource manifests of selected application(s)','{
  "role":       { "value": "flux-app:view_%_%",
                  "indexKeyMap": { "14": "Env", "16": "App" } },
  "entityName": { "value": "%", "indexKeyMap": { "0": "App" } },
  "environment":{ "value": "%", "indexKeyMap": { "0": "Env" } },
  "action":     { "value": "view", "indexKeyMap": {} },
  "entity":     { "value": "%", "indexKeyMap": { "0": "Entity" } },
  "accessType": { "value": "flux-app", "indexKeyMap": {} }
}','now()','1','now()','1',true,false),

    ('apps','flux-app','admin','Admin',
     'Complete access on selected Flux CD application(s)','{
  "role":       { "value": "flux-app:admin_%_%",
                  "indexKeyMap": { "15": "Env", "17": "App" } },
  "entityName": { "value": "%", "indexKeyMap": { "0": "App" } },
  "environment":{ "value": "%", "indexKeyMap": { "0": "Env" } },
  "action":     { "value": "admin", "indexKeyMap": {} },
  "entity":     { "value": "%", "indexKeyMap": { "0": "Entity" } },
  "accessType": { "value": "flux-app", "indexKeyMap": {} }
}','now()','1','now()','1',true,false);