CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_object_dep_relations;

CREATE TABLE IF NOT EXISTS "public"."devtron_resource_object_dep_relations"
(
    "id"                             int NOT NULL DEFAULT nextval('id_seq_devtron_resource_object_dep_relations'::regclass),
    "component_dt_res_object_id"     int,
    "component_dt_res_schema_id"     int,
    "dependency_dt_res_object_id"    int,
    "dependency_dt_res_schema_id"    int,
    "type_of_dependency"             VARCHAR(50),
    "created_on"                     timestamptz,
    "created_by"                     integer,
    "updated_on"                     timestamptz,
    "updated_by"                     integer,
    PRIMARY KEY ("id"),
    CONSTRAINT "dep_mapping_component_object_id_fk" FOREIGN KEY ("component_dt_res_object_id") REFERENCES "public"."devtron_resource_object" ("id"),
    CONSTRAINT "dep_mapping_component_schema_id_fk" FOREIGN KEY ("component_dt_res_schema_id") REFERENCES "public"."devtron_resource_schema" ("id"),
    CONSTRAINT "dep_mapping_dependency_object_id_fk" FOREIGN KEY ("component_dt_res_object_id") REFERENCES "public"."devtron_resource_object" ("id"),
    CONSTRAINT "dep_mapping_dependency_schema_id_fk" FOREIGN KEY ("dependency_dt_res_schema_id") REFERENCES "public"."devtron_resource_schema" ("id")
);


INSERT INTO devtron_resource(kind, display_name, icon,is_exposed, parent_kind_id, deleted, created_on, created_by, updated_on,
                             updated_by)
VALUES ('tenant', 'Tenant', '',true, 0, false, now(), 1, now(), 1),
       ('installation', 'Installation', '',false, 0, false, now(), 1, now(), 1),
       ('environment', 'Environment', '',false, 0, false, now(), 1, now(), 1);

INSERT INTO devtron_resource_schema(devtron_resource_id, version, schema, sample_schema, latest, created_on, created_by, updated_on,
                                    updated_by)
VALUES ((select id from devtron_resource where kind = 'tenant'), 'alpha1',
        '{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Tenant Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "tenant"
        },
        "version":
        {
            "type": "string",
            "enum":
            [
                "alpha1"
            ]
        },
        "overview":
        {
            "type": "object",
            "properties":
            {
                "id":
                {
                    "type": "number"
                },
                "idType":
                {
                    "type": "string",
                    "enum":
                    [
                        "resourceObjectId",
                        "oldObjectId"
                    ]
                },
                "name":
                {
                    "type": "string"
                },
                "icon":
                {
                    "type": "string",
                    "format": "uri"
                },
                "description":
                {
                    "type": "string"
                },
                "tenantId":
                {
                    "type": "string"
                },
                "createdOn":
                {
                    "type": "string"
                },
                "createdBy":
                {
                    "type": "object",
                    "refType": "#/references/users"
                },
                "tags":
                {
                    "additionalProperties":
                    {
                        "type": "string"
                    }
                },
                "metadata":
                {
                    "type": "object",
                    "properties":
                    {
                        "Contacts":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Owner":
                                {
                                    "type": "object",
                                    "refType": "#/references/users"
                                },
                                "On pager duty":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "object",
                                        "refType": "#/references/users"
                                    }
                                },
                                "Team":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Growth team",
                                        "Support team",
                                        "Platform team",
                                        "Operations team"
                                    ]
                                },
                                "3rd party contacts":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "string"
                                    }
                                }
                            },
                            "required":
                            [
                                "Owner"
                            ]
                        },
                        "Networking & Others":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Cluster type":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Production",
                                        "Non production"
                                    ]
                                },
                                "Exposed to":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Public",
                                        "Private"
                                    ]
                                },
                                "VPC peered":
                                {
                                    "type": "array",
                                    "items":
                                    {
                                        "type": "string"
                                    }
                                },
                                "Documentation":
                                {
                                    "type": "string",
                                    "format": "uri"
                                }
                            }
                        },
                        "Backup":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Backup strategy":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Full backup",
                                        "Incremental",
                                        "Snapshot"
                                    ]
                                },
                                "Backup retention policy (days)":
                                {
                                    "type": "integer"
                                }
                            }
                        }
                    },
                    "required":
                    [
                        "Contacts"
                    ]
                }
            },
            "required":
            [
                "id",
                "tenantId"
            ]
        },
        "dependencies":
        {
            "type": "array"
        }
    },
    "required":
    [
        "version",
        "kind",
        "overview"
    ]
}','{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Tenant Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "tenant"
        },
        "version":
        {
            "type": "string",
            "enum":
            [
                "alpha1"
            ]
        },
        "overview":
        {
            "type": "object",
            "properties":
            {
                "id":
                {
                    "type": "number"
                },
                "idType":
                {
                    "type": "string",
                    "enum":
                    [
                        "resourceObjectId",
                        "oldObjectId"
                    ]
                },
                "name":
                {
                    "type": "string"
                },
                "icon":
                {
                    "type": "string",
                    "format": "uri"
                },
                "description":
                {
                    "type": "string"
                },
                "tenantId":
                {
                    "type": "string"
                },
                "createdOn":
                {
                    "type": "string"
                },
                "createdBy":
                {
                    "type": "object",
                    "refType": "#/references/users"
                },
                "tags":
                {
                    "additionalProperties":
                    {
                        "type": "string"
                    }
                },
                "metadata":
                {
                    "type": "object",
                    "properties":
                    {
                        "Contacts":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Owner":
                                {
                                    "type": "object",
                                    "refType": "#/references/users"
                                },
                                "On pager duty":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "object",
                                        "refType": "#/references/users"
                                    }
                                },
                                "Team":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Growth team",
                                        "Support team",
                                        "Platform team",
                                        "Operations team"
                                    ]
                                },
                                "3rd party contacts":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "string"
                                    }
                                }
                            },
                            "required":
                            [
                                "Owner"
                            ]
                        },
                        "Networking & Others":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Cluster type":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Production",
                                        "Non production"
                                    ]
                                },
                                "Exposed to":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Public",
                                        "Private"
                                    ]
                                },
                                "VPC peered":
                                {
                                    "type": "array",
                                    "items":
                                    {
                                        "type": "string"
                                    }
                                },
                                "Documentation":
                                {
                                    "type": "string",
                                    "format": "uri"
                                }
                            }
                        },
                        "Backup":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Backup strategy":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Full backup",
                                        "Incremental",
                                        "Snapshot"
                                    ]
                                },
                                "Backup retention policy (days)":
                                {
                                    "type": "integer"
                                }
                            }
                        }
                    },
                    "required":
                    [
                        "Contacts"
                    ]
                }
            },
            "required":
            [
                "id",
                "tenantId"
            ]
        },
        "dependencies":
        {
            "type": "array"
        }
    },
    "required":
    [
        "version",
        "kind",
        "overview"
    ]
}',true, now(), 1, now(), 1),
       ((select id from devtron_resource where kind = 'installation'), 'alpha1',
        '{
            "$schema": "https://json-schema.org/draft/2020-12/schema",
            "title": "Installation Schema",
            "type": "object",
            "properties":
            {
                "kind":
                {
                    "const": "installation"
                },
                "version":
                {
                    "type": "string",
                    "enum":
                    [
                        "alpha1"
                    ]
                },
                "overview":
                {
                    "type": "object",
                    "properties":
                    {
                        "id":
                        {
                            "type": "number"
                        },
                        "idType":
                        {
                            "type": "string",
                            "enum":
                            [
                                "resourceObjectId",
                                "oldObjectId"
                            ]
                        },
                        "installationId":
                        {
                            "type": "string"
                        },
                        "name":
                        {
                            "type": "string"
                        },
                        "icon":
                        {
                            "type": "string",
                            "contentEncoding": "base64"
                        },
                        "description":
                        {
                            "type": "string"
                        },
                        "createdOn":
                        {
                            "type": "string"
                        },
                        "createdBy":
                        {
                            "type": "object",
                            "refType": "#/references/users"
                        },
                        "tags":
                        {
                            "additionalProperties":
                            {
                                "type": "string"
                            }
                        }
                    },
                    "required":
                    [
                        "id",
                        "idType",
                        "installationId"
                    ]
                },
                "dependencies":
                {
                    "type": "array"
                }
            },
            "required":
            [
                "version",
                "kind",
                "overview"
            ]
        }', '{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Installation Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "installation"
        },
        "version":
        {
            "type": "string",
            "enum":
            [
                "alpha1"
            ]
        },
        "overview":
        {
            "type": "object",
            "properties":
            {
                "id":
                {
                    "type": "number"
                },
                "idType":
                {
                    "type": "string",
                    "enum":
                    [
                        "resourceObjectId",
                        "oldObjectId"
                    ]
                },
                "installationId":
                {
                    "type": "string"
                },
                "name":
                {
                    "type": "string"
                },
                "icon":
                {
                    "type": "string",
                    "contentEncoding": "base64"
                },
                "description":
                {
                    "type": "string"
                },
                "createdOn":
                {
                    "type": "string"
                },
                "createdBy":
                {
                    "type": "object",
                    "refType": "#/references/users"
                },
                "tags":
                {
                    "additionalProperties":
                    {
                        "type": "string"
                    }
                }
            },
            "required":
            [
                "id",
                "idType",
                "installationId"
            ]
        },
        "dependencies":
        {
            "type": "array"
        }
    },
    "required":
    [
        "version",
        "kind",
        "overview"
    ]
}',  true, now(), 1, now(), 1),
((select id from devtron_resource where kind = 'environment'), 'v1',
        '{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Environment Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "environment"
        },
        "version":
        {
            "type": "string",
            "enum":
            [
                "v1"
            ]
        },
        "overview":
        {
            "type": "object"
        },
        "dependencies":
        {
            "type": "array"
        }
    },
    "required":
    [
        "version",
        "kind"
    ]
}','{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Environment Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "environment"
        },
        "version":
        {
            "type": "string",
            "enum":
            [
                "v1"
            ]
        },
        "overview":
        {
            "type": "object"
        },
        "dependencies":
        {
            "type": "array"
        }
    },
    "required":
    [
        "version",
        "kind"
    ]
}', true, now(), 1, now(), 1);
