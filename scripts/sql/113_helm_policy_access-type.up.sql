


ALTER TABLE "public"."default_auth_policy"
    ADD COLUMN access_type varchar(50);

UPDATE "public"."default_auth_policy"
SET access_type = 'devtron-app'
WHERE role_type = 'manager'OR  role_type = 'trigger' OR  role_type = 'view'OR  role_type ='admin';


UPDATE "public"."default_auth_policy"
SET access_type = 'cluster'
WHERE role_type = 'clusterEdit'OR role_type = 'clusterView'OR  role_type = 'clusterAdmin';


UPDATE "public"."default_auth_policy"
SET access_type = 'globalEntity'
WHERE role_type = 'entityView' OR role_type = 'entitySpecific'OR role_type = 'entityAll';

SELECT setval('id_seq_default_auth_policy', (SELECT MAX(id) FROM default_auth_policy));

INSERT INTO "public"."default_auth_policy" ( "role_type", "policy", "created_on", "created_by", "updated_on", "updated_by","access_type") VALUES
                                                                                                                                     ( 'admin', '{
    "data": [
         {
            "type": "p",
            "sub": "helm-app:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "helm-app",
            "act": "*",
            "obj": "{{.TeamObj}}/{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "helm-app:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "helm-app:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1','helm-app'),
                                                                                                                                     ('edit', '{
    "data": [
        {
            "type": "p",
            "sub": "helm-app:edit_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "helm-app",
            "act": "get",
            "obj": "{{.TeamObj}}/{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "helm-app:edit_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "helm-app",
            "act": "update",
            "obj": "{{.TeamObj}}/{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "helm-app:edit_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        },
        {
            "type": "p",
            "sub": "helm-app:edit_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1','helm-app'),
                                                                                                                                     ('view', '{
    "data": [
         {
            "type": "p",
            "sub": "helm-app:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "helm-app",
            "act": "get",
            "obj": "{{.TeamObj}}/{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "helm-app:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        },
        {
            "type": "p",
            "sub": "helm-app:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1','helm-app');


ALTER TABLE "public"."default_auth_role"
    ADD COLUMN access_type varchar(50);

UPDATE "public"."default_auth_role"
SET access_type = 'devtron-app'
WHERE role_type = 'manager'OR  role_type = 'trigger' OR  role_type = 'view'OR  role_type ='admin';


UPDATE "public"."default_auth_role"
SET access_type = 'cluster'
WHERE role_type = 'clusterEdit'OR role_type = 'clusterView'OR  role_type = 'clusterAdmin';


UPDATE "public"."default_auth_role"
SET access_type = 'globalEntity'
WHERE role_type = 'entitySpecificAdmin' OR role_type = 'entitySpecificView'OR role_type = 'roleSpecific';

SELECT setval('id_seq_default_auth_role', (SELECT MAX(id) FROM default_auth_role));



INSERT INTO "public"."default_auth_role" ( "role_type", "role", "created_on", "created_by", "updated_on", "updated_by","access_type") VALUES
                                                                                                                                 ( 'admin', '{
    "role": "helm-app:admin_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects":
    [
        "helm-app:admin_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "admin",
    "accessType": "helm-app"
}', 'now()', '1', 'now()', '1','helm-app'),
                                                                                                                                 ( 'edit', '{
   "role": "helm-app:edit_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects":
    [
        "helm-app:edit_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "edit",
    "accessType": "helm-app"
}', 'now()', '1', 'now()', '1','helm-app'),
                                                                                                                                 ( 'view', '{
     "role": "helm-app:view_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects":
    [
        "helm-app:view_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "view",
    "accessType": "helm-app"
}', 'now()', '1', 'now()', '1','helm-app');
