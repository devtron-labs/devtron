-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_default_auth_policy;

-- Table Definition
CREATE TABLE "public"."default_auth_policy" (
                                          "id" int NOT NULL DEFAULT nextval('id_seq_default_auth_policy'::regclass),
                                          "role_type" varchar(250) NOT NULL,
                                          "policy" text NOT NULL,
                                          "created_on" timestamptz,
                                          "created_by" integer,
                                          "updated_on" timestamptz,
                                          "updated_by" integer,
                                          PRIMARY KEY ("id")
);

INSERT INTO "public"."default_auth_policy" ("id", "role_type", "policy", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'manager', '{
    "data": [
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "*",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "*",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "*",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "user",
            "act": "*",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "notification",
            "act": "*",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "*",
            "obj": "{{.EnvObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('2', 'admin', '{
    "data": [
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "*",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "*",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('3', 'trigger', '{
    "data": [
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "get",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "trigger",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "trigger",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "get",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('4', 'view', '{
    "data": [
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "get",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "get",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        },
        {
            "type": "p",
            "sub": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('5', 'entityAll', '{
    "data": [
        {
            "type": "p",
            "sub": "role:{{.Entity}}_admin",
            "res": "{{.Entity}}",
            "act": "*",
            "obj": "*"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('6', 'entityView','{
    "data": [
        {
            "type": "p",
            "sub": "role:{{.Entity}}_view",
            "res": "{{.Entity}}",
            "act": "get",
            "obj": "*"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('7', 'entitySpecific','{
    "data": [
        {
            "type": "p",
            "sub": "role:{{.Entity}}_{{.EntityName}}_specific",
            "res": "{{.Entity}}",
            "act": "update",
            "obj": "{{.EntityName}}"
        },
       {
            "type": "p",
            "sub": "role:{{.Entity}}_{{.EntityName}}_specific",
            "res": "{{.Entity}}",
            "act": "get",
            "obj": "{{.EntityName}}"
        }
    ]
}', 'now()', '1', 'now()', '1');




-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_default_auth_role;

-- Table Definition
CREATE TABLE "public"."default_auth_role" (
                                        "id" int NOT NULL DEFAULT nextval('id_seq_default_auth_role'::regclass),
                                        "role_type" varchar(250) NOT NULL,
                                        "role" text NOT NULL,
                                        "created_on" timestamptz,
                                        "created_by" integer,
                                        "updated_on" timestamptz,
                                        "updated_by" integer,
                                        PRIMARY KEY ("id")
);

INSERT INTO "public"."default_auth_role" ("id", "role_type", "role", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'manager', '{
    "role": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:manager_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "manager",
    "access_type": ""
}', 'now()', '1', 'now()', '1'),
('2', 'admin', '{
    "role": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:admin_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "admin",
    "access_type": ""
}', 'now()', '1', 'now()', '1'),
('3', 'trigger', '{
    "role": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:trigger_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "trigger",
    "access_type": ""
}', 'now()', '1', 'now()', '1'),
('4', 'view', '{
    "role": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:view_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "action": "view",
    "access_type": ""
}', 'now()', '1', 'now()', '1'),
('5', 'entitySpecificAdmin', '{
    "role": "role:{{.Entity}}_admin",
    "casbinSubjects": [
        "role:{{.Entity}}_admin"
    ],
    "entity": "{{.Entity}}",
    "team": "",
    "application": "",
    "environment": "",
    "action": "admin"
}', 'now()', '1', 'now()', '1'),
('6', 'entitySpecificView', '{
    "role": "role:{{.Entity}}_view",
    "casbinSubjects": [
        "role:{{.Entity}}_view"
    ],
    "entity": "{{.Entity}}",
    "team": "",
    "application": "",
    "environment": "",
    "action": "view"
}', 'now()', '1', 'now()', '1'),
('7', 'roleSpecific', '{
    "role": "role:{{.Entity}}_{{.EntityName}}_specific",
    "casbinSubjects": [
        "role:{{.Entity}}_{{.EntityName}}_specific"
    ],
    "entity": "{{.Entity}}",
    "team": "",
    "entityName": "{{.EntityName}}",
    "environment": "",
    "action": "update"
}', 'now()', '1', 'now()', '1');