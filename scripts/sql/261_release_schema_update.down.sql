UPDATE devtron_resource_schema set schema = '{
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
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release we keep data as devtron resource only. To differ between nature of these two types of id values.",
                    "enum":
                    [
                        "resourceObjectId",
                        "oldObjectId"
                    ]
                },
                "firstReleasedOn":
                {
                    "type": "string",
                    "format": "date-time"
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
                    "type": "string"
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
                        "Type of release":
                        {
                            "type": "string",
                            "enum":
                            [
                                "Major",
                                "Minor",
                                "Patch"
                            ]
                        },
                        "Release Managers":
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
                        "On-Duty":
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
                        "Milestones":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Release planned start date":
                                {
                                    "type": "string",
                                    "format": "date"
                                },
                                "30% milestone date":
                                {
                                    "type": "string",
                                    "format": "date"
                                },
                                "70% milestone date":
                                {
                                    "type": "string",
                                    "format": "date"
                                },
                                "Release end date":
                                {
                                    "type": "string",
                                    "format": "date"
                                }
                            }
                        },
                        "Target customers":
                        {
                            "type": "array",
                            "uniqueItems": true,
                            "items":
                            {
                                "type": "string"
                            }
                        },
                        "Released customers":
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
                        "Type of release",
                        "Release Managers",
                        "On-Duty"
                    ]
                }
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
                    "properties":
                    {
                        "status":
                        {
                            "type": "string",
                            "enum":
                            [
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
}' where devtron_resource_id=(select id from devtron_resource where kind = 'release');