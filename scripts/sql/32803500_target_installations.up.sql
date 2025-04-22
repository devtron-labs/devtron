BEGIN;

-- Drop the existing index if it exists because it did not have the version to it, added new index with version
DROP INDEX IF EXISTS idx_unique_policy_name_policy_of;
-- Create unique index for global_policy table
CREATE UNIQUE INDEX idx_unique_policy_name_policy_of
    ON global_policy (name,policy_of,version)
    WHERE deleted = false;

INSERT INTO global_policy(name, policy_of, version, description, policy_json, enabled, deleted, created_by, created_on, updated_by, updated_on)
VALUES('ReleaseActionCheckPolicy', 'RELEASE_ACTION_CHECK', 'V2', 'Policy used for validating different actions requested on release.',
       '{
  "definitions":
  [
    {
      "operationType": "patch",
      "operationPaths":
      [
        "overview.description",
        "overview.releaseNote",
        "overview.tags",
        "overview.name",
        "overview.metadata",
        "status.config.lock",
        "status.config"
      ],
      "possibleFromStates":
      [
        {
          "configStatus": "draft",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "readyForRelease",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "hold",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        }
      ]
    },
    {
      "operationType": "patch",
      "operationPaths":
      [
        "target"
      ],
      "stateTo":
      {
        "configStatus": "*",
        "dependencyArtifactStatus": "*",
        "rolloutStatus": "*",
        "lockStatus": "*",
        "isReleaseTrackAppPropagationOffending": "*",
        "areAllReleasedTargetPresent": "yes"
      },
      "possibleFromStates":
      [
        {
          "configStatus": "draft",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "readyForRelease",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "hold",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        }
      ]
    },
    {
      "operationType": "patch",
      "operationPaths":
      [
        "dependency.applications"
      ],
      "stateTo":
      {
        "configStatus": "draft",
        "dependencyArtifactStatus": "*",
        "rolloutStatus": "notDeployed",
        "lockStatus": "unLocked",
        "isReleaseTrackAppPropagationOffending": "yes",
        "areAllReleasedTargetPresent": "*"
      },
      "possibleFromStates":
      [
        {
          "configStatus": "draft",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "unLocked",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        }
      ]
    },
    {
      "operationType": "patch",
      "operationPaths":
      [
        "dependency.applications.image"
      ],
      "possibleFromStates":
      [
        {
          "configStatus": "draft",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "unLocked",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        }
      ]
    },
    {
      "operationType": "patch",
      "operationPaths":
      [
        "dependency.applications.instruction"
      ],
      "possibleFromStates":
      [
        {
          "configStatus": "draft",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "unLocked",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "readyForRelease",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "unLocked",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "hold",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "unLocked",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        }
      ]
    },
    {
      "operationType": "deploymentTrigger",
      "possibleFromStates":
      [
        {
          "configStatus": "readyForRelease",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "locked",
          "isReleaseTrackAppPropagationOffending": "yes",
          "areAllReleasedTargetPresent": "yes"

        },
        {
          "configStatus": "readyForRelease",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "partiallyDeployed",
          "lockStatus": "locked",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "yes"
        },
        {
          "configStatus": "readyForRelease",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "completelyDeployed",
          "lockStatus": "locked",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "yes"
        }
      ]
    },
    {
      "operationType": "delete",
      "possibleFromStates":
      [
        {
          "configStatus": "draft",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "readyForRelease",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "hold",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "notDeployed",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "rescind",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        },
        {
          "configStatus": "corrupted",
          "dependencyArtifactStatus": "*",
          "rolloutStatus": "*",
          "lockStatus": "*",
          "isReleaseTrackAppPropagationOffending": "*",
          "areAllReleasedTargetPresent": "*"
        }
      ]
    }
  ],
  "consequence": "BLOCK"
}', true, false, 1, now(),1,now());

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
        },
        "target":
        {
            "type": "object"
        }
    }
}' where devtron_resource_id=(select id from devtron_resource where kind = 'release');

COMMIT;