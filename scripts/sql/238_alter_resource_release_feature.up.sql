ALTER TABLE "public"."devtron_resource_object"
    ADD COLUMN IF NOT EXISTS identifier text;

ALTER TABLE devtron_resource_object_audit
    ADD COLUMN IF NOT EXISTS audit_operation_path text[];


ALTER TABLE devtron_resource_schema ALTER COLUMN version TYPE varchar(10);

INSERT INTO devtron_resource(kind, display_name, icon,is_exposed, parent_kind_id, deleted, created_on, created_by, updated_on,
                             updated_by)
VALUES ('release-track', 'Release track', '',false, 0, false, now(), 1, now(), 1),
       ('release', 'Release', '',false, 0, false, now(), 1, now(), 1);

INSERT INTO devtron_resource_schema(devtron_resource_id, version, schema, sample_schema, latest, created_on, created_by, updated_on,
                                    updated_by)
VALUES ((select id from devtron_resource where kind = 'release-track'), 'alpha1',
        '{
            "$schema": "https://json-schema.org/draft/2020-12/schema",
            "title": "Release track Schema",
            "type": "object",
            "properties":
            {
                "kind":
                {
                    "const": "release-track"
                },
                "version":
                {
                    "type": "string",
                    "enum": ["alpha1"]
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
                        "idType":{
                            "type": "string",
                            "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release track we keep data as devtron resource only. To differ between nature of these two types of id values.",
                            "enum": ["resourceObjectId", "oldObjectId"]
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
                        "idType"
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
    "title": "Release track Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "release-track"
        },
        "version":
        {
            "type": "string",
            "enum": ["alpha1"]
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
                "idType":{
                    "type": "string",
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release track we keep data as devtron resource only. To differ between nature of these two types of id values.",
                    "enum": ["resourceObjectId", "oldObjectId"]
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
                "idType"
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
}',
        true, now(), 1, now(), 1),
       ((select id from devtron_resource where kind = 'release'), 'alpha1',
        '{
            "$schema": "https://json-schema.org/draft/2020-12/schema",
            "title": "Release Schema",
            "type": "object",
            "properties":
            {
                "kind":
                {
                    "const": "release"
                },
                "version":
                {
                    "type": "string",
                    "enum": ["alpha1"]
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
                        "idType":{
                            "type": "string",
                            "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release we keep data as devtron resource only. To differ between nature of these two types of id values.",
                            "enum": ["resourceObjectId", "oldObjectId"]
                        },
                        "releaseVersion":
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
                        "note":
                        {
                            "type":"string"
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
                        },
                        "metadata":
                        {}
                    },
                    "required":
                    [
                        "id",
                        "releaseVersion"
                    ]
                },
                "status":
                {
                    "type": "object",
                    "properties":
                    {
                        "config":
                        {
                            "type": "object",
                            "properties": {
                                "status":
                                {
                                    "type":"string",
                                    "enum": [
                                        "draft",
                                        "readyForRelease",
                                        "hold"
                                    ]
                                },
                                "lock":
                                {
                                    "type": "boolean"
                                }
                            },
                            "required":
                            [
                                "status"
                            ]
                        }
                    },
                    "required":
                    [
                        "config"
                    ]
                },
                "taskMapping":
                {
                    "type": "array"
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
                "overview",
                "status"
            ]
        }', '{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Release Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "release"
        },
        "version":
        {
            "type": "string",
            "enum": ["alpha1"]
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
                "idType":{
                    "type": "string",
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release we keep data as devtron resource only. To differ between nature of these two types of id values.",
                    "enum": ["resourceObjectId", "oldObjectId"]
                },
                "releaseVersion":
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
                "note":
                {
                    "type":"string"
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
                },
                "metadata":
                {}
            },
            "required":
            [
                "id",
                "releaseVersion"
            ]
        },
        "status":
        {
            "type": "object",
            "properties":
            {
                "config":
                {
                    "type": "object",
                    "properties": {
                        "status":
                        {
                            "type":"string",
                            "enum": [
                                "draft",
                                "readyForRelease",
                                "hold"
                            ]
                        },
                        "lock":
                        {
                            "type": "boolean"
                        }
                    },
                    "required":
                    [
                        "status"
                    ]
                }
            },
            "required":
            [
                "config"
            ]
        },
        "taskMapping":
        {
            "type": "array"
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
        "overview",
        "status"
    ]
}',  true, now(), 1, now(), 1);


update devtron_resource_schema
set schema='{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Cluster Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "cluster"
        },
        "version":
        {
            "type": "string",
            "enum": ["v1"]
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
                "idType":{
                    "type": "string",
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release track we keep data as devtron resource only. To differ between nature of these two types of id values.",
                    "enum": ["resourceObjectId", "oldObjectId"]
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
                "metadata"
            ]
        },
        "actions":
        {
            "type": "array"
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
}'
where devtron_resource_id = (select id from devtron_resource where kind = 'cluster')
  and version = 'v1';

update devtron_resource_schema
set schema='{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Job Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "job"
        },
        "version":
        {
            "type": "string",
            "enum": ["v1"]
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
                "idType":{
                    "type": "string",
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release track we keep data as devtron resource only. To differ between nature of these two types of id values.",
                    "enum": ["resourceObjectId", "oldObjectId"]
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
                                "POCs":
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
                                "Access manager":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "object",
                                        "refType": "#/references/users"
                                    }
                                }
                            },
                            "required":
                            [
                                "Owner"
                            ]
                        },
                        "About job":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Type of job":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Deployment",
                                        "Migration",
                                        "Backup",
                                        "Others"
                                    ]
                                },
                                "For environment":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "Production",
                                        "Dev",
                                        "Staging",
                                        "QA",
                                        "UAT"
                                    ]
                                },
                                "Documentation":
                                {
                                    "type": "string",
                                    "format": "uri"
                                }
                            }
                        },
                        "Operational schedule":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Preferred run":
                                {
                                    "type": "string"
                                },
                                "Maintenance time":
                                {
                                    "type": "string"
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
                "metadata"
            ]
        },
        "actions":
        {
            "type": "array"
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
}'
where devtron_resource_id = (select id from devtron_resource where kind = 'job')
  and version = 'v1';

update devtron_resource_schema
set schema='{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Devtron Application Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "application/devtron-application"
        },
        "version":
        {
            "type": "string",
            "enum": ["v1"]
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
                "idType":{
                    "type": "string",
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release track we keep data as devtron resource only. To differ between nature of these two types of id values.",
                    "enum": ["resourceObjectId", "oldObjectId"]
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
                },
                "metadata":
                {
                    "type": "object",
                    "properties":
                    {
                        "Owners & Pager Duty":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Code owners":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "minItems": 1,
                                    "items":
                                    {
                                        "type": "object",
                                        "refType": "#/references/users"
                                    }
                                },
                                "On pager duty":
                                {
                                    "type": "object",
                                    "refType": "#/references/users"
                                }
                            },
                            "required":
                            [
                                "Code owners"
                            ]
                        },
                        "Service details":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Framework":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "string",
                                        "enum":
                                        [
                                            "Django",
                                            "Ruby on Rails",
                                            "Laravel",
                                            "Angular",
                                            "React",
                                            "jQuery",
                                            "ASP.NET Core",
                                            "Bootstrap"
                                        ]
                                    }
                                },
                                "Language":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "string",
                                        "enum":
                                        [
                                            "Java",
                                            "Python",
                                            "PHP",
                                            "Go",
                                            "Ruby",
                                            "Node"
                                        ]
                                    }
                                },
                                "Communication method":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "GraphQL",
                                        "gRPC",
                                        "Message Queue",
                                        "NATS",
                                        "REST API",
                                        "WebSocket"
                                    ]
                                },
                                "Internet facing":
                                {
                                    "type": "boolean"
                                }
                            }
                        },
                        "Documentation":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Service Documentation":
                                {
                                    "type": "string",
                                    "format": "uri"
                                },
                                "API Contract":
                                {
                                    "type": "string",
                                    "format": "uri"
                                },
                                "Runbook":
                                {
                                    "type": "string",
                                    "format": "uri"
                                }
                            }
                        }
                    },
                    "required":
                    [
                        "Owners & Pager Duty"
                    ]
                }
            },
            "required":
            [
                "id",
                "metadata"
            ]
        },
        "actions":
        {
            "type": "array"
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
}'
where devtron_resource_id = (select id from devtron_resource where kind = 'devtron-application')
  and version = 'v1';

update devtron_resource_schema
set schema='{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Helm Application Schema",
    "type": "object",
    "properties":
    {
        "kind":
        {
            "const": "application/helm-application"
        },
        "version":
        {
            "type": "string",
            "enum": ["v1"]
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
                "idType":{
                    "type": "string",
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release track we keep data as devtron resource only. To differ between nature of these two types of id values.",
                    "enum": ["resourceObjectId", "oldObjectId"]
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
                },
                "metadata":
                {
                    "type": "object",
                    "properties":
                    {
                        "Owners & Pager Duty":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Code owners":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "minItems": 1,
                                    "items":
                                    {
                                        "type": "object",
                                        "refType": "#/references/users"
                                    }
                                },
                                "On pager duty":
                                {
                                    "type": "object",
                                    "refType": "#/references/users"
                                }
                            },
                            "required":
                            [
                                "Code owners"
                            ]
                        },
                        "Service details":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Framework":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "string",
                                        "enum":
                                        [
                                            "Django",
                                            "Ruby on Rails",
                                            "Laravel",
                                            "Angular",
                                            "React",
                                            "jQuery",
                                            "ASP.NET Core",
                                            "Bootstrap"
                                        ]
                                    }
                                },
                                "Language":
                                {
                                    "type": "array",
                                    "uniqueItems": true,
                                    "items":
                                    {
                                        "type": "string",
                                        "enum":
                                        [
                                            "Java",
                                            "Python",
                                            "PHP",
                                            "Go",
                                            "Ruby",
                                            "Node"
                                        ]
                                    }
                                },
                                "Communication method":
                                {
                                    "type": "string",
                                    "enum":
                                    [
                                        "GraphQL",
                                        "gRPC",
                                        "Message Queue",
                                        "NATS",
                                        "REST API",
                                        "WebSocket"
                                    ]
                                },
                                "Internet facing":
                                {
                                    "type": "boolean"
                                }
                            }
                        },
                        "Documentation":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Service Documentation":
                                {
                                    "type": "string",
                                    "format": "uri"
                                },
                                "API Contract":
                                {
                                    "type": "string",
                                    "format": "uri"
                                },
                                "Runbook":
                                {
                                    "type": "string",
                                    "format": "uri"
                                }
                            }
                        }
                    },
                    "required":
                    [
                        "Owners & Pager Duty"
                    ]
                }
            },
            "required":
            [
                "id",
                "metadata"
            ]
        },
        "actions":
        {
            "type": "array"
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
}'
where devtron_resource_id = (select id from devtron_resource where kind = 'helm-application')
  and version = 'v1';