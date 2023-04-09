alter table pipeline add column user_approval_config character varying(1000);

CREATE SEQUENCE IF NOT EXISTS id_seq_deployment_approval_request;

CREATE TABLE public.deployment_approval_request (
"id"                                            integer NOT NULL DEFAULT nextval('id_seq_deployment_approval_request'::regclass),
"pipeline_id"                                   integer,
"artifact_id"                                   integer,
"active"                                        BOOL,
"artifact_deployment_triggered"                 BOOL DEFAULT false,
"created_on"                                    timestamptz,
"created_by"                                    int4,
"updated_on"                                    timestamptz,
"updated_by"                                    int4,
CONSTRAINT "deployment_approval_request_pipeline_id_fkey" FOREIGN KEY ("pipeline_id") REFERENCES "public"."pipeline" ("id"),
CONSTRAINT "deployment_approval_request_artifact_id_fkey" FOREIGN KEY ("artifact_id") REFERENCES "public"."ci_artifact" ("id"),
PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_deployment_approval_user_data;

CREATE TABLE public.deployment_approval_user_data (
"id"                                            integer NOT NULL DEFAULT nextval('id_seq_deployment_approval_user_data'::regclass),
"approval_request_id"                           integer,
"user_id"                                       integer,
"user_response"                                 integer,
"comments"                                      character varying(1000),
"created_on"                                    timestamptz,
"created_by"                                    int4,
"updated_on"                                    timestamptz,
"updated_by"                                    int4,
CONSTRAINT "deployment_approval_user_data_approval_request_id_fkey" FOREIGN KEY ("approval_request_id") REFERENCES "public"."deployment_approval_request" ("id"),
CONSTRAINT "deployment_approval_user_data_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id"),
PRIMARY KEY ("id"),
UNIQUE("approval_request_id","user_id")
);

alter table cd_workflow_runner add column deployment_approval_request_id int CONSTRAINT "cd_workflow_runner_deployment_approval_request_id_fkey" REFERENCES "public"."deployment_approval_request" ("id");

Alter table roles add column approver bool;

DROP INDEX "public"."role_unique";

CREATE UNIQUE INDEX IF NOT EXISTS "role_unique" ON "public"."roles" USING BTREE ("role","access_type","approver");

update default_auth_role
set role = '{
    "role": "role:manager_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:manager_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "approver":{{.Approver}},
    "action": "manager",
    "accessType": "devtron-app"
}' where id = 1;

update default_auth_role
set role = '{
    "role": "role:admin_{{.Team}}_{{.Env}}_{{.App}}",
     "casbinSubjects": [
         "role:admin_{{.Team}}_{{.Env}}_{{.App}}"
     ],
     "team": "{{.Team}}",
     "entityName": "{{.App}}",
     "environment": "{{.Env}}",
     "approver":{{.Approver}},
     "action": "admin",
     "accessType": "devtron-app"
}' where id = 2;

update default_auth_role
set role = '{
    "role": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
     "casbinSubjects": [
         "role:trigger_{{.Team}}_{{.Env}}_{{.App}}"
     ],
     "team": "{{.Team}}",
     "entityName": "{{.App}}",
     "environment": "{{.Env}}",
     "approver":{{.Approver}},
     "action": "trigger",
     "accessType": "devtron-app"
}' where id = 3;

update default_auth_role
set role = '{
    "role": "role:view_{{.Team}}_{{.Env}}_{{.App}}",
    "casbinSubjects": [
        "role:view_{{.Team}}_{{.Env}}_{{.App}}"
    ],
    "team": "{{.Team}}",
    "entityName": "{{.App}}",
    "environment": "{{.Env}}",
    "approver":{{.Approver}},
    "action": "view",
    "accessType": "devtron-app"
}' where id = 4;
