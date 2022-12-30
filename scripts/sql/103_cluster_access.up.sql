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