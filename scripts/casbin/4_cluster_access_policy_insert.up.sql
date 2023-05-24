INSERT INTO "public"."casbin_rule" ("p_type", "v0", "v1", "v2", "v3", "v4", "v5")
    VALUES ('p', 'role:super-admin___', '*/*', '*', '*/*/*', 'allow', ''), --- v1=cluster/ns, v2=group/kind/resource
           ('p', 'role:super-admin___', '*/*/user', '*', '*/*/*', 'allow', '') ; --- v1=cluster/ns/user, v2=group/kind/resource