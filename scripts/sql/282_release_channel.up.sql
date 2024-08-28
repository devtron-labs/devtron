INSERT INTO devtron_resource(kind, display_name, icon,is_exposed, parent_kind_id, deleted, created_on, created_by, updated_on,
                             updated_by)
VALUES ('release-channel', 'Release Channel', '',true, 0, false, now(), 1, now(), 1);

INSERT INTO devtron_resource_schema(devtron_resource_id, version, schema, sample_schema, latest, created_on, created_by, updated_on,
                                    updated_by)
VALUES ((select id from devtron_resource where kind = 'release-channel'), 'alpha1',
        '{
    "type": "object",
    "title": "Release Channel Schema",
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "required":
    [
        "version",
        "kind",
        "overview"
    ],
    "properties":
    {
        "kind":
        {
            "const": "release-channel"
        },
        "version":
        {
            "enum":
            [
                "alpha1"
            ],
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
                "icon":
                {
                    "type": "string",
                    "format": "uri"
                },
                "name":
                {
                    "type": "string"
                },
                "tags":
                {
                    "additionalProperties":
                    {
                        "type": "string"
                    }
                },
                "idType":
                {
                    "enum":
                    [
                        "resourceObjectId",
                        "oldObjectId"
                    ],
                    "type": "string"
                },
                "default":
                {
                    "type": "boolean"
                },
                "metadata":
                {
                    "type": "object",
                    "properties":
                    {}
                },
                "createdBy":
                {
                    "type": "object",
                    "refType": "#/references/users"
                },
                "createdOn":
                {
                    "type": "string"
                },
                "description":
                {
                    "type": "string"
                },
                "releaseChannelId":
                {
                    "type": "string"
                }
            },
            "required":
            [
                "id",
                "idType",
                "releaseChannelId"
            ]
        },
        "dependencies":
        {
            "type": "array"
        }
    }
}','{
    "type": "object",
    "title": "Release Channel Schema",
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "required":
    [
        "version",
        "kind",
        "overview"
    ],
    "properties":
    {
        "kind":
        {
            "const": "release-channel"
        },
        "version":
        {
            "enum":
            [
                "alpha1"
            ],
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
                "icon":
                {
                    "type": "string",
                    "format": "uri"
                },
                "name":
                {
                    "type": "string"
                },
                "tags":
                {
                    "additionalProperties":
                    {
                        "type": "string"
                    }
                },
                "idType":
                {
                    "enum":
                    [
                        "resourceObjectId",
                        "oldObjectId"
                    ],
                    "type": "string"
                },
                "default":
                {
                    "type": "boolean"
                },
                "metadata":
                {
                    "type": "object",
                    "properties":
                    {}
                },
                "createdBy":
                {
                    "type": "object",
                    "refType": "#/references/users"
                },
                "createdOn":
                {
                    "type": "string"
                },
                "description":
                {
                    "type": "string"
                },
                "releaseChannelId":
                {
                    "type": "string"
                }
            },
            "required":
            [
                "id",
                "idType",
                "releaseChannelId"
            ]
        },
        "dependencies":
        {
            "type": "array"
        }
    }
}',true, now(), 1, now(), 1);


