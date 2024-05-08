

UPDATE "public"."default_auth_role"
SET role= '{
    "role": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:manager_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "manager",
    "entity": "{{.Entity}}",
    "accessType": "devtron-app"
}'
WHERE role_type='manager' AND id=1;

UPDATE "public"."default_auth_role"
SET role= '{
    "role": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:admin_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "admin",
    "entity": "{{.Entity}}",
    "accessType": "devtron-app"
}'
WHERE role_type='admin' AND id =2;

UPDATE "public"."default_auth_role"
SET role= '{
    "role": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:trigger_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "trigger",
    "entity": "{{.Entity}}",
    "accessType": "devtron-app"
}'
WHERE role_type='trigger' AND id =3;

UPDATE "public"."default_auth_role"
SET role= '{
    "role": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:view_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "view",
    "entity": "{{.Entity}}",
    "accessType": "devtron-app"
}'
WHERE role_type='view' AND id =4;

ALTER TABLE "public"."default_auth_policy"
    ADD COLUMN access_type varchar(50);

ALTER TABLE "public"."default_auth_policy"
    ADD COLUMN entity varchar(50);

UPDATE "public"."default_auth_policy"
SET entity = 'apps', access_type ='devtron-app'
WHERE role_type = 'manager'OR  role_type = 'trigger' OR  role_type = 'view'OR  role_type ='admin';

UPDATE "public"."default_auth_policy"
SET entity = 'cluster', role_type='edit'
WHERE role_type = 'clusterEdit';

UPDATE "public"."default_auth_policy"
SET entity = 'cluster', role_type='view'
WHERE role_type = 'clusterView';

UPDATE "public"."default_auth_policy"
SET entity = 'cluster', role_type='admin'
WHERE role_type = 'clusterAdmin';

UPDATE "public"."default_auth_policy"
SET entity = 'chart-group', role_type='update'
WHERE role_type = 'entitySpecific';

UPDATE "public"."default_auth_policy"
SET entity = 'chart-group', role_type='view'
WHERE role_type = 'entityView';

UPDATE "public"."default_auth_policy"
SET entity = 'chart-group', role_type='admin'
WHERE role_type = 'entityAll';


SELECT setval('id_seq_default_auth_policy', (SELECT MAX(id) FROM default_auth_policy));

INSERT INTO "public"."default_auth_policy" ( "role_type", "policy", "created_on", "created_by", "updated_on", "updated_by","access_type","entity") VALUES
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
}', 'now()', '1', 'now()', '1','helm-app','apps'),
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
}', 'now()', '1', 'now()', '1','helm-app','apps'),
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
}', 'now()', '1', 'now()', '1','helm-app','apps');


ALTER TABLE "public"."default_auth_role"
    ADD COLUMN access_type varchar(50);

ALTER TABLE "public"."default_auth_role"
    ADD COLUMN entity varchar(50);

UPDATE "public"."default_auth_role"
SET access_type = 'devtron-app' , entity  ='apps'
WHERE role_type = 'manager'OR  role_type = 'trigger' OR  role_type = 'view'OR  role_type ='admin';



UPDATE "public"."default_auth_role"
SET entity = 'cluster', role_type='edit'
WHERE role_type = 'clusterEdit';

UPDATE "public"."default_auth_role"
SET entity = 'cluster', role_type='view'
WHERE role_type = 'clusterView';

UPDATE "public"."default_auth_role"
SET entity = 'cluster', role_type='admin'
WHERE role_type = 'clusterAdmin';



UPDATE "public"."default_auth_role"
SET entity = 'chart-group', role_type='view'
WHERE role_type = 'entitySpecificView';

UPDATE "public"."default_auth_role"
SET entity = 'chart-group', role_type='update'
WHERE role_type = 'roleSpecific';

UPDATE "public"."default_auth_role"
SET entity = 'chart-group', role_type='admin'
WHERE role_type = 'entitySpecificAdmin';



SELECT setval('id_seq_default_auth_role', (SELECT MAX(id) FROM default_auth_role));



INSERT INTO "public"."default_auth_role" ( "role_type", "role", "created_on", "created_by", "updated_on", "updated_by","access_type","entity") VALUES
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
    "entity": "{{.Entity}}",
    "accessType": "helm-app"
}', 'now()', '1', 'now()', '1','helm-app','apps'),
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
    "entity": "{{.Entity}}",
    "accessType": "helm-app"
}', 'now()', '1', 'now()', '1','helm-app','apps'),
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
    "entity": "{{.Entity}}",
    "accessType": "helm-app"
}', 'now()', '1', 'now()', '1','helm-app','apps');


