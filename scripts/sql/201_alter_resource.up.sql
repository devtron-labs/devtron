INSERT INTO devtron_resource(kind, display_name, icon, parent_kind_id, deleted, created_on, created_by, updated_on,
                             updated_by)
VALUES ('cd-pipeline', 'Cd Pipeline', '', 0, false, now(), 1, now(), 1);


INSERT INTO devtron_resource_schema(devtron_resource_id, version, schema, latest, created_on, created_by, updated_on,
                                    updated_by)
VALUES ((select id from devtron_resource where kind = 'cd-pipeline'), 'v1',
        '{
          "$schema": "https://json-schema.org/draft/2020-12/schema",
          "title": "CD Pipeline Schema",
          "type": "object",
          "properties": {
            "version": {
              "type": "string"
            },
            "kind": {
              "type": "string"
            },
            "overview": {
              "type": "object"
            },
            "actions": {
              "type": "array"
            },
            "dependencies": {
              "type": "array"
            }
          },
          "required": [
            "version",
            "kind"
          ]
        }',
        true, now(), 1, now(), 1);

update devtron_resource_schema
set schema='{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Cluster Schema",
    "type": "object",
    "properties":
    {
        "version":
        {
            "type": "string"
        },
        "kind":
        {
            "type": "string"
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
        "version":
        {
            "type": "string"
        },
        "kind":
        {
            "type": "string"
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
        "version":
        {
            "type": "string"
        },
        "kind":
        {
            "type": "string"
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
        "version":
        {
            "type": "string"
        },
        "kind":
        {
            "type": "string"
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