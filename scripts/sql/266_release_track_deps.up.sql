ALTER TABLE devtron_resource_object_dep_relations ADD COLUMN IF NOT EXISTS dependency_object_identifier VARCHAR(150);
ALTER TABLE devtron_resource_object_dep_relations ADD COLUMN IF NOT EXISTS component_object_identifier VARCHAR(150);

UPDATE global_policy set policy_json='{
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
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "enforcementStatus": "*"
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
                "enforcementStatus": "allowed"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked",
                    "enforcementStatus": "*"
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
                    "enforcementStatus": "*"
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
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked",
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked",
                    "enforcementStatus": "*"
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
                    "enforcementStatus": "allowed"
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
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "*",
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "*",
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "rescind",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "enforcementStatus": "*"
                },
                {
                    "configStatus": "corrupted",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "enforcementStatus": "*"
                }
            ]
        }
    ],
    "consequence": "BLOCK"
}' where policy_of='RELEASE_ACTION_CHECK';
