CREATE SEQUENCE IF NOT EXISTS id_seq_terminal_access_templates;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."terminal_access_templates"
(
    "id"            integer NOT NULL DEFAULT nextval('id_seq_terminal_access_templates'::regclass),
    "template_name" VARCHAR(1000),
    "template_data" text,
    "created_on"    timestamptz,
    "created_by"    int4,
    "updated_on"    timestamptz,
    "updated_by"    int4,
    PRIMARY KEY ("id")
);

ALTER TABLE ONLY public.terminal_access_templates
    ADD CONSTRAINT terminal_access_template_name_unique UNIQUE (template_name);


CREATE SEQUENCE IF NOT EXISTS id_seq_user_terminal_access_data;

-- Table Definition
CREATE TABLE  IF NOT EXISTS "public"."user_terminal_access_data"
(
    "id"         integer NOT NULL DEFAULT nextval('id_seq_user_terminal_access_data'::regclass),
    "user_id"    int4,
    "cluster_id" integer,
    "pod_name"   VARCHAR(1000),
    "node_name"  VARCHAR(1000),
    "status"     VARCHAR(1000),
    "metadata"   json,
    "created_on" timestamptz,
    "created_by" int4,
    "updated_on" timestamptz,
    "updated_by" int4,
    PRIMARY KEY ("id")
);

ALTER TABLE "public"."user_terminal_access_data"
    ADD FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id");

ALTER TABLE "public"."user_terminal_access_data"
    ADD FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster" ("id");


INSERT into terminal_access_templates(template_name, template_data, created_on, created_by, updated_on, updated_by) VALUES
('terminal-access-service-account','{"apiVersion":"v1","kind":"ServiceAccount","metadata":{"name":"${pod_name}-sa","namespace":"${default_namespace}"}}', now(), 1, now(), 1),
('terminal-access-role-binding','{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"ClusterRoleBinding","metadata":{"name":"${pod_name}-crb"},"subjects":[{"kind":"ServiceAccount","name":"${pod_name}-sa","namespace":"${default_namespace}"}],"roleRef":{"kind":"ClusterRole","name":"cluster-admin","apiGroup":"rbac.authorization.k8s.io"}}', now(), 1, now(), 1),
('terminal-access-pod','{"apiVersion":"v1","kind":"Pod","metadata":{"name":"${pod_name}"},"spec":{"serviceAccountName":"${pod_name}-sa","nodeSelector":{"kubernetes.io/hostname":"${node_name}"},"containers":[{"name":"devtron-debug-terminal","image":"${base_image}","command":["/bin/sh","-c","--"],"args":["while true; do sleep 600; done;"]}],"tolerations":[{"key":"kubernetes.azure.com/scalesetpriority","operator":"Equal","value":"spot","effect":"NoSchedule"}]}}', now(), 1, now(), 1);

INSERT INTO attributes(key, value, active, created_on, created_by, updated_on, updated_by) VALUES ('DEFAULT_TERMINAL_IMAGE_LIST', 'quay.io/devtron/ubuntu-k8s-utils:latest,quay.io/devtron/alpine-k8s-utils:latest,quay.io/devtron/centos-k8s-utils:latest,quay.io/devtron/alpine-netshoot:latest', 't', NOW(), 1,NOW(), 1);