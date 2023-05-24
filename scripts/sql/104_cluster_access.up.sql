INSERT INTO "public"."default_auth_policy" ("id", "role_type", "policy", "created_on", "created_by", "updated_on", "updated_by") VALUES
('8', 'clusterAdmin', '{
    "data": [
        {
            "type": "p",
            "sub": "role:clusterAdmin_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}",
            "res": "{{.ClusterObj}}/{{.NamespaceObj}}",
            "act": "*",
            "obj": "{{.GroupObj}}/{{.KindObj}}/{{.ResourceObj}}"
        },
        {
            "type": "p",
            "sub": "role:clusterAdmin_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}",
            "res": "{{.ClusterObj}}/{{.NamespaceObj}}/user",
            "act": "*",
            "obj": "{{.GroupObj}}/{{.KindObj}}/{{.ResourceObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('9', 'clusterEdit', '{
    "data": [
        {
            "type": "p",
            "sub": "role:clusterEdit_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}",
            "res": "{{.ClusterObj}}/{{.NamespaceObj}}",
            "act": "*",
            "obj": "{{.GroupObj}}/{{.KindObj}}/{{.ResourceObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1'),
('10', 'clusterView', '{
    "data": [
        {
            "type": "p",
            "sub": "role:clusterView_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}",
            "res": "{{.ClusterObj}}/{{.NamespaceObj}}",
            "act": "get",
            "obj": "{{.GroupObj}}/{{.KindObj}}/{{.ResourceObj}}"
        }
    ]
}', 'now()', '1', 'now()', '1');


ALTER TABLE "roles"
    ADD COLUMN "cluster" text,
    ADD COLUMN "namespace" text,
    ADD COLUMN "group" text,
    ADD COLUMN "kind" text,
    ADD COLUMN "resource" text;



INSERT INTO "public"."default_auth_role" ("id", "role_type", "role", "created_on", "created_by", "updated_on", "updated_by") VALUES
('8', 'clusterAdmin', '{
    "role": "role:clusterAdmin_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}",
    "casbinSubjects": [
        "role:role:clusterAdmin_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}"
    ],
    "entity": "{{.Entity}}",
    "cluster": "{{.Cluster}}",
    "namespace": "{{.Namespace}}",
    "group": "{{.Group}}",
    "kind": "{{.Kind}}",
    "resource": "{{.Resource}}",
    "action": "admin",
    "access_type": ""
}', 'now()', '1', 'now()', '1'),
('9', 'clusterEdit', '{
    "role": "role:clusterEdit_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}",
    "casbinSubjects": [
        "role:clusterEdit_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}"
    ],
    "entity": "{{.Entity}}",
    "cluster": "{{.Cluster}}",
    "namespace": "{{.Namespace}}",
    "group": "{{.Group}}",
    "kind": "{{.Kind}}",
    "resource": "{{.Resource}}",
    "action": "edit",
    "access_type": ""
}', 'now()', '1', 'now()', '1'),
('10', 'clusterView', '{
    "role": "role:clusterView_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}",
    "casbinSubjects": [
        "role:clusterView_{{.Cluster}}_{{.Namespace}}_{{.Group}}_{{.Kind}}_{{.Resource}}"
    ],
    "entity": "{{.Entity}}",
    "cluster": "{{.Cluster}}",
    "namespace": "{{.Namespace}}",
    "group": "{{.Group}}",
    "kind": "{{.Kind}}",
    "resource": "{{.Resource}}",
    "action": "view",
    "access_type": ""
}', 'now()', '1', 'now()', '1');