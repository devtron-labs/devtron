BEGIN;

DELETE FROM global_policy where policy_of = 'RELEASE_ACTION_CHECK' AND version = 'V2';
-- reverting to old state
DROP INDEX IF EXISTS idx_unique_policy_name_policy_of_version;
-- reverting to old state
CREATE UNIQUE INDEX idx_unique_policy_name_policy_of
    ON global_policy (name,policy_of)
    WHERE deleted = false;

UPDATE devtron_resource_schema set schema = '{
    "type": "object",
    "title": "Release Schema",
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "required":
    [
        "version",
        "kind",
        "overview",
        "status"
    ],
    "properties":
    {
        "kind":
        {
            "const": "release"
        },
        "status":
        {
            "type": "object",
            "required":
            [
                "config"
            ],
            "properties":
            {
                "config":
                {
                    "type": "object",
                    "required":
                    [
                        "status"
                    ],
                    "properties":
                    {
                        "lock":
                        {
                            "type": "boolean"
                        },
                        "status":
                        {
                            "enum":
                            [
                                "draft",
                                "readyForRelease",
                                "hold"
                            ],
                            "type": "string"
                        }
                    }
                }
            }
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
            "required":
            [
                "id",
                "releaseVersion"
            ],
            "properties":
            {
                "id":
                {
                    "type": "number"
                },
                "icon":
                {
                    "type": "string",
                    "contentEncoding": "base64"
                },
                "name":
                {
                    "type": "string"
                },
                "note":
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
                    "type": "string",
                    "description": "for existing resources in the system we keep original ids of their tables in id field. Like id of apps table is kept for devtron applications. But in release we keep data as devtron resource only. To differ between nature of these two types of id values."
                },
                "metadata":
                {
                    "type": "object",
                    "required":
                    [
                        "Type of release",
                        "Release Managers",
                        "On-Duty"
                    ],
                    "properties":
                    {
                        "On-Duty":
                        {
                            "type": "array",
                            "items":
                            {
                                "type": "object",
                                "refType": "#/references/users"
                            },
                            "minItems": 1,
                            "uniqueItems": true
                        },
                        "Milestones":
                        {
                            "type": "object",
                            "properties":
                            {
                                "Release end date":
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
                                "Release planned start date":
                                {
                                    "type": "string",
                                    "format": "date"
                                }
                            }
                        },
                        "Type of release":
                        {
                            "enum":
                            [
                                "Major",
                                "Minor",
                                "Patch"
                            ],
                            "type": "string"
                        },
                        "Release Managers":
                        {
                            "type": "array",
                            "items":
                            {
                                "type": "object",
                                "refType": "#/references/users"
                            },
                            "minItems": 1,
                            "uniqueItems": true
                        },
                        "Target customers":
                        {
                            "type": "array",
                            "items":
                            {
                                "type": "string"
                            },
                            "uniqueItems": true
                        },
                        "Released customers":
                        {
                            "type": "array",
                            "items":
                            {
                                "type": "string"
                            },
                            "uniqueItems": true
                        }
                    }
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
                "releaseVersion":
                {
                    "type": "string"
                },
                "firstReleasedOn":
                {
                    "type": "string",
                    "format": "date-time"
                }
            }
        },
        "taskMapping":
        {
            "type": "array"
        },
        "dependencies":
        {
            "type": "array"
        }
    }
}' where devtron_resource_id=(select id from devtron_resource where kind = 'release');

COMMIT;