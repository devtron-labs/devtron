ALTER TABLE devtron_resource_object_audit
    ADD COLUMN audit_operation_path text;


INSERT INTO devtron_resource(kind, display_name, icon, parent_kind_id, deleted, created_on, created_by, updated_on,
                             updated_by)
VALUES ('release-track', 'Release track', '', 0, false, now(), 1, now(), 1),
       ('release', 'Release', '', 0, false, now(), 1, now(), 1);

INSERT INTO devtron_resource_schema(devtron_resource_id, version, schema, latest, created_on, created_by, updated_on,
                                    updated_by)
VALUES ((select id from devtron_resource where kind = 'release-track'), 'alpha1',
'{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Release track Schema",
    "type": "object",
    "properties":
    {
        "version":
        {
            "const": "release-track"
        },
        "kind":
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
                    "enum": ["resourceId", "oldObjectId"]
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
        "version":
        {
            "const": "release"
        },
        "kind":
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
                    "enum": ["resourceId", "oldObjectId"]
                },
                "releaseVersion":
                {
                    "type": "string",
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
                },
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
}',   true, now(), 1, now(), 1);